package gen

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/text/gstr"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen model [PATH] [OPTION]

ARGUMENT
    PATH  the destination for storing generated files, not necessary, default is "./app/model"

OPTION
    -l, --link    database configuration, please refer to: https://goframe.org/database/gdb/config
    -t, --table   generate models only for given tables, multiple table names separated with ',' 
    -g, --group   used with "-c" option, specifying the configuration group name for database,
                  it's not necessary and the default value is "default"
    -c, --config  used to specify the configuration file for database, it's commonly not necessary.
                  If "-l" is not passed, it will search "./config.toml" and "./config/config.toml" 
                  in current working directory in default.
    -p, --prefix  remove specified prefix of the table, multiple prefix separated with ',' 
                  

EXAMPLES
    gf gen model
    gf gen model -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
    gf gen model ./model -l "mssql:sqlserver://sa:12345678@127.0.0.1:1433?database=test"
    gf gen model ./model -c config.yaml -g user-center -t user,user_detail,user_login
    gf gen model -p user_,p_

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
		"p,prefix": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	genType := parser.GetArg(2)
	if genType == "" {
		mlog.Print("generating type cannot be empty")
		Help()
		return
	}
	switch genType {
	case "model":
		doGenModel(parser)

	case "dao":
		doGenDao(parser)
	}
}
