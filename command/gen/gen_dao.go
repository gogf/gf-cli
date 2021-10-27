package gen

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/gcfg"
	"strings"

	"github.com/gogf/gf-cli/v2/library/mlog"
	"github.com/gogf/gf-cli/v2/library/utils"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/olekukonko/tablewriter"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

// generateDaoReq is the input parameter for generating dao.
type generateDaoReq struct {
	TableName           string // TableName specifies the table name of the table.
	NewTableName        string // NewTableName specifies the prefix-stripped name of the table.
	GroupName           string // GroupName specifies the group name of database configuration node for generated DAO.
	ModName             string // ModName specifies the module name of current golang project, which is used for import purpose.
	ImportPrefix        string // ImportPrefix is the custom import prefix for generated go files.
	JsonCase            string // JsonCase specifies the case of generated 'json' tag for model struct, value from gstr.Case* function names.
	DirPath             string // DirPath specifies the directory path for generated files.
	StdTime             bool   // StdTime defines using time.Time from stdlib instead of gtime.Time for generated time/date fields of tables.
	GJsonSupport        bool   // GJsonSupport defines using *gjson.Json instead of string for generated json fields of tables.
	OverwriteDao        bool   // Overwrite all dao files both inside/outside internal folder.
	DescriptionTag      bool   // Add comment to description tag for each field.
	NoJsonTag           bool   // No jso tag will be created for each field.
	NoModelComment      bool   // No model comment will be added for each field.
	ModelFileName       string // Custom name for storing generated model content.
	ModelFileNameForDao string // Custom file name generating model for DAO operations like Where/Data
	TplDaoIndexPath     string // TplDaoIndexPath specifies the file path for generating dao index files.
	TplDaoInternalPath  string // TplDaoInternalPath specifies the file path for generating dao internal files.
	TplModelIndexPath   string // TplModelIndexPath specifies the file path for generating model index content.
	TplModelStructPath  string // TplModelStructPath specifies the file path for generating model struct content.
}

const (
	genDaoDefaultPath          = "./app"
	nodeNameGenDaoInConfigFile = "gfcli.gen.dao"
	defaultModelIndexFileName  = "model.go"
)

func HelpDao() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen dao [OPTION]

