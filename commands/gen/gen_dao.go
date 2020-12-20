package gen

import (
	"bytes"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"strings"
)

const (
	genDaoDefaultPath          = "./app"
	nodeNameGenDaoInConfigFile = "gfcli.gen.dao"
)

// doGenDao implements the "gen dao" command.
func doGenDao(parser *gcmd.Parser) {
	config := g.Cfg()
	if config.Available() {
		v := config.GetVar(nodeNameGenDaoInConfigFile)
		if v.IsSlice() {
			for i := 0; i < len(v.Interfaces()); i++ {
				doGenDaoForArray(i, parser)
			}
		} else {
			doGenDaoForArray(-1, parser)
		}
	} else {
		doGenDaoForArray(-1, parser)
	}
	mlog.Print("done!")
}

// doGenDaoForArray implements the "gen dao" command for configuration array.
func doGenDaoForArray(index int, parser *gcmd.Parser) {
	var (
		err          error
		db           gdb.DB
		modName      = getOptionOrConfigForDao(index, parser, "mod")                     // Go module name, eg: github.com/gogf/gf.
		dirPath      = getOptionOrConfigForDao(index, parser, "path", genDaoDefaultPath) // Generated directory path.
		tablesStr    = getOptionOrConfigForDao(index, parser, "tables")                  // Tables that will be generated.
		prefixName   = getOptionOrConfigForDao(index, parser, "prefix")                  // Add prefix to DAO and Model struct name.
		removePrefix = getOptionOrConfigForDao(index, parser, "remove-prefix")           // Remove prefix from table name.
		linkInfo     = getOptionOrConfigForDao(index, parser, "link")                    // Custom database link.
		configPath   = getOptionOrConfigForDao(index, parser, "config")                  // Config file path, eg: ./config/db.toml.
		configGroup  = getOptionOrConfigForDao(index, parser, "group", "default")        // Group name of database configuration node for generated DAO.
	)
	removePrefixArray := gstr.SplitAndTrim(removePrefix, ",")
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
	// It reads database configuration from project configuration file.
	if configPath != "" {
		path, err := gfile.Search(configPath)
		if err != nil {
			mlog.Fatalf("search configuration file '%s' failed: %v", configPath, err)
		}
		if err := g.Cfg().SetPath(gfile.Dir(path)); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", path, err)
		}
		g.Cfg().SetFileName(gfile.Basename(path))
	}
	// It uses user passed database configuration.
	if linkInfo != "" {
		tempGroup := gtime.TimestampNanoStr()
		match, _ := gregex.MatchString(`([a-z]+):(.+)`, linkInfo)
		if len(match) == 3 {
			gdb.AddConfigNode(tempGroup, gdb.ConfigNode{
				Type:     gstr.Trim(match[1]),
				LinkInfo: gstr.Trim(match[2]),
			})
			db, _ = gdb.Instance(tempGroup)
		}
	} else {
		db = g.DB(configGroup)
	}
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	tableNames := ([]string)(nil)
	if tablesStr != "" {
		tableNames = gstr.SplitAndTrim(tablesStr, ",")
	} else {
		tableNames, err = db.Tables()
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
	}

	for _, tableName := range tableNames {
		newTableName := tableName
		for _, v := range removePrefixArray {
			newTableName = gstr.TrimLeftStr(newTableName, v, 1)
		}
		generateDaoAndModelContentFile(
			db,
			tableName,
			newTableName,
			prefixName,
			configGroup,
			modName,
			dirPath,
		)
	}
}

