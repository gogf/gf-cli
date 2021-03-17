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
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"strings"
)

const (
	genModelPath                 = "./app/model"
	nodeNameGenModelInConfigFile = "gfcli.gen.model"
)

func HelpModel() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen model [OPTION]

OPTION
    -/--path             directory path for generated files.
    -l, --link           database configuration, the same as the ORM configuration of GoFrame.
    -t, --tables         generate models only for given tables, multiple table names separated with ',' 
    -g, --group          specifying the configuration group name of database for generated ORM instance,
                         it's not necessary and the default value is "default"
    -c, --config         used to specify the configuration file for database, it's commonly not necessary.
                         If "-l" is not passed, it will search "./config.toml" and "./config/config.toml" 
                         in current working directory in default.
    -p, --prefix         add prefix for all table of specified link/database tables.
    -r, --removePrefix   remove specified prefix of the table, multiple prefix separated with ',' 
    -m, --mod            module name for generated golang file imports.
                  
CONFIGURATION SUPPORT
    Options are also supported by configuration file.
    It's suggested using configuration file instead of command line arguments making producing. 
    The configuration node name is "gf.gen.model", which also supports multiple databases, for example:
    [gfcli]
        [[gfcli.gen.model]]
            link   = "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
            tables = "order,products"
        [[gfcli.gen.model]]
            link   = "mysql:root:12345678@tcp(127.0.0.1:3306)/primary"
            path   = "./my-app"
            prefix = "primary_"
            tables = "user, userDetail"

EXAMPLES
    gf gen model
    gf gen model -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
    gf gen model -path ./model -c config.yaml -g user-center -t user,user_detail,user_login
    gf gen model -r user_
`))
}

// doGenModel implements the "gen model" command.
func doGenModel() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"path":           true,
		"m,mod":          true,
		"l,link":         true,
		"t,tables":       true,
		"g,group":        true,
		"c,config":       true,
		"p,prefix":       true,
		"r,removePrefix": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	config := g.Cfg()
	if config.Available() {
		v := config.GetVar(nodeNameGenModelInConfigFile)
		if v.IsEmpty() && g.IsEmpty(parser.GetOptAll()) {
			mlog.Fatal(`command arguments and configurations not found for generating model files`)
		}
		if v.IsSlice() {
			for i := 0; i < len(v.Interfaces()); i++ {
				doGenModelForArray(i, parser)
			}
		} else {
			doGenModelForArray(-1, parser)
		}
	} else {
		doGenModelForArray(-1, parser)
	}
	mlog.Print("done!")
}

func doGenModelForArray(index int, parser *gcmd.Parser) {
	var (
		err          error
		genPath      = getOptionOrConfigForModel(index, parser, "path", genModelPath)
		tableOpt     = getOptionOrConfigForModel(index, parser, "tables")
		linkInfo     = getOptionOrConfigForModel(index, parser, "link")
		configFile   = getOptionOrConfigForModel(index, parser, "config")
		configGroup  = getOptionOrConfigForModel(index, parser, "group", gdb.DefaultGroupName)
		removePrefix = getOptionOrConfigForModel(index, parser, "removePrefix")
	)
	// Make it compatible with old CLI version.
	if removePrefix == "" {
		removePrefix = getOptionOrConfigForModel(index, parser, "remove-prefix")
	}
	removePrefixArray := gstr.SplitAndTrim(removePrefix, ",")
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
	// Custom configuration file.
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

	if err := gfile.Mkdir(genPath); err != nil {
		mlog.Fatalf("mkdir for generating path '%s' failed: %v", genPath, err)
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
		for _, v := range removePrefixArray {
			variable = gstr.TrimLeftStr(variable, v)
		}
		generateModelContentFile(db, table, variable, genPath, configGroup)
	}
}

// generateModelContentFile generates the model content of given table.
// The parameter <variable> specifies the variable name for the table, which
// is the prefix-stripped name of the table.
//
// Note that, this function will generate 3 files under <folderPath>/<packageName>/:
// file.go        : the package index go file, developer can fill the file with model logic;
// file_entity.go : the entity definition go file, it can be overwrote by gf-cli tool, don't edit it;
// file_model.go  : the active record design model definition go file, it can be overwrote by gf-cli tool, don't edit it;
func generateModelContentFile(db gdb.DB, table, variable, folderPath, groupName string) {
	fieldMap, err := db.TableFields(table)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", table, err)
	}
	camelName := gstr.CaseCamel(variable)
	structDefine := generateStructDefinition(fieldMap)
	packageImports := ""
	if strings.Contains(structDefine, "gtime.Time") {
		packageImports = gstr.Trim(`
import (
	"database/sql"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/os/gtime"
)`)
	} else {
		packageImports = gstr.Trim(`
