package gen

import (
	"bytes"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gogf/gf-cli/library/allyes"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"strings"
)

const (
	genDaoDaoPath              = "./app/dao"
	genDaoModelPath            = "./app/model"
	nodeNameGenDaoInConfigFile = "gfcli.gen.dao"
)

// doGenDao implements the "gen dao" command.
func doGenDao(parser *gcmd.Parser) {
	var err error
	if !allyes.Check() {
		notEmptyPaths := garray.NewStrArray()
		if !gfile.IsEmpty(genDaoModelPath) {
			notEmptyPaths.Append(genDaoModelPath)
		}
		if !gfile.IsEmpty(genDaoDaoPath) {
			notEmptyPaths.Append(genDaoDaoPath)
		}
		if !notEmptyPaths.IsEmpty() {
			paths := `'` + notEmptyPaths.Join(`' and '`) + `'`
			s := gcmd.Scanf("paths %s is not empty, files might be overwrote, continue? [y/n]: ", paths)
			if strings.EqualFold(s, "n") {
				return
			}
		}
	}

	var (
		modName     = getOptionForDao(parser, "mod")
		tableOpt    = getOptionForDao(parser, "table")
		linkInfo    = getOptionForDao(parser, "link")
		configFile  = getOptionForDao(parser, "config")
		configGroup = getOptionForDao(parser, "group", gdb.DEFAULT_GROUP_NAME)
		prefixArray = gstr.SplitAndTrim(parser.GetOpt("prefix"), ",")
	)
	if modName == "" {
		if !gfile.Exists("go.mod") {
			mlog.Fatal("go.mod does not exist in current working directory")
		}
		var (
			goModContent = gfile.GetContents("go.mod")
			match, _     = gregex.MatchString(`module\s+(.+)\s+`, goModContent)
		)
		if len(match) > 1 {
			modName = match[1]
		} else {
			mlog.Fatal("module name does not found in go.mod")
		}
	}
	// It uses user passed database configuration.
	if linkInfo != "" {
		path := gfile.TempDir() + gfile.Separator + "config.toml"
		if err := gfile.PutContents(path, fmt.Sprintf("[database]\n\tlink=\"%s\"", linkInfo)); err != nil {
			mlog.Fatalf("write configuration file to '%s' failed: %v", path, err)
		}
		defer gfile.Remove(path)
		if err := g.Cfg().SetPath(gfile.TempDir()); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", gfile.TempDir(), err)
		}
	}
	// It reads database configuration from project confifuration file.
	if configFile != "" {
		path, err := gfile.Search(configFile)
		if err != nil {
			mlog.Fatalf("search configuration file '%s' failed: %v", configFile, err)
		}
		if err := g.Cfg().SetPath(gfile.Dir(path)); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", path, err)
		}
		g.Cfg().SetFileName(gfile.Basename(path))
	}

	db := g.DB(configGroup)
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	tables := ([]string)(nil)
	if tableOpt != "" {
		tables = gstr.SplitAndTrim(tableOpt, ",")
	} else {
		tables, err = db.Tables()
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
	}

	for _, table := range tables {
		variable := table
		for _, v := range prefixArray {
			variable = gstr.TrimLeftStr(variable, v)
		}
		generateDaoAndModelContentFile(db, table, variable, configGroup, modName)
	}
	mlog.Print("done!")
}

// generateDaoAndModelContentFile generates the dao and model content of given table.
// The parameter <variable> specifies the variable name for the table, which
// is the prefix-stripped name of the table.
func generateDaoAndModelContentFile(db gdb.DB, tableName, variable, groupName, modName string) {
	fieldMap, err := db.TableFields(tableName)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", tableName, err)
	}
	var (
		tableNameCamelCase      = gstr.CamelCase(variable)
		tableNameCamelLowerCase = gstr.CamelLowerCase(variable)
		tableNameSnakeCase      = gstr.SnakeCase(variable)
		structDefine            = generateStructDefinitionV2(tableNameCamelCase, fieldMap)
		packageImports          = ""
	)
	if strings.Contains(structDefine, "gtime.Time") {
		packageImports = gstr.Trim(`
import (
	"github.com/gogf/gf/os/gtime"
)`)
	} else {
		packageImports = ""
	}
	fileName := gstr.Trim(tableNameSnakeCase, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}
	// model - index
	path := gfile.Join(genDaoModelPath, fileName+".go")
	if !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(templateDaoModelIndexContent, g.MapStrStr{
			"{TplModName}":            modName,
			"{TplTableName}":          tableName,
			"{TplTableNameCamelCase}": tableNameCamelCase,
		})
		if err := gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			mlog.Print("generated:", path)
		}
	}
	// model - internal
	path = gfile.Join(genDaoModelPath, "internal", fileName+".go")
	entityContent := gstr.ReplaceByMap(templateDaoModelInternalContent, g.MapStrStr{
		"{TplTableName}":          tableName,
		"{TplTableNameCamelCase}": tableNameCamelCase,
		"{TplPackageImports}":     packageImports,
		"{TplStructDefine}":       structDefine,
	})
	if err := gfile.PutContents(path, strings.TrimSpace(entityContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		mlog.Print("generated:", path)
	}
	// dao - index
	path = gfile.Join(genDaoDaoPath, fileName+".go")
	if !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(templateDaoDaoIndexContent, g.MapStrStr{
			"{TplModName}":                 modName,
			"{TplTableName}":               tableName,
			"{TplTableNameCamelCase}":      tableNameCamelCase,
			"{TplTableNameCamelLowerCase}": tableNameCamelLowerCase,
		})
		if err := gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			mlog.Print("generated:", path)
		}
	}
	// dao - internal
	path = gfile.Join(genDaoDaoPath, "internal", fileName+".go")
	modelContent := gstr.ReplaceByMap(templateDaoDaoInternalContent, g.MapStrStr{
		"{TplModName}":                 modName,
		"{TplTableName}":               tableName,
		"{TplGroupName}":               groupName,
		"{TplTableNameCamelCase}":      tableNameCamelCase,
		"{TplTableNameCamelLowerCase}": tableNameCamelLowerCase,
		"{TplStructDefine}":            structDefine,
		"{TplColumnDefine}":            gstr.Trim(generateColumnDefinitionV2(fieldMap)),
		"{TplColumnNames}":             gstr.Trim(generateColumnNamesV2(fieldMap)),
	})
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		mlog.Print("generated:", path)
	}
}

