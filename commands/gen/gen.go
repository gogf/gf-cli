package gen

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DEFAULT_GEN_PATH = "./api/model"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen model [PATH] [OPTION]

ARGUMENT
    PATH  the destination for storing generated files, not necessary, default is "./api/model"

OPTION
    -n, --name    package name for generated go files, it's the configuration group name in default
    -l, --link    database configuration, please refer to: https://goframe.org/database/gdb/config
    -c, --config  used to specify the configuration file for database, it's not necessary, 
                  if "-l" is not passed, it will search "./config.toml" and "./config/config.toml" 
                  in current working directory
    -g, --group   used with "-c" option, specifying the configuration group name for database,
                  it's not necessary, default value is "default"

EXAMPLES
    gf gen model
    gf gen model -n=dao
    gf gen model -l=mysql:root:12345678@tcp(127.0.0.1:3306)/test
    gf gen model ./model -l=mssql:sqlserver://sa:12345678@127.0.0.1:1433?database=test
    gf gen model ./model -c=config.yaml -g=user-center

DESCRIPTION
    The "gen" command is designed for multiple generating purposes.
    It's currently supporting generating go files for ORM model.
`))
}

func Run() {
	genType := gcmd.Value.Get(2)
	if genType == "" {
		mlog.Fatal("generating type cannot be empty")
	}
	genPath := gcmd.Value.Get(3, DEFAULT_GEN_PATH)
	linkInfo := gcmd.Option.Get("link", gcmd.Option.Get("l"))
	configFile := gcmd.Option.Get("config", gcmd.Option.Get("c"))
	configGroup := gcmd.Option.Get("group", gcmd.Option.Get("g", gdb.DEFAULT_GROUP_NAME))
	packageName := gcmd.Option.Get("name", gcmd.Option.Get("n", configGroup))

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

	realPath, err := gfile.Search(genPath)
	if err != nil {
		mlog.Fatalf("invalid generating path '%s': %v", genPath, err)
	}
	folderPath := realPath + gfile.Separator + packageName
	if err := gfile.Mkdir(folderPath); err != nil {
		mlog.Fatalf("mkdir for generating path '%s' failed: %v", folderPath, err)
	}

	db := g.DB(configGroup)
	tables, err := db.Tables()
	if err != nil {
		mlog.Fatalf("fetching tables failed: %v", err)
	}
	for _, table := range tables {
		generateModelFile(db, table, folderPath)
	}
}

func generateModelFile(db gdb.DB, table string, path string) {
	//fields, err := db.TableFields(table)
	//if err != nil {
	//	mlog.Fatalf("fetching tables fields failed for table '%s': %v", table, err)
	//}
	modelContent := gstr.ReplaceByMap(templateModel, g.MapStrStr{
		"{template}": table,
		"{Template}": gstr.CamelCase(table),
	})
	path = path + gfile.Separator + table + ".go"
	if err := gfile.PutContents(path, modelContent); err != nil {
		mlog.Fatalf("writing model content to '%s' failed: %v", path, err)
	}
}