OPTION
    -/--path             directory path for generated files.
    -l, --link           database configuration, the same as the ORM configuration of GoFrame.
    -t, --tables         generate models only for given tables, multiple table names separated with ',' 
    -e, --tablesEx       generate models excluding given tables, multiple table names separated with ',' 
    -c, --config         used to specify the configuration file for database, it's commonly not necessary.
    -g, --group          specifying the configuration group name of database for generated ORM instance,
                         it's not necessary and the default value is "default"
    -p, --prefix         add prefix for all table of specified link/database tables.
    -r, --removePrefix   remove specified prefix of the table, multiple prefix separated with ','
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
    -/--stdTime          use time.Time from stdlib instead of gtime.Time for generated time/date fields of tables.
    -/--gJsonSupport     use gJsonSupport to use *gjson.Json instead of string for generated json fields of tables.
    -/--importPrefix     custom import prefix for generated go files.
    -/--overwriteDao     overwrite all dao files both inside/outside internal folder
    -/--modelFile        custom file name for storing generated model content.
    -/--modelFileForDao  custom file name generating model for DAO operations like Where/Data. It's empty in default.
    -/--descriptionTag   add comment to description tag for each field.
    -/--noJsonTag        no json tag will be added for each field.
    -/--noModelComment   no model comment will be added for each field.
    -/--tplDaoIndex      template content for Dao index files generating.
    -/--tplDaoInternal   template content for Dao internal files generating.
    -/--tplModelIndex    template content for Model index files generating.
    -/--tplModelStruct   template content for Model internal files generating.
                  
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
		"path":           true,
		"m,mod":          true,
		"l,link":         true,
		"t,tables":       true,
		"e,tablesEx":     true,
		"g,group":        true,
		"c,config":       true,
		"p,prefix":       true,
		"r,removePrefix": true,
		"j,jsonCase":     true,
		"stdTime":        false,
		"gJsonSupport":   false,
		"overwriteDao":   false,
		"modelFile":      true,
		"modelForDao":    false,
		"descriptionTag": false,
		"noJsonTag":      false,
		"noModelComment": false,
		"tplDaoIndex":    true,
		"tplDaoInternal": true,
		"tplModelIndex":  true,
		"tplModelStruct": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	var (
		ctx    = context.TODO()
		config = g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	)
	if config.Available() {
		v := config.MustGet(ctx, nodeNameGenDaoInConfigFile)
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
		err                 error
		db                  gdb.DB
		modName             string                                                                           // Go module name, eg: github.com/gogf/gf.
		importPrefix        = getOptionOrConfigForDao(index, parser, "importPrefix")                         // The import prefix for generated go files.
		dirPath             = getOptionOrConfigForDao(index, parser, "path", genDaoDefaultPath)              // Generated directory path.
		tablesStr           = getOptionOrConfigForDao(index, parser, "tables")                               // Tables that will be generated.
		tablesEx            = getOptionOrConfigForDao(index, parser, "tablesEx")                             // Tables that will be excluded for generating.
		prefixName          = getOptionOrConfigForDao(index, parser, "prefix")                               // Add prefix to DAO and Model struct name.
		linkInfo            = getOptionOrConfigForDao(index, parser, "link")                                 // Custom database link.
		configPath          = getOptionOrConfigForDao(index, parser, "config")                               // Config file path, eg: ./config/db.toml.
		configGroup         = getOptionOrConfigForDao(index, parser, "group", "default")                     // Group name of database configuration node for generated DAO.
		removePrefix        = getOptionOrConfigForDao(index, parser, "removePrefix")                         // Remove prefix from table name.
		jsonCase            = getOptionOrConfigForDao(index, parser, "jsonCase", "CamelLower")               // Case configuration for 'json' tag.
		stdTime             = containsOptionOrConfigForDao(index, parser, "stdTime", false)                  // Use time.Time from stdlib instead of gtime.Time for generated time/date fields of tables.
		gJsonSupport        = containsOptionOrConfigForDao(index, parser, "gJsonSupport", false)             // Use gJsonSupport to use *gjson.Json instead of string for generated json fields of tables.
		overwriteDao        = containsOptionOrConfigForDao(index, parser, "overwriteDao", false)             // Overwrite all dao files both inside/outside internal folder
		modelFileName       = getOptionOrConfigForDao(index, parser, "modelFile", defaultModelIndexFileName) // Custom file name for storing generated model content.
		modelFileNameForDao = getOptionOrConfigForDao(index, parser, "modelFileForDao")                      // Custom file name generating model for DAO operations like Where/Data. It's empty in default.
		descriptionTag      = containsOptionOrConfigForDao(index, parser, "descriptionTag", false)           // Add comment to description tag for each field.
		noJsonTag           = containsOptionOrConfigForDao(index, parser, "noJsonTag", false)                // No json tag will be added for each field.
		noModelComment      = containsOptionOrConfigForDao(index, parser, "noModelComment", false)           // No model comment will be added for each field.
		tplDaoIndexPath     = getOptionOrConfigForDao(index, parser, "tplDaoIndex")                          // Template file path for generating dao index files.
		tplDaoInternalPath  = getOptionOrConfigForDao(index, parser, "tplDaoInternal")                       // Template file path for generating dao internal files.
		tplModelIndexPath   = getOptionOrConfigForDao(index, parser, "tplModelIndex")                        // Template file path for generating model index files.
		tplModelStructPath  = getOptionOrConfigForDao(index, parser, "tplModelStruct")                       // Template file path for generating model internal files.
	)
	if dirRealPath := gfile.RealPath(dirPath); dirRealPath == "" {
		mlog.Fatalf(`path "%s" does not exist`, dirPath)
	}
	if tplDaoIndexPath != "" && (!gfile.Exists(tplDaoIndexPath) || !gfile.IsReadable(tplDaoIndexPath)) {
		mlog.Fatalf("template file for dao index files generating does not exist or is not readable: %s", tplDaoIndexPath)
	}
	if tplDaoInternalPath != "" && (!gfile.Exists(tplDaoInternalPath) || !gfile.IsReadable(tplDaoInternalPath)) {
		mlog.Fatalf("template internal for dao internal files generating does not exist or is not readable: %s: %s", tplDaoInternalPath)
	}
	if tplModelIndexPath != "" && (!gfile.Exists(tplModelIndexPath) || !gfile.IsReadable(tplModelIndexPath)) {
		mlog.Fatalf("template file for model index files generating does not exist or is not readable: %s: %s", tplModelIndexPath)
	}
	if tplModelStructPath != "" && (!gfile.Exists(tplModelStructPath) || !gfile.IsReadable(tplModelStructPath)) {
		mlog.Fatalf("template file for model internal files generating does not exist or is not readable: %s: %s", tplModelStructPath)
	}
	// Make it compatible with old CLI version for option name: remove-prefix
	if removePrefix == "" {
		removePrefix = getOptionOrConfigForDao(index, parser, "remove-prefix")
	}
	removePrefixArray := gstr.SplitAndTrim(removePrefix, ",")
	if importPrefix == "" {
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
		var (
			config = g.Cfg().GetAdapter().(*gcfg.AdapterFile)
		)
		path, err := gfile.Search(configPath)
		if err != nil {
			mlog.Fatalf("search configuration file '%s' failed: %v", configPath, err)
		}
		if err := config.SetPath(gfile.Dir(path)); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", path, err)
		}
		config.SetFileName(gfile.Basename(path))
	}
	// It uses user passed database configuration.
	if linkInfo != "" {
		tempGroup := gtime.TimestampNanoStr()
		match, _ := gregex.MatchString(`([a-z]+):(.+)`, linkInfo)
		if len(match) == 3 {
			gdb.AddConfigNode(tempGroup, gdb.ConfigNode{
				Type: gstr.Trim(match[1]),
				Link: gstr.Trim(match[2]),
			})
			db, _ = gdb.Instance(tempGroup)
		}
	} else {
		db = g.DB(configGroup)
	}
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	var tableNames []string
	if tablesStr != "" {
		tableNames = gstr.SplitAndTrim(tablesStr, ",")
	} else {
		tableNames, err = db.Tables(context.TODO())
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
	}
	// Table excluding.
	if tablesEx != "" {
		array := garray.NewStrArrayFrom(tableNames)
		for _, v := range gstr.SplitAndTrim(tablesEx, ",") {
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
		newTableName = prefixName + newTableName
		newTableNames[i] = newTableName
		// Dao.
		generateDaoContentFile(db, generateDaoReq{
			TableName:          tableName,
			NewTableName:       newTableName,
			GroupName:          configGroup,
			ModName:            modName,
			ImportPrefix:       importPrefix,
			JsonCase:           jsonCase,
			DirPath:            dirPath,
			StdTime:            gconv.Bool(stdTime),
			GJsonSupport:       gJsonSupport,
			OverwriteDao:       overwriteDao,
			DescriptionTag:     descriptionTag,
			NoJsonTag:          noJsonTag,
			NoModelComment:     noModelComment,
			TplDaoIndexPath:    tplDaoIndexPath,
			TplDaoInternalPath: tplDaoInternalPath,
			TplModelIndexPath:  tplModelIndexPath,
			TplModelStructPath: tplModelStructPath,
		})
	}
	// Model.
	generateDaoModelContentFile(db, tableNames, newTableNames, generateDaoReq{
		JsonCase:           jsonCase,
		DirPath:            dirPath,
		StdTime:            gconv.Bool(stdTime),
		GJsonSupport:       gconv.Bool(gJsonSupport),
		DescriptionTag:     descriptionTag,
		NoJsonTag:          noJsonTag,
		NoModelComment:     noModelComment,
		ModelFileName:      modelFileName,
		TplDaoInternalPath: tplDaoInternalPath,
		TplModelIndexPath:  tplModelIndexPath,
		TplModelStructPath: tplModelStructPath,
	})
	// Model for dao.
	if modelFileNameForDao != "" {
		generateModelForDaoContentFile(db, tableNames, newTableNames, generateDaoReq{
			JsonCase:            jsonCase,
			DirPath:             dirPath,
			StdTime:             gconv.Bool(stdTime),
			GJsonSupport:        gconv.Bool(gJsonSupport),
			DescriptionTag:      descriptionTag,
			NoJsonTag:           noJsonTag,
			NoModelComment:      noModelComment,
			ModelFileName:       modelFileName,
			ModelFileNameForDao: modelFileNameForDao,
			TplDaoInternalPath:  tplDaoInternalPath,
			TplModelIndexPath:   tplModelIndexPath,
			TplModelStructPath:  tplModelStructPath,
		})
	}
}

