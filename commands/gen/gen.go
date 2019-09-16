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
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"strings"
	//_ "github.com/mattn/go-oci8"
)

const (
	DEFAULT_GEN_MODEL_PATH      = "./app/model"
	DEFAULT_GEN_MODEL_INIT_NAME = "initialization"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen model [PATH] [OPTION]

ARGUMENT
    PATH  the destination for storing generated files, not necessary, default is "./app/model"

OPTION
    -n, --name    package name for generated go files, it's the configuration group name in default
    -l, --link    database configuration, please refer to: https://goframe.org/database/gdb/config
    -t, --table   generate models only for given tables, multiple table names separated with ',' 
    -g, --group   used with "-c" option, specifying the configuration group name for database,
                  it's not necessary and the default value is "default"
    -c, --config  used to specify the configuration file for database, it's commonly not necessary, 
                  if "-l" is not passed, it will search "./config.toml" and "./config/config.toml" 
                  in current working directory

EXAMPLES
    gf gen model
    gf gen model -n dao
    gf gen model -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
    gf gen model ./model -l "mssql:sqlserver://sa:12345678@127.0.0.1:1433?database=test"
    gf gen model ./model -c config.yaml -g user-center -t user,user_detail,user_login

DESCRIPTION
    The "gen" command is designed for multiple generating purposes.
    It's currently supporting generating go files for ORM models.
`))
}

func Run() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"n,name":   true,
		"l,link":   true,
		"t,table":  true,
		"g,group":  true,
		"c,config": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	genType := parser.GetArg(2)
	if genType == "" {
		mlog.Fatal("generating type cannot be empty")
	}
	genPath := parser.GetArg(3, DEFAULT_GEN_MODEL_PATH)
	if !gfile.IsEmpty(genPath) {
		s := gcmd.Scanf("path '%s' is not empty, files might be overwrote, continue? [y/n]: ", genPath)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	tableOpt := parser.GetOpt("table")
	linkInfo := parser.GetOpt("link")
	configFile := parser.GetOpt("config")
	configGroup := parser.GetOpt("group", gdb.DEFAULT_GROUP_NAME)
	packageName := parser.GetOpt("name", configGroup)

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

	if configFile != "" {
		path, err := gfile.Search(configFile)
		if err != nil {
			mlog.Fatalf("search configuration file '%s' failed: %v", configFile, err)
		}
		if err := g.Cfg().SetPath(path); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", path, err)
		}
		if err := g.Cfg().SetFileName(gfile.Basename(path)); err != nil {
			mlog.Fatalf("set configuration file name '%s' failed: %v", gfile.Basename(path), err)
		}
	}

	folderPath := genPath + gfile.Separator + packageName
	if err := gfile.Mkdir(folderPath); err != nil {
		mlog.Fatalf("mkdir for generating path '%s' failed: %v", folderPath, err)
	}

	db := g.DB(configGroup)
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	tables := ([]string)(nil)
	if tableOpt != "" {
		tables = gstr.SplitAndTrimSpace(tableOpt, ",")
	} else {
		tables, err = db.Tables()
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
		if strings.EqualFold(packageName, gdb.DEFAULT_GROUP_NAME) {
			packageName += "s"
			mlog.Printf(`package name '%s' is a reserved word of go, so it's renamed to '%s'`, gdb.DEFAULT_GROUP_NAME, packageName)
		}
	}

	for _, table := range tables {
		generateModelContentFile(db, table, folderPath, packageName, configGroup)
	}
	mlog.Print("done!")
}

func generateModelContentFile(db gdb.DB, table string, folderPath, packageName, groupName string) {
	fields, err := db.TableFields(table)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", table, err)
	}
	camelName := gstr.CamelCase(table)
	structDefine := generateStructDefine(table, fields)
	extraImports := ""
	if strings.Contains(structDefine, "gtime.Time") {
		extraImports = `import (
	"github.com/gogf/gf/os/gtime"
)
`
	}
	modelContent := gstr.ReplaceByMap(templateModelContent, g.MapStrStr{
		"{TplTableName}":    table,
		"{TplModelName}":    camelName,
		"{TplGroupName}":    groupName,
		"{TplPackageName}":  packageName,
		"{TplExtraImports}": extraImports,
		"{TplStructDefine}": structDefine,
	})
	name := gstr.Trim(gstr.SnakeCase(table), "-_.")
	if len(name) > 5 && name[len(name)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		name += "_table"
	}
	path := folderPath + gfile.Separator + name + ".go"
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing model content to '%s' failed: %v", path, err)
	}
}

func generateStructDefine(table string, fields map[string]*gdb.TableField) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fields))
	for _, field := range fields {
		array[field.Index] = generateStructField(field)
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	stContent := buffer.String()
	// Let's do this hack for tablewriter!
	stContent = gstr.Replace(stContent, "  #", "")
	buffer.Reset()
	buffer.WriteString("type " + gstr.CamelCase(table) + " struct {\n")
	buffer.WriteString(stContent)
	buffer.WriteString("}")
	return buffer.String()
}

func generateStructField(field *gdb.TableField) []string {
	var typeName, ormTag, jsonTag string
	t, _ := gregex.ReplaceString(`\(.+\)`, "", field.Type)
	t = strings.ToLower(t)
	switch t {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		typeName = "[]byte"

	case "bit", "int", "tinyint", "small_int", "medium_int":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint"
		} else {
			typeName = "int"
		}

	case "big_int":
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
	jsonTag = gstr.SnakeCase(field.Name)
	ormTag = jsonTag
	if gstr.ContainsI(field.Key, "pri") {
		ormTag += ",primary"
	}
	if gstr.ContainsI(field.Key, "uni") {
		ormTag += ",unique"
	}
	return []string{
		"    #" + gstr.CamelCase(field.Name),
		" #" + typeName,
		" #" + fmt.Sprintf("`"+`orm:"%s"`, ormTag),
		" #" + fmt.Sprintf(`json:"%s"`+"`", jsonTag),
	}
}
