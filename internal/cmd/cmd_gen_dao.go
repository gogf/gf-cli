package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf-cli/v2/internal/consts"
	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf-cli/v2/utility/utils"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gtag"
	"github.com/olekukonko/tablewriter"
)

const (
	commandGenDaoConfig = `gfcli.gen.dao`
	commandGenDaoUsage  = `gf gen dao [OPTION]`
	commandGenDaoBrief  = `automatically generate go files for dao/dto/entity`
	commandGenDaoEg     = `
gf gen dao
gf gen dao -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
gf gen dao -p ./model -c config.yaml -g user-center -t user,user_detail,user_login
gf gen dao -r user_
`

	commandGenDaoAd = `
CONFIGURATION SUPPORT
    Options are also supported by configuration file.
    It's suggested using configuration file instead of command line arguments making producing. 
    The configuration node name is "gf.gen.dao", which also supports multiple databases, for example(config.yaml):
    gfcli:
      gen:
	  - dao:
          link:     "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
          tables:   "order,products"
          jsonCase: "CamelLower"
      - dao:
          link:   "mysql:root:12345678@tcp(127.0.0.1:3306)/primary"
          path:   "./my-app"
          prefix: "primary_"
          tables: "user, userDetail"
`
	commandGenDaoBriefPath            = `directory path for generated files`
	commandGenDaoBriefLink            = `database configuration, the same as the ORM configuration of GoFrame`
	commandGenDaoBriefTables          = `generate models only for given tables, multiple table names separated with ','`
	commandGenDaoBriefTablesEx        = `generate models excluding given tables, multiple table names separated with ','`
	commandGenDaoBriefPrefix          = `add prefix for all table of specified link/database tables`
	commandGenDaoBriefRemovePrefix    = `remove specified prefix of the table, multiple prefix separated with ','`
	commandGenDaoBriefStdTime         = `use time.Time from stdlib instead of gtime.Time for generated time/date fields of tables`
	commandGenDaoBriefGJsonSupport    = `use gJsonSupport to use *gjson.Json instead of string for generated json fields of tables`
	commandGenDaoBriefImportPrefix    = `custom import prefix for generated go files`
	commandGenDaoBriefOverwriteDao    = `overwrite all dao files both inside/outside internal folder`
	commandGenDaoBriefModelFile       = `custom file name for storing generated model content`
	commandGenDaoBriefModelFileForDao = `custom file name generating model for DAO operations like Where/Data. It's empty in default`
	commandGenDaoBriefDescriptionTag  = `add comment to description tag for each field`
	commandGenDaoBriefNoJsonTag       = `no json tag will be added for each field`
	commandGenDaoBriefNoModelComment  = `no model comment will be added for each field`
	commandGenDaoBriefGroup           = `
specifying the configuration group name of database for generated ORM instance,
it's not necessary and the default value is "default"
`
	commandGenDaoBriefJsonCase = `
generated json tag case for model struct, cases are as follows:
| Case            | Example            |
|---------------- |--------------------|
| Camel           | AnyKindOfString    | 
| CamelLower      | anyKindOfString    | default
| Snake           | any_kind_of_string |
| SnakeScreaming  | ANY_KIND_OF_STRING |
| SnakeFirstUpper | rgb_code_md5       |
| Kebab           | any-kind-of-string |
| KebabScreaming  | ANY-KIND-OF-STRING |
`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandGenDaoConfig`:               commandGenDaoConfig,
		`commandGenDaoUsage`:                commandGenDaoUsage,
		`commandGenDaoBrief`:                commandGenDaoBrief,
		`commandGenDaoEg`:                   commandGenDaoEg,
		`commandGenDaoAd`:                   commandGenDaoAd,
		`commandGenDaoBriefPath`:            commandGenDaoBriefPath,
		`commandGenDaoBriefLink`:            commandGenDaoBriefLink,
		`commandGenDaoBriefTables`:          commandGenDaoBriefTables,
		`commandGenDaoBriefTablesEx`:        commandGenDaoBriefTablesEx,
		`commandGenDaoBriefPrefix`:          commandGenDaoBriefPrefix,
		`commandGenDaoBriefRemovePrefix`:    commandGenDaoBriefRemovePrefix,
		`commandGenDaoBriefStdTime`:         commandGenDaoBriefStdTime,
		`commandGenDaoBriefGJsonSupport`:    commandGenDaoBriefGJsonSupport,
		`commandGenDaoBriefImportPrefix`:    commandGenDaoBriefImportPrefix,
		`commandGenDaoBriefOverwriteDao`:    commandGenDaoBriefOverwriteDao,
		`commandGenDaoBriefModelFile`:       commandGenDaoBriefModelFile,
		`commandGenDaoBriefModelFileForDao`: commandGenDaoBriefModelFileForDao,
		`commandGenDaoBriefDescriptionTag`:  commandGenDaoBriefDescriptionTag,
		`commandGenDaoBriefNoJsonTag`:       commandGenDaoBriefNoJsonTag,
		`commandGenDaoBriefNoModelComment`:  commandGenDaoBriefNoModelComment,
		`commandGenDaoBriefGroup`:           commandGenDaoBriefGroup,
		`commandGenDaoBriefJsonCase`:        commandGenDaoBriefJsonCase,
	})
}

type (
	commandGenDaoInput struct {
		g.Meta         `name:"dao" config:"{commandGenDaoConfig}" usage:"{commandGenDaoUsage}" brief:"{commandGenDaoBrief}" eg:"{commandGenDaoEg}" ad:"{commandGenDaoAd}"`
		Path           string `name:"path"            short:"p" brief:"{commandGenDaoBriefPath}" d:"internal"`
		Link           string `name:"link"            short:"l" brief:"{commandGenDaoBriefLink}"`
		Tables         string `name:"tables"          short:"t" brief:"{commandGenDaoBriefTables}"`
		TablesEx       string `name:"tablesEx"        short:"e" brief:"{commandGenDaoBriefTablesEx}"`
		Group          string `name:"group"           short:"g" brief:"{commandGenDaoBriefGroup}" d:"default"`
		Prefix         string `name:"prefix"          short:"f" brief:"{commandGenDaoBriefPrefix}"`
		RemovePrefix   string `name:"removePrefix"    short:"r" brief:"{commandGenDaoBriefRemovePrefix}"`
		JsonCase       string `name:"jsonCase"        short:"j" brief:"{commandGenDaoBriefJsonCase}" d:"CamelLower"`
		ImportPrefix   string `name:"importPrefix"    short:"i" brief:"{commandGenDaoBriefImportPrefix}"`
		StdTime        bool   `name:"stdTime"         short:"s" brief:"{commandGenDaoBriefStdTime}"         orphan:"true"`
		GJsonSupport   bool   `name:"gJsonSupport"    short:"n" brief:"{commandGenDaoBriefGJsonSupport}"    orphan:"true"`
		OverwriteDao   bool   `name:"overwriteDao"    short:"o" brief:"{commandGenDaoBriefOverwriteDao}"    orphan:"true"`
		DescriptionTag bool   `name:"descriptionTag"  short:"d" brief:"{commandGenDaoBriefDescriptionTag}"  orphan:"true"`
		NoJsonTag      bool   `name:"noJsonTag"       short:"k" brief:"{commandGenDaoBriefNoJsonTag"        orphan:"true"`
		NoModelComment bool   `name:"noModelComment"  short:"m" brief:"{commandGenDaoBriefNoModelComment}"  orphan:"true"`
	}
	commandGenDaoOutput struct{}

	commandGenDaoInternalInput struct {
		commandGenDaoInput
		TableName    string // TableName specifies the table name of the table.
		NewTableName string // NewTableName specifies the prefix-stripped name of the table.
		ModName      string // ModName specifies the module name of current golang project, which is used for import purpose.
	}
)

func (c commandGen) Dao(ctx context.Context, in commandGenDaoInput) (out *commandGenDaoOutput, err error) {
	if g.Cfg().Available(ctx) {
		v := g.Cfg().MustGet(ctx, commandGenDaoConfig)
		if v.IsSlice() {
			for i := 0; i < len(v.Interfaces()); i++ {
				doGenDaoForArray(ctx, i, in)
			}
		} else {
			doGenDaoForArray(ctx, -1, in)
		}
	} else {
		doGenDaoForArray(ctx, -1, in)
	}
	mlog.Print("done!")
	return
}

// doGenDaoForArray implements the "gen dao" command for configuration array.
func doGenDaoForArray(ctx context.Context, index int, in commandGenDaoInput) {
	var (
		err     error
		db      gdb.DB
		modName string // Go module name, eg: github.com/gogf/gf.
	)
	if index >= 0 {
		err = g.Cfg().MustGet(
			ctx,
			fmt.Sprintf(`%s.%d`, commandGenDaoConfig, index),
		).Scan(&in)
		if err != nil {
			mlog.Fatalf(`invalid configuration of "%s": %+v`, commandGenDaoConfig, err)
		}
	}
	if dirRealPath := gfile.RealPath(in.Path); dirRealPath == "" {
		mlog.Fatalf(`path "%s" does not exist`, in.Path)
	}
	removePrefixArray := gstr.SplitAndTrim(in.RemovePrefix, ",")
	if in.ImportPrefix == "" {
		if !gfile.Exists("go.mod") {
			mlog.Fatal("go.mod does not exist in current working directory")
		}
		var (
			goModContent = gfile.GetContents("go.mod")
			match, _     = gregex.MatchString(`^module\s+(.+)\s*`, goModContent)
		)
		if len(match) > 1 {
			modName = gstr.Trim(match[1])
		} else {
			mlog.Fatal("module name does not found in go.mod")
		}
	}

	// It uses user passed database configuration.
	if in.Link != "" {
		tempGroup := gtime.TimestampNanoStr()
		match, _ := gregex.MatchString(`([a-z]+):(.+)`, in.Link)
		if len(match) == 3 {
			gdb.AddConfigNode(tempGroup, gdb.ConfigNode{
				Type: gstr.Trim(match[1]),
				Link: gstr.Trim(match[2]),
			})
			db, _ = gdb.Instance(tempGroup)
		}
	} else {
		db = g.DB(in.Group)
	}
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	var tableNames []string
	if in.Tables != "" {
		tableNames = gstr.SplitAndTrim(in.Tables, ",")
	} else {
		tableNames, err = db.Tables(context.TODO())
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
	}
	// Table excluding.
	if in.TablesEx != "" {
		array := garray.NewStrArrayFrom(tableNames)
		for _, v := range gstr.SplitAndTrim(in.TablesEx, ",") {
			array.RemoveValue(v)
		}
		tableNames = array.Slice()
	}

	// Generating dao & model go files one by one according to given table name.
	newTableNames := make([]string, len(tableNames))
	for i, tableName := range tableNames {
		newTableName := tableName
		for _, v := range removePrefixArray {
			newTableName = gstr.TrimLeftStr(newTableName, v, 1)
		}
		newTableName = in.Prefix + newTableName
		newTableNames[i] = newTableName
		// Dao.
		generateDao(ctx, db, commandGenDaoInternalInput{
			commandGenDaoInput: in,
			TableName:          tableName,
			NewTableName:       newTableName,
			ModName:            modName,
		})
	}
	// Model.
	generateEntity(ctx, db, tableNames, newTableNames, commandGenDaoInternalInput{
		commandGenDaoInput: in,
		ModName:            modName,
	})
}

// generateDaoContentFile generates the dao and model content of given table.
func generateDao(ctx context.Context, db gdb.DB, in commandGenDaoInternalInput) {
	// Generating table data preparing.
	fieldMap, err := db.TableFields(ctx, in.TableName)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", in.TableName, err)
	}
	var (
		dirRealPath             = gfile.RealPath(in.Path)
		dirPathDao              = gfile.Join(in.Path, "dao")
		tableNameCamelCase      = gstr.CaseCamel(in.NewTableName)
		tableNameCamelLowerCase = gstr.CaseCamelLower(in.NewTableName)
		tableNameSnakeCase      = gstr.CaseSnake(in.NewTableName)
		importPrefix            = in.ImportPrefix
	)
	if importPrefix == "" {
		if dirRealPath == "" {
			dirRealPath = in.Path
			importPrefix = dirRealPath
			importPrefix = gstr.Trim(dirRealPath, "./")
		} else {
			importPrefix = gstr.Replace(dirRealPath, gfile.Pwd(), "")
		}
		importPrefix = gstr.Replace(importPrefix, gfile.Separator, "/")
		importPrefix = gstr.Join(g.SliceStr{in.ModName, importPrefix}, "/")
		importPrefix, _ = gregex.ReplaceString(`\/{2,}`, `/`, gstr.Trim(importPrefix, "/"))
	}

	fileName := gstr.Trim(tableNameSnakeCase, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}

	// dao - index
	generateDaoIndex(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName, in)

	// dao - internal
	generateDaoInternal(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName, fieldMap, in)
}

func getImportPartContent(source string) string {
	var (
		packageImportsArray = garray.NewStrArray()
	)
	// Time package recognition.
	if strings.Contains(source, "gtime.Time") {
		packageImportsArray.Append(`"github.com/gogf/gf/v2/os/gtime"`)
	} else if strings.Contains(source, "time.Time") {
		packageImportsArray.Append(`"time"`)
	}

	if strings.Contains(source, "gjson.Json") {
		packageImportsArray.Append(`"github.com/gogf/gf/v2/encoding/gjson"`)
	}

	// Generate and write content to golang file.
	packageImportsStr := ""
	if packageImportsArray.Len() > 0 {
		packageImportsStr = fmt.Sprintf("import(\n%s\n)", packageImportsArray.Join("\n"))
	}
	return packageImportsStr
}

func generateEntity(ctx context.Context, db gdb.DB, tableNames, newTableNames []string, in commandGenDaoInternalInput) {
	var (
		modelContent string
		dirPathModel = gfile.Join(in.Path, "model")
	)

	// Model content.
	for i, tableName := range tableNames {
		fieldMap, err := db.TableFields(ctx, tableName)
		if err != nil {
			mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", in.TableName, err)
		}
		newTableName := newTableNames[i]
		modelContent += generateDaoModelStructContent(
			newTableName,
			gstr.CaseCamel(newTableName),
			"",
			generateStructDefinitionForModel(gstr.CaseCamel(newTableName), fieldMap, in),
		)
		modelContent += "\n"
	}

	// Generate and write content to golang file.
	modelContent = gstr.ReplaceByMap(getTplModelIndexContent(""), g.MapStrStr{
		"{TplPackageImports}": getImportPartContent(modelContent),
		"{TplModelStructs}":   modelContent,
	})
	var (
		err error
		//path = gfile.Join(dirPathModel, req.ModelFileName)
		path = gfile.Join(dirPathModel, "model.go")
	)
	err = gfile.PutContents(path, strings.TrimSpace(modelContent))
	if err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		utils.GoFmt(path)
		mlog.Print("generated:", path)
	}
}

// dto.
//func generateModelForDaoContentFile(ctx context.Context, db gdb.DB, tableNames, newTableNames []string, in commandGenDaoInternalInput) {
//	var (
//		modelContent string
//		dirPathModel = gfile.Join(in.Path, "model")
//	)
//	in.NoJsonTag = true
//	in.DescriptionTag = false
//	in.NoModelComment = false
//	// Model content.
//	for i, tableName := range tableNames {
//		fieldMap, err := db.TableFields(ctx, tableName)
//		if err != nil {
//			mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", in.TableName, err)
//		}
//		newTableName := newTableNames[i]
//
//		modelForDaoStructContent := generateStructDefinitionForModel(gstr.CaseCamel(newTableName), fieldMap, in)
//		// replace struct types from "Xxx" to "XxxForDao".
//		modelForDaoStructContent, _ = gregex.ReplaceStringFuncMatch(
//			`(type)\s+([A-Z]\w*?)\s+(struct\s+{)`,
//			modelForDaoStructContent,
//			func(match []string) string {
//				return fmt.Sprintf(`%s %sForDao %s`, match[1], match[2], match[3])
//			},
//		)
//		// replace all types to interface{}.
//		modelForDaoStructContent, _ = gregex.ReplaceStringFuncMatch(
//			"([A-Z]\\w*?)\\s+([\\w\\*\\.]+?)\\s+(//)",
//			modelForDaoStructContent,
//			func(match []string) string {
//				// If the type is already a pointer/slice/map, it does nothing.
//				if !gstr.HasPrefix(match[2], "*") && !gstr.HasPrefix(match[2], "[]") && !gstr.HasPrefix(match[2], "map") {
//					return fmt.Sprintf(`%s interface{} %s`, match[1], match[3])
//				}
//				return match[0]
//			},
//		)
//
//		modelContent += generateModelForDaoStructContent(
//			tableName,
//			gstr.CaseCamel(newTableName),
//			modelForDaoStructContent,
//		)
//		modelContent += "\n"
//	}
//	// Generate and write content to golang file.
//	modelContent = gstr.ReplaceByMap(consts.TemplateModelForDaoIndexContent, g.MapStrStr{
//		"{TplPackageImports}": getImportPartContent(modelContent),
//		"{TplModelStructs}":   modelContent,
//	})
//	var (
//		err  error
//		path = gfile.Join(dirPathModel, in.ModelFileNameForDao)
//	)
//	err = gfile.PutContents(path, strings.TrimSpace(modelContent))
//	if err != nil {
//		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
//	} else {
//		utils.GoFmt(path)
//		mlog.Print("generated:", path)
//	}
//}

func generateDaoModelStructContent(tableName, tableNameCamelCase, tplModelStructPath, structDefine string) string {
	return gstr.ReplaceByMap(getTplModelStructContent(tplModelStructPath), g.MapStrStr{
		"{TplTableName}":          tableName,
		"{TplTableNameCamelCase}": tableNameCamelCase,
		"{TplStructDefine}":       structDefine,
	})
}

func generateModelForDaoStructContent(tableName, tableNameCamelCase, structDefine string) string {
	return gstr.ReplaceByMap(consts.TemplateModelForDaoStructContent, g.MapStrStr{
		"{TplTableName}":          tableName,
		"{TplTableNameCamelCase}": tableNameCamelCase,
		"{TplStructDefine}":       structDefine,
	})
}

func generateDaoIndex(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName string, in commandGenDaoInternalInput) {
	path := gfile.Join(dirPathDao, fileName+".go")
	if in.OverwriteDao || !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(getTplDaoIndexContent(""), g.MapStrStr{
			"{TplImportPrefix}":            importPrefix,
			"{TplTableName}":               in.TableName,
			"{TplTableNameCamelCase}":      tableNameCamelCase,
			"{TplTableNameCamelLowerCase}": tableNameCamelLowerCase,
		})
		if err := gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			utils.GoFmt(path)
			mlog.Print("generated:", path)
		}
	}
}

func generateDaoInternal(
	tableNameCamelCase, tableNameCamelLowerCase, importPrefix string,
	dirPathDao, fileName string,
	fieldMap map[string]*gdb.TableField,
	in commandGenDaoInternalInput,
) {
	path := gfile.Join(dirPathDao, "internal", fileName+".go")
	modelContent := gstr.ReplaceByMap(getTplDaoInternalContent(""), g.MapStrStr{
		"{TplImportPrefix}":            importPrefix,
		"{TplTableName}":               in.TableName,
		"{TplGroupName}":               in.Group,
		"{TplTableNameCamelCase}":      tableNameCamelCase,
		"{TplTableNameCamelLowerCase}": tableNameCamelLowerCase,
		"{TplColumnDefine}":            gstr.Trim(generateColumnDefinitionForDao(fieldMap)),
		"{TplColumnNames}":             gstr.Trim(generateColumnNamesForDao(fieldMap)),
	})
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		utils.GoFmt(path)
		mlog.Print("generated:", path)
	}
}

// generateStructDefinitionForModel generates and returns the struct definition for specified table.
func generateStructDefinitionForModel(structName string, fieldMap map[string]*gdb.TableField, in commandGenDaoInternalInput) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKeyForDao(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructFieldForModel(field, in)
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
	stContent = gstr.Replace(stContent, "` ", "`")
	stContent = gstr.Replace(stContent, "``", "")
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	buffer.WriteString(stContent)
	buffer.WriteString("}")
	return buffer.String()
}

// generateStructFieldForModel generates and returns the attribute definition for specified field.
func generateStructFieldForModel(field *gdb.TableField, in commandGenDaoInternalInput) []string {
	var (
		typeName string
		jsonTag  = getJsonTagFromCase(field.Name, in.JsonCase)
	)
	t, _ := gregex.ReplaceString(`\(.+\)`, "", field.Type)
	t = gstr.Split(gstr.Trim(t), " ")[0]
	t = gstr.ToLower(t)
	switch t {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		typeName = "[]byte"

	case "bit", "int", "int2", "tinyint", "small_int", "smallint", "medium_int", "mediumint", "serial":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint"
		} else {
			typeName = "int"
		}

	case "int4", "int8", "big_int", "bigint", "bigserial":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint64"
		} else {
			typeName = "int64"
		}

	case "real":
		typeName = "float32"

	case "float", "double", "decimal", "smallmoney", "numeric":
		typeName = "float64"

	case "bool":
		typeName = "bool"

	case "datetime", "timestamp", "date", "time":
		if in.StdTime {
			typeName = "time.Time"
		} else {
			typeName = "*gtime.Time"
		}
	case "json":
		if in.GJsonSupport {
			typeName = "*gjson.Json"
		} else {
			typeName = "string"
		}
	default:
		// Automatically detect its data type.
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
			if in.StdTime {
				typeName = "time.Time"
			} else {
				typeName = "*gtime.Time"
			}
		default:
			typeName = "string"
		}
	}

	var (
		tagKey = "`"
		result = []string{
			"    #" + gstr.CaseCamel(field.Name),
			" #" + typeName,
		}
		descriptionTag = gstr.Replace(formatComment(field.Comment), `"`, `\"`)
	)

	result = append(result, " #"+fmt.Sprintf(tagKey+`json:"%s"`, jsonTag))
	result = append(result, " #"+fmt.Sprintf(`description:"%s"`+tagKey, descriptionTag))
	result = append(result, " #"+fmt.Sprintf(`// %s`, formatComment(field.Comment)))

	for k, v := range result {
		if in.NoJsonTag {
			v, _ = gregex.ReplaceString(`json:".+"`, ``, v)
		}
		if !in.DescriptionTag {
			v, _ = gregex.ReplaceString(`description:".*"`, ``, v)
		}
		if in.NoModelComment {
			v, _ = gregex.ReplaceString(`//.+`, ``, v)
		}
		result[k] = v
	}
	return result
}