// generateDaoContentFile generates the dao and model content of given table.
func generateDaoContentFile(db gdb.DB, req generateDaoReq) {
	// Generating table data preparing.
	fieldMap, err := db.TableFields(context.TODO(), req.TableName)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", req.TableName, err)
	}
	var (
		dirRealPath             = gfile.RealPath(req.DirPath)
		dirPathDao              = gfile.Join(req.DirPath, "dao")
		tableNameCamelCase      = gstr.CaseCamel(req.NewTableName)
		tableNameCamelLowerCase = gstr.CaseCamelLower(req.NewTableName)
		tableNameSnakeCase      = gstr.CaseSnake(req.NewTableName)
		importPrefix            = req.ImportPrefix
	)
	if importPrefix == "" {
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
	}

	fileName := gstr.Trim(tableNameSnakeCase, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}

	// dao - index
	generateDaoIndex(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName, req)

	// dao - internal
	generateDaoInternal(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName, fieldMap, req)
}

func generateDaoModelContentFile(db gdb.DB, tableNames, newTableNames []string, req generateDaoReq) {
	var (
		modelContent        string
		packageImportsArray = garray.NewStrArray()
		dirPathModel        = gfile.Join(req.DirPath, "model")
	)

	// Model content.
	for i, tableName := range tableNames {
		fieldMap, err := db.TableFields(context.TODO(), tableName)
		if err != nil {
			mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", req.TableName, err)
		}
		newTableName := newTableNames[i]
		modelContent += generateDaoModelStructContent(
			newTableName,
			gstr.CaseCamel(newTableName),
			req.TplModelStructPath,
			generateStructDefinitionForModel(gstr.CaseCamel(newTableName), fieldMap, req),
		)
		modelContent += "\n"
	}
	// Time package recognition.
	if strings.Contains(modelContent, "gtime.Time") {
		packageImportsArray.Append(`"github.com/gogf/gf/v2/os/gtime"`)
	} else if strings.Contains(modelContent, "time.Time") {
		packageImportsArray.Append(`"time"`)
	}

	if strings.Contains(modelContent, "gjson.Json") {
		packageImportsArray.Append(`"github.com/gogf/gf/v2/encoding/gjson"`)
	}

	// Generate and write content to golang file.
	packageImportsStr := ""
	if packageImportsArray.Len() > 0 {
		packageImportsStr = fmt.Sprintf("import(\n%s\n)", packageImportsArray.Join("\n"))
	}
	modelContent = gstr.ReplaceByMap(getTplModelIndexContent(req.TplModelIndexPath), g.MapStrStr{
		"{TplPackageImports}": packageImportsStr,
		"{TplModelStructs}":   modelContent,
	})
	var (
		err  error
		path = gfile.Join(dirPathModel, req.ModelFileName)
	)
	err = gfile.PutContents(path, strings.TrimSpace(modelContent))
	if err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		utils.GoFmt(path)
		mlog.Print("generated:", path)
	}
}

