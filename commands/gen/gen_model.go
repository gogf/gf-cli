package gen

import (
	"bytes"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gogf/gf-cli/library/allyes"
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
	DEFAULT_GEN_MODEL_PATH = "./app/model"
)

// doGenModel implements the "gen model" command.
func doGenModel(parser *gcmd.Parser) {
	var err error
	genPath := parser.GetArg(3, DEFAULT_GEN_MODEL_PATH)
	if !gfile.IsEmpty(genPath) && !allyes.Check() {
		s := gcmd.Scanf("path '%s' is not empty, files might be overwrote, continue? [y/n]: ", genPath)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	tableOpt := parser.GetOpt("table")
	linkInfo := parser.GetOpt("link")
	configFile := parser.GetOpt("config")
	configGroup := parser.GetOpt("group", gdb.DEFAULT_GROUP_NAME)
	prefixArray := gstr.SplitAndTrim(parser.GetOpt("prefix"), ",")

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
		for _, v := range prefixArray {
			variable = gstr.TrimLeftStr(variable, v)
		}
		generateModelContentFile(db, table, variable, genPath, configGroup)
	}
	mlog.Print("done!")
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
	camelName := gstr.CamelCase(variable)
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
	packageName := gstr.SnakeCase(variable)
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
	for _, field := range fieldMap {
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
	return []string{
		"    #" + gstr.CamelCase(field.Name),
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
	for _, field := range fieldMap {
		comment := gstr.Trim(gstr.ReplaceByArray(field.Comment, g.SliceStr{
			"\n", " ",
			"\r", " ",
		}))
		array[field.Index] = []string{
			"        #" + gstr.CamelCase(field.Name),
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
	for _, field := range fieldMap {
		array[field.Index] = []string{
			"        #" + gstr.CamelCase(field.Name) + ":",
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