// formatComment formats the comment string to fit the golang code without any lines.
func formatComment(comment string) string {
	comment = gstr.ReplaceByArray(comment, g.SliceStr{
		"\n", " ",
		"\r", " ",
	})
	comment = gstr.Replace(comment, `\n`, " ")
	comment = gstr.Trim(comment)
	return comment
}

// generateColumnDefinitionForDao generates and returns the column names definition for specified table.
func generateColumnDefinitionForDao(fieldMap map[string]*gdb.TableField) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyForDao(fieldMap)
	)
	for index, name := range names {
		var (
			field   = fieldMap[name]
			comment = gstr.Trim(gstr.ReplaceByArray(field.Comment, g.SliceStr{
				"\n", " ",
				"\r", " ",
			}))
		)
		array[index] = []string{
			"    #" + gstr.CaseCamel(field.Name),
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
			"            #" + gstr.CaseCamel(field.Name) + ":",
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

func getTplDaoIndexContent(tplDaoIndexPath string) string {
	if tplDaoIndexPath != "" {
		return gfile.GetContents(tplDaoIndexPath)
	}
	return consts.TemplateDaoDaoIndexContent
}

func getTplDaoInternalContent(tplDaoInternalPath string) string {
	if tplDaoInternalPath != "" {
		return gfile.GetContents(tplDaoInternalPath)
	}
	return consts.TemplateDaoDaoInternalContent
}

func getTplModelIndexContent(tplModelIndexPath string) string {
	if tplModelIndexPath != "" {
		return gfile.GetContents(tplModelIndexPath)
	}
	return consts.TemplateDaoModelIndexContent
}

func getTplModelStructContent(tplModelStructPath string) string {
	if tplModelStructPath != "" {
		return gfile.GetContents(tplModelStructPath)
	}
	return consts.TemplateDaoModelStructContent
}

// getJsonTagFromCase call gstr.Case* function to convert the s to specified case.
func getJsonTagFromCase(str, caseStr string) string {
	switch gstr.ToLower(caseStr) {
	case gstr.ToLower("Camel"):
		return gstr.CaseCamel(str)

	case gstr.ToLower("CamelLower"):
		return gstr.CaseCamelLower(str)

	case gstr.ToLower("Kebab"):
		return gstr.CaseKebab(str)

	case gstr.ToLower("KebabScreaming"):
		return gstr.CaseKebabScreaming(str)

	case gstr.ToLower("Snake"):
		return gstr.CaseSnake(str)

	case gstr.ToLower("SnakeFirstUpper"):
		return gstr.CaseSnakeFirstUpper(str)

	case gstr.ToLower("SnakeScreaming"):
		return gstr.CaseSnakeScreaming(str)
	}
	return str
}

func sortFieldKeyForDao(fieldMap map[string]*gdb.TableField) []string {
	names := make(map[int]string)
	for _, field := range fieldMap {
		names[field.Index] = field.Name
	}
	var (
		i      = 0
		j      = 0
		result = make([]string, len(names))
	)
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