// generateDaoAndModelContentFile generates the dao and model content of given table.
//
// The parameter <tableName> specifies the table name of the table.
// The parameter <newTableName> specifies the prefix-stripped name of the table.
// The parameter <prefixName> specifies the custom prefix name for generated dao and model struct.
// The parameter <groupName> specifies the group name of database configuration node for generated DAO.
// The parameter <modName> specifies the module name of current golang project, which is used for import purpose.
// The parameter <dirPath> specifies the directory path for generated files.
func generateDaoAndModelContentFile(db gdb.DB, tableName, newTableName, prefixName, groupName, modName, dirPath string) {
	fieldMap, err := db.TableFields(tableName)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", tableName, err)
	}
	// Change the `newTableName` if `prefixName` is given.
	newTableName = prefixName + newTableName
	var (
		dirPathDao              = gstr.Trim(gfile.Join(dirPath, "dao"), "./")
		dirPathModel            = gstr.Trim(gfile.Join(dirPath, "model"), "./")
		tableNameCamelCase      = gstr.CamelCase(newTableName)
		tableNameCamelLowerCase = gstr.CamelLowerCase(newTableName)
		tableNameSnakeCase      = gstr.SnakeCase(newTableName)
		structDefine            = generateStructDefinitionForDao(tableNameCamelCase, fieldMap)
		packageImports          = ""
		importPrefix            = ""
		dirRealPath             = gfile.RealPath(dirPath)
	)
	if dirRealPath == "" {
		dirRealPath = dirPath
		importPrefix = dirRealPath
		importPrefix = gstr.Trim(dirRealPath, "./")
	} else {
		importPrefix = gstr.Replace(dirRealPath, gfile.Pwd(), "")
	}
	importPrefix = gstr.Replace(importPrefix, gfile.Separator, "/")
	importPrefix = gstr.Join(g.SliceStr{modName, importPrefix}, "/")
	importPrefix, _ = gregex.ReplaceString(`\/{2,}`, `/`, gstr.Trim(importPrefix, "/"))
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
	path := gfile.Join(dirPathModel, fileName+".go")
	if !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(templateDaoModelIndexContent, g.MapStrStr{
			"{TplImportPrefix}":       importPrefix,
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
	path = gfile.Join(dirPathModel, "internal", fileName+".go")
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
	path = gfile.Join(dirPathDao, fileName+".go")
	if !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(templateDaoDaoIndexContent, g.MapStrStr{
			"{TplImportPrefix}":            importPrefix,
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
	path = gfile.Join(dirPathDao, "internal", fileName+".go")
	modelContent := gstr.ReplaceByMap(templateDaoDaoInternalContent, g.MapStrStr{
		"{TplImportPrefix}":            importPrefix,
		"{TplTableName}":               tableName,
		"{TplGroupName}":               groupName,
		"{TplTableNameCamelCase}":      tableNameCamelCase,
		"{TplTableNameCamelLowerCase}": tableNameCamelLowerCase,
		"{TplStructDefine}":            structDefine,
		"{TplColumnDefine}":            gstr.Trim(generateColumnDefinitionForDao(fieldMap)),
		"{TplColumnNames}":             gstr.Trim(generateColumnNamesForDao(fieldMap)),
	})
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		mlog.Print("generated:", path)
	}
}

// generateStructDefinitionForDao generates and returns the struct definition for specified table.
func generateStructDefinitionForDao(structName string, fieldMap map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKeyForDao(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructFieldForDao(field)
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

// generateStructFieldForDao generates and returns the attribute definition for specified field.
func generateStructFieldForDao(field *gdb.TableField) []string {
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

// generateColumnDefinitionForDao generates and returns the column names definition for specified table.
func generateColumnDefinitionForDao(fieldMap map[string]*gdb.TableField) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyForDao(fieldMap)
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

// generateColumnNamesForDao generates and returns the column names assignment content of column struct
// for specified table.
func generateColumnNamesForDao(fieldMap map[string]*gdb.TableField) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyForDao(fieldMap)
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

func sortFieldKeyForDao(fieldMap map[string]*gdb.TableField) []string {
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

// getOptionOrConfigForDao retrieves option value from parser and configuration file.
// It returns the default value specified by parameter <value> is no value found.
func getOptionOrConfigForDao(index int, parser *gcmd.Parser, name string, defaultValue ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" && g.Config().Available() {
		g.Cfg().SetViolenceCheck(true)
		if index >= 0 {
			result = g.Cfg().GetString(fmt.Sprintf(`%s.%d.%s`, nodeNameGenDaoInConfigFile, index, name))
		} else {
			result = g.Cfg().GetString(fmt.Sprintf(`%s.%s`, nodeNameGenDaoInConfigFile, name))
		}
	}
	if result == "" && len(defaultValue) > 0 {
		result = defaultValue[0]
	}
	return
}
