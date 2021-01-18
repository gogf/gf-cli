package gen

import (
	"bytes"
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/olekukonko/tablewriter"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

// generateDaoReq is the input parameter for generating dao.
type generateDaoReq struct {
	TableName            string // TableName specifies the table name of the table.
	NewTableName         string // NewTableName specifies the prefix-stripped name of the table.
	PrefixName           string // PrefixName specifies the custom prefix name for generated dao and model struct.
	GroupName            string // GroupName specifies the group name of database configuration node for generated DAO.
	ModName              string // ModName specifies the module name of current golang project, which is used for import purpose.
	JsonCase             string // JsonCase specifies the case of generated 'json' tag for model struct, value from gstr.Case* function names.
	DirPath              string // DirPath specifies the directory path for generated files.
	TplDaoIndexPath      string // TplDaoIndexPath specifies the file path for generating dao index files.
	TplDaoInternalPath   string // TplDaoInternalPath specifies the file path for generating dao internal files.
	TplModelIndexPath    string // TplModelIndexPath specifies the file path for generating model index files.
	TplModelInternalPath string // TplModelInternalPath specifies the file path for generating model internal files.
}

const (
	genDaoDefaultPath          = "./app"
	nodeNameGenDaoInConfigFile = "gfcli.gen.dao"
)

func HelpDao() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen dao [OPTION]

OPTION
    -/--path             directory path for generated files.
    -l, --link           database configuration, the same as the ORM configuration of GoFrame.
    -t, --tables         generate models only for given tables, multiple table names separated with ',' 
    -g, --group          specifying the configuration group name of database for generated ORM instance,
                         it's not necessary and the default value is "default"
    -p, --prefix         add prefix for all table of specified link/database tables.
    -r, --removePrefix   remove specified prefix of the table, multiple prefix separated with ',' 
    -m, --mod            module name for generated golang file imports.
    -j, --jsonCase       generated json tag case for model struct, cases are as follows:
                         | Case            | Example            |
                         |---------------- |--------------------|
                         | Camel           | AnyKindOfString    | 
                         | CamelLower      | anyKindOfString    | default
                         | Snake           | any_kind_of_string |
                         | SnakeScreaming  | ANY_KIND_OF_STRING |
                         | SnakeFirstUpper | rgb_code_md5       |
                         | Kebab           | any-kind-of-string |
                         | KebabScreaming  | ANY-KIND-OF-STRING |
    -/--tplDaoIndex      template content for Dao index files generating.
    -/--tplDaoInternal   template content for Dao internal files generating.
    -/--tplModelIndex    template content for Model index files generating.
    -/--tplModelInternal template content for Model internal files generating.
                  
CONFIGURATION SUPPORT
    Options are also supported by configuration file.
    It's suggested using configuration file instead of command line arguments making producing. 
    The configuration node name is "gf.gen.dao", which also supports multiple databases, for example:
    [gfcli]
        [[gfcli.gen.dao]]
            link     = "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
            tables   = "order,products"
            jsonCase = "CamelLower"
        [[gfcli.gen.dao]]
            link   = "mysql:root:12345678@tcp(127.0.0.1:3306)/primary"
            path   = "./my-app"
            prefix = "primary_"
            tables = "user, userDetail"

EXAMPLES
    gf gen dao
    gf gen dao -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
    gf gen dao -path ./model -c config.yaml -g user-center -t user,user_detail,user_login
    gf gen dao -r user_
`))
}

// doGenDao implements the "gen dao" command.
func doGenDao() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"path":             true,
		"m,mod":            true,
		"l,link":           true,
		"t,tables":         true,
		"g,group":          true,
		"c,config":         true,
		"p,prefix":         true,
		"r,removePrefix":   true,
		"j,jsonCase":       true,
		"tplDaoIndex":      true,
		"tplDaoInternal":   true,
		"tplModelIndex":    true,
		"tplModelInternal": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}

	config := g.Cfg()
	if config.Available() {
		v := config.GetVar(nodeNameGenDaoInConfigFile)
		if v.IsEmpty() && g.IsEmpty(parser.GetOptAll()) {
			mlog.Fatal(`command arguments and configurations not found for generating dao files`)
		}
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
		err                  error
		db                   gdb.DB
		modName              = getOptionOrConfigForDao(index, parser, "mod")                     // Go module name, eg: github.com/gogf/gf.
		dirPath              = getOptionOrConfigForDao(index, parser, "path", genDaoDefaultPath) // Generated directory path.
		tablesStr            = getOptionOrConfigForDao(index, parser, "tables")                  // Tables that will be generated.
		prefixName           = getOptionOrConfigForDao(index, parser, "prefix")                  // Add prefix to DAO and Model struct name.
		linkInfo             = getOptionOrConfigForDao(index, parser, "link")                    // Custom database link.
		configPath           = getOptionOrConfigForDao(index, parser, "config")                  // Config file path, eg: ./config/db.toml.
		configGroup          = getOptionOrConfigForDao(index, parser, "group", "default")        // Group name of database configuration node for generated DAO.
		removePrefix         = getOptionOrConfigForDao(index, parser, "removePrefix")            // Remove prefix from table name.
		jsonCase             = getOptionOrConfigForDao(index, parser, "jsonCase", "CamelLower")  // Case configuration for 'json' tag.
		tplDaoIndexPath      = getOptionOrConfigForDao(index, parser, "tplDaoIndex")             // Template file path for generating dao index files.
		tplDaoInternalPath   = getOptionOrConfigForDao(index, parser, "tplDaoInternal")          // Template file path for generating dao internal files.
		tplModelIndexPath    = getOptionOrConfigForDao(index, parser, "tplModelIndex")           // Template file path for generating model index files.
		tplModelInternalPath = getOptionOrConfigForDao(index, parser, "tplModelInternal")        // Template file path for generating model internal files.
	)
	if tplDaoIndexPath != "" && (!gfile.Exists(tplDaoIndexPath) || !gfile.IsReadable(tplDaoIndexPath)) {
		mlog.Fatalf("template file for dao index files generating does not exist or is not readable: %s", tplDaoIndexPath)
	}
	if tplDaoInternalPath != "" && (!gfile.Exists(tplDaoInternalPath) || !gfile.IsReadable(tplDaoInternalPath)) {
		mlog.Fatalf("template internal for dao internal files generating does not exist or is not readable: %s: %s", tplDaoInternalPath)
	}
	if tplModelIndexPath != "" && (!gfile.Exists(tplModelIndexPath) || !gfile.IsReadable(tplModelIndexPath)) {
		mlog.Fatalf("template file for model index files generating does not exist or is not readable: %s: %s", tplModelIndexPath)
	}
	if tplModelInternalPath != "" && (!gfile.Exists(tplModelInternalPath) || !gfile.IsReadable(tplModelInternalPath)) {
		mlog.Fatalf("template file for model internal files generating does not exist or is not readable: %s: %s", tplModelInternalPath)
	}
	// Make it compatible with old CLI version for option name: remove-prefix
	if removePrefix == "" {
		removePrefix = getOptionOrConfigForDao(index, parser, "remove-prefix")
	}
	removePrefixArray := gstr.SplitAndTrim(removePrefix, ",")
	if modName == "" {
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
		req := &generateDaoReq{
			TableName:            tableName,
			NewTableName:         newTableName,
			PrefixName:           prefixName,
			GroupName:            configGroup,
			ModName:              modName,
			JsonCase:             jsonCase,
			DirPath:              dirPath,
			TplDaoIndexPath:      tplDaoIndexPath,
			TplDaoInternalPath:   tplDaoInternalPath,
			TplModelIndexPath:    tplModelIndexPath,
			TplModelInternalPath: tplModelInternalPath,
		}
		generateDaoAndModelContentFile(db, req)
	}
}

// generateDaoAndModelContentFile generates the dao and model content of given table.
func generateDaoAndModelContentFile(db gdb.DB, req *generateDaoReq) {
	fieldMap, err := db.TableFields(req.TableName)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", req.TableName, err)
	}
	// Change the `newTableName` if `prefixName` is given.
	newTableName := req.PrefixName + req.NewTableName
	var (
		dirPathDao              = gstr.Trim(gfile.Join(req.DirPath, "dao"), "./")
		dirPathModel            = gstr.Trim(gfile.Join(req.DirPath, "model"), "./")
		tableNameCamelCase      = gstr.CaseCamel(newTableName)
		tableNameCamelLowerCase = gstr.CaseCamelLower(newTableName)
		tableNameSnakeCase      = gstr.CaseSnake(newTableName)
		structDefine            = generateStructDefinitionForDao(tableNameCamelCase, fieldMap, req)
		packageImports          = ""
		importPrefix            = ""
		dirRealPath             = gfile.RealPath(req.DirPath)
	)
	if dirRealPath == "" {
		dirRealPath = req.DirPath
		importPrefix = dirRealPath
		importPrefix = gstr.Trim(dirRealPath, "./")
	} else {
		importPrefix = gstr.Replace(dirRealPath, gfile.Pwd(), "")
	}
	importPrefix = gstr.Replace(importPrefix, gfile.Separator, "/")
	importPrefix = gstr.Join(g.SliceStr{req.ModName, importPrefix}, "/")
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
		indexContent := gstr.ReplaceByMap(getTplModelIndexContent(req.TplModelIndexPath), g.MapStrStr{
			"{TplImportPrefix}":       importPrefix,
			"{TplTableName}":          req.TableName,
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
	entityContent := gstr.ReplaceByMap(getTplModelInternalContent(req.TplModelInternalPath), g.MapStrStr{
		"{TplTableName}":          req.TableName,
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
		indexContent := gstr.ReplaceByMap(getTplDaoIndexContent(req.TplDaoIndexPath), g.MapStrStr{
			"{TplImportPrefix}":            importPrefix,
			"{TplTableName}":               req.TableName,
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
	modelContent := gstr.ReplaceByMap(getTplDaoInternalContent(req.TplDaoInternalPath), g.MapStrStr{
		"{TplImportPrefix}":            importPrefix,
		"{TplTableName}":               req.TableName,
		"{TplGroupName}":               req.GroupName,
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
func generateStructDefinitionForDao(structName string, fieldMap map[string]*gdb.TableField, req *generateDaoReq) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKeyForDao(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructFieldForDao(field, req)
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
func generateStructFieldForDao(field *gdb.TableField, req *generateDaoReq) []string {
	var typeName, ormTag, jsonTag, comment string
	t, _ := gregex.ReplaceString(`\(.+\)`, "", field.Type)
	t = gstr.Split(gstr.Trim(t), " ")[0]
	t = gstr.ToLower(t)
	switch t {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		typeName = "[]byte"

	case "bit", "int", "tinyint", "small_int", "smallint", "medium_int", "mediumint", "serial":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint"
		} else {
			typeName = "int"
		}

	case "int8", "big_int", "bigint", "bigserial":
		if gstr.ContainsI(field.Type, "unsigned") {
			typeName = "uint64"
		} else {
			typeName = "int64"
		}

	case "real":
		typeName = "float32"

	case "float", "double", "decimal", "smallmoney":
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
	jsonTag = getJsonTagFromCase(field.Name, req.JsonCase)
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
	return templateDaoDaoIndexContent
}

func getTplDaoInternalContent(tplDaoInternalPath string) string {
	if tplDaoInternalPath != "" {
		return gfile.GetContents(tplDaoInternalPath)
	}
	return templateDaoDaoInternalContent
}

func getTplModelIndexContent(tplModelIndexPath string) string {
	if tplModelIndexPath != "" {
		return gfile.GetContents(tplModelIndexPath)
	}
	return templateDaoModelIndexContent
}

func getTplModelInternalContent(tplModelInternalPath string) string {
	if tplModelInternalPath != "" {
		return gfile.GetContents(tplModelInternalPath)
	}
	return templateDaoModelInternalContent
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