import (
	"database/sql"
	"github.com/gogf/gf/database/gdb"
)`)
	}
	packageName := gstr.CaseSnake(variable)
	fileName := gstr.Trim(packageName, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}
	// index
	path := gfile.Join(folderPath, packageName, fileName+".go")
	if !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(templateIndexContent, g.MapStrStr{
			"{TplTableName}":      table,
			"{TplModelName}":      camelName,
			"{TplGroupName}":      groupName,
			"{TplPackageName}":    packageName,
			"{TplPackageImports}": packageImports,
			"{TplStructDefine}":   structDefine,
		})
		if err := gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			mlog.Print("generated:", path)
		}
	}
	// entity
	path = gfile.Join(folderPath, packageName, fileName+"_entity.go")
	entityContent := gstr.ReplaceByMap(templateEntityContent, g.MapStrStr{
		"{TplTableName}":      table,
		"{TplModelName}":      camelName,
		"{TplGroupName}":      groupName,
		"{TplPackageName}":    packageName,
		"{TplPackageImports}": packageImports,
		"{TplStructDefine}":   structDefine,
	})
	if err := gfile.PutContents(path, strings.TrimSpace(entityContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		mlog.Print("generated:", path)
	}
	// model
	path = gfile.Join(folderPath, packageName, fileName+"_model.go")
	modelContent := gstr.ReplaceByMap(templateModelContent, g.MapStrStr{
		"{TplTableName}":      table,
		"{TplModelName}":      camelName,
		"{TplGroupName}":      groupName,
		"{TplPackageName}":    packageName,
		"{TplPackageImports}": packageImports,
		"{TplStructDefine}":   structDefine,
		"{TplColumnDefine}":   gstr.Trim(generateColumnDefinition(fieldMap)),
		"{TplColumnNames}":    gstr.Trim(generateColumnNames(fieldMap)),
	})
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		mlog.Print("generated:", path)
	}
}

// generateStructDefinition generates and returns the struct definition for specified table.
func generateStructDefinition(fieldMap map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKey(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructField(field)
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
	buffer.WriteString("type Entity struct {\n")
	buffer.WriteString(stContent)
	buffer.WriteString("}")
	return buffer.String()
}

// generateStructField generates and returns the attribute definition for specified field.
func generateStructField(field *gdb.TableField) []string {
	var typeName, ormTag, jsonTag, comment string
	t, _ := gregex.ReplaceString(`\(.+\)`, "", field.Type)
	t = gstr.Split(gstr.Trim(t), " ")[0]
	t = gstr.ToLower(t)
	switch t {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		typeName = "[]byte"

	case "bit", "int", "int2", "tinyint", "small_int", "smallint", "medium_int", "mediumint":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint"
		} else {
			typeName = "int"
		}

	case "big_int", "bigint", "int8", "int4":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint64"
		} else {
			typeName = "int64"
		}

	case "float", "double", "decimal", "numeric":
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
	jsonTag = gstr.CaseSnake(field.Name)
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
		"    #" + gstr.CaseCamel(field.Name),
		" #" + typeName,
		" #" + fmt.Sprintf("`"+`orm:"%s"`, ormTag),
		" #" + fmt.Sprintf(`json:"%s"`+"`", jsonTag),
		" #" + fmt.Sprintf(`// %s`, comment),
	}
}

// generateColumnDefinition generates and returns the column names definition for specified table.
func generateColumnDefinition(fieldMap map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKey(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		comment := gstr.Trim(gstr.ReplaceByArray(field.Comment, g.SliceStr{
			"\n", " ",
			"\r", " ",
		}))
		array[index] = []string{
			"        #" + gstr.CaseCamel(field.Name),
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

// generateColumnNames generates and returns the column names assignment content of column struct
// for specified table.
func generateColumnNames(fieldMap map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKey(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = []string{
			"        #" + gstr.CaseCamel(field.Name) + ":",
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

func sortFieldKey(fieldMap map[string]*gdb.TableField) []string {
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

// getOptionOrConfigForModel retrieves option value from parser and configuration file.
// It returns the default value specified by parameter <value> is no value found.
func getOptionOrConfigForModel(index int, parser *gcmd.Parser, name string, defaultValue ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" && g.Config().Available() {
		g.Cfg().SetViolenceCheck(true)
		if index >= 0 {
			result = g.Cfg().GetString(fmt.Sprintf(`%s.%d.%s`, nodeNameGenModelInConfigFile, index, name))
		} else {
			result = g.Cfg().GetString(fmt.Sprintf(`%s.%s`, nodeNameGenModelInConfigFile, name))
		}
	}
	if result == "" && len(defaultValue) > 0 {
		result = defaultValue[0]
	}
	return
}