// generateStructDefinitionV2 generates and returns the struct definition for specified table.
func generateStructDefinitionV2(structName string, fieldMap map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKeyV2(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructFieldV2(field)
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	stContent := buffer.String()
	// Let's do this hack of table writer for indent!
	stContent = gstr.Replace(stContent, "  #", "")
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	buffer.WriteString(stContent)
	buffer.WriteString("}")
	return buffer.String()
}

// generateStructFieldV2 generates and returns the attribute definition for specified field.
func generateStructFieldV2(field *gdb.TableField) []string {
	var typeName, ormTag, jsonTag, comment string
	t, _ := gregex.ReplaceString(`\(.+\)`, "", field.Type)
	t = gstr.Split(gstr.Trim(t), " ")[0]
	t = gstr.ToLower(t)
	switch t {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		typeName = "[]byte"

	case "bit", "int", "tinyint", "small_int", "smallint", "medium_int", "mediumint":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint"
		} else {
			typeName = "int"
		}

	case "big_int", "bigint":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint64"
		} else {
			typeName = "int64"
		}

	case "float", "double", "decimal":
		typeName = "float64"

	case "bool":
		typeName = "bool"

	case "datetime", "timestamp", "date", "time":
		typeName = "*gtime.Time"

	default:
		// Auto detecting type.
		switch {
		case strings.Contains(t, "int"):
			typeName = "int"
		case strings.Contains(t, "text") || strings.Contains(t, "char"):
			typeName = "string"
		case strings.Contains(t, "float") || strings.Contains(t, "double"):
			typeName = "float64"
		case strings.Contains(t, "bool"):
			typeName = "bool"
		case strings.Contains(t, "binary") || strings.Contains(t, "blob"):
			typeName = "[]byte"
		case strings.Contains(t, "date") || strings.Contains(t, "time"):
			typeName = "*gtime.Time"
		default:
			typeName = "string"
		}
	}
	ormTag = field.Name
	jsonTag = gstr.SnakeCase(field.Name)
	if gstr.ContainsI(field.Key, "pri") {
		ormTag += ",primary"
	}
	if gstr.ContainsI(field.Key, "uni") {
		ormTag += ",unique"
	}
	comment = gstr.ReplaceByArray(field.Comment, g.SliceStr{
		"\n", " ",
		"\r", " ",
	})
	comment = gstr.Trim(comment)
	comment = gstr.Replace(comment, `\n`, " ")
	return []string{
		"    #" + gstr.CamelCase(field.Name),
		" #" + typeName,
		" #" + fmt.Sprintf("`"+`orm:"%s"`, ormTag),
		" #" + fmt.Sprintf(`json:"%s"`+"`", jsonTag),
		" #" + fmt.Sprintf(`// %s`, comment),
	}
}

// generateColumnDefinitionV2 generates and returns the column names definition for specified table.
func generateColumnDefinitionV2(fieldMap map[string]*gdb.TableField) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyV2(fieldMap)
	)
	for index, name := range names {
		field := fieldMap[name]
		comment := gstr.Trim(gstr.ReplaceByArray(field.Comment, g.SliceStr{
			"\n", " ",
			"\r", " ",
		}))
		array[index] = []string{
			"    #" + gstr.CamelCase(field.Name),
			" # " + "string",
			" #" + fmt.Sprintf(`// %s`, comment),
		}
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	defineContent := buffer.String()
	// Let's do this hack of table writer for indent!
	defineContent = gstr.Replace(defineContent, "  #", "")
	buffer.Reset()
	buffer.WriteString(defineContent)
	return buffer.String()
}

// generateColumnNamesV2 generates and returns the column names assignment content of column struct
// for specified table.
func generateColumnNamesV2(fieldMap map[string]*gdb.TableField) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyV2(fieldMap)
	)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = []string{
			"            #" + gstr.CamelCase(field.Name) + ":",
			fmt.Sprintf(` #"%s",`, field.Name),
		}
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	namesContent := buffer.String()
	// Let's do this hack of table writer for indent!
	namesContent = gstr.Replace(namesContent, "  #", "")
	buffer.Reset()
	buffer.WriteString(namesContent)
	return buffer.String()
}

func sortFieldKeyV2(fieldMap map[string]*gdb.TableField) []string {
	names := make(map[int]string)
	for _, field := range fieldMap {
		names[field.Index] = field.Name
	}
	result := make([]string, len(names))
	i := 0
	j := 0
	for {
		if len(names) == 0 {
			break
		}
		if val, ok := names[i]; ok {
			result[j] = val
			j++
			delete(names, i)
		}
		i++
	}
	return result
}

// getOptionForDao retrieves option value from parser and configuration file.
// It returns the default value specified by parameter <value> is no value found.
func getOptionForDao(parser *gcmd.Parser, name string, value ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" && g.Config().Available() {
		g.Config().SetViolenceCheck(true)
		result = g.Config().GetString(nodeNameGenDaoInConfigFile + "." + name)
	}
	if result == "" && len(value) > 0 {
		result = value[0]
	}
	return
}