// model struct for dao.
func generateModelForDaoContentFile(db gdb.DB, tableNames, newTableNames []string, req generateDaoReq) {
	var (
		modelContent string
		dirPathModel = gfile.Join(req.DirPath, "model")
	)
	req.NoJsonTag = true
	req.DescriptionTag = false
	req.NoModelComment = false
	// Model content.
	for i, tableName := range tableNames {
		fieldMap, err := db.TableFields(context.TODO(), tableName)
		if err != nil {
			mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", req.TableName, err)
		}
		newTableName := newTableNames[i]

		modelForDaoStructContent := generateStructDefinitionForModel(gstr.CaseCamel(newTableName), fieldMap, req)
		// replace struct types from "Xxx" to "XxxForDao".
		modelForDaoStructContent, _ = gregex.ReplaceStringFuncMatch(`(type)\s+([A-Z]\w*?)\s+(struct\s+{)`, modelForDaoStructContent, func(match []string) string {
			return fmt.Sprintf(`%s %sForDao %s`, match[1], match[2], match[3])
		})
		// replace all types to interface{}.
		modelForDaoStructContent, _ = gregex.ReplaceString("([A-Z]\\w*?)\\s+([\\w\\*\\.]+?)\\s+(`orm)", "$1 interface{} $3", modelForDaoStructContent)

		modelContent += generateModelForDaoStructContent(
			tableName,
			gstr.CaseCamel(newTableName),
			modelForDaoStructContent,
		)
		modelContent += "\n"
	}
	// Generate and write content to golang file.
	modelContent = gstr.ReplaceByMap(templateModelForDaoIndexContent, g.MapStrStr{
		"{TplModelStructs}": modelContent,
	})
	var (
		err  error
		path = gfile.Join(dirPathModel, req.ModelFileNameForDao)
	)
	err = gfile.PutContents(path, strings.TrimSpace(modelContent))
	if err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		utils.GoFmt(path)
		mlog.Print("generated:", path)
	}
}

func generateDaoModelStructContent(tableName, tableNameCamelCase, tplModelStructPath, structDefine string) string {
	return gstr.ReplaceByMap(getTplModelStructContent(tplModelStructPath), g.MapStrStr{
		"{TplTableName}":          tableName,
		"{TplTableNameCamelCase}": tableNameCamelCase,
		"{TplStructDefine}":       structDefine,
	})
}

func generateModelForDaoStructContent(tableName, tableNameCamelCase, structDefine string) string {
	return gstr.ReplaceByMap(templateModelForDaoStructContent, g.MapStrStr{
		"{TplTableName}":          tableName,
		"{TplTableNameCamelCase}": tableNameCamelCase,
		"{TplStructDefine}":       structDefine,
	})
}

func generateDaoIndex(tableNameCamelCase, tableNameCamelLowerCase, importPrefix, dirPathDao, fileName string, req generateDaoReq) {
	path := gfile.Join(dirPathDao, fileName+".go")
	if req.OverwriteDao || !gfile.Exists(path) {
		indexContent := gstr.ReplaceByMap(getTplDaoIndexContent(req.TplDaoIndexPath), g.MapStrStr{
			"{TplImportPrefix}":            importPrefix,
			"{TplTableName}":               req.TableName,
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
	req generateDaoReq,
) {
	path := gfile.Join(dirPathDao, "internal", fileName+".go")
	modelContent := gstr.ReplaceByMap(getTplDaoInternalContent(req.TplDaoInternalPath), g.MapStrStr{
		"{TplImportPrefix}":            importPrefix,
		"{TplTableName}":               req.TableName,
		"{TplGroupName}":               req.GroupName,
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
func generateStructDefinitionForModel(structName string, fieldMap map[string]*gdb.TableField, req generateDaoReq) string {
	buffer := bytes.NewBuffer(nil)
	array := make([][]string, len(fieldMap))
	names := sortFieldKeyForDao(fieldMap)
	for index, name := range names {
		field := fieldMap[name]
		array[index] = generateStructFieldForModel(field, req)
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
func generateStructFieldForModel(field *gdb.TableField, req generateDaoReq) []string {
	var (
		typeName string
		jsonTag  = getJsonTagFromCase(field.Name, req.JsonCase)
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
		if req.StdTime {
			typeName = "time.Time"
		} else {
			typeName = "*gtime.Time"
		}
	case "json":
		if req.GJsonSupport {
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
			if req.StdTime {
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
		if req.NoJsonTag {
			v, _ = gregex.ReplaceString(`json:".+"`, ``, v)
		}
		if !req.DescriptionTag {
			v, _ = gregex.ReplaceString(`description:".*"`, ``, v)
		}
		if req.NoModelComment {
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

func getTplModelStructContent(tplModelStructPath string) string {
	if tplModelStructPath != "" {
		return gfile.GetContents(tplModelStructPath)
	}
	return templateDaoModelStructContent
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
// It returns the default value specified by parameter `value` is no value found.
func getOptionOrConfigForDao(index int, parser *gcmd.Parser, name string, defaultValue ...string) (result string) {
	var (
		ctx    = context.TODO()
		config = g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	)
	result = parser.GetOpt(name).String()
	if result == "" && config.Available() {
		config.SetViolenceCheck(true)
		if index >= 0 {
			result = config.MustGet(ctx, fmt.Sprintf(`%s.%d.%s`, nodeNameGenDaoInConfigFile, index, name)).String()
		} else {
			result = g.Cfg().MustGet(ctx, fmt.Sprintf(`%s.%s`, nodeNameGenDaoInConfigFile, name)).String()
		}
	}
	if result == "" && len(defaultValue) > 0 {
		result = defaultValue[0]
	}
	return
}

// containsOptionOrConfigForDao checks option value from parser and configuration file.
// It returns true if given `name` is in command option or configured `true` in configuration file.
func containsOptionOrConfigForDao(index int, parser *gcmd.Parser, name string, defaultValue ...bool) (result bool) {
	var (
		ctx    = context.TODO()
		config = g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	)
	result = parser.ContainsOpt(name)
	if !result && config.Available() {
		config.SetViolenceCheck(true)
		if index >= 0 {
			result = config.MustGet(ctx, fmt.Sprintf(`%s.%d.%s`, nodeNameGenDaoInConfigFile, index, name)).Bool()
		} else {
			result = config.MustGet(ctx, fmt.Sprintf(`%s.%s`, nodeNameGenDaoInConfigFile, name)).Bool()
		}
	}
	if !result && len(defaultValue) > 0 {
		result = defaultValue[0]
	}
	return
}
