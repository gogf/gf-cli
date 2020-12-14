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
    gf gen [TYPE] [OPTION]

TYPE
    dao     generate dao and model files.
    model   generate model files, note that these generated model files are different from model files 
            of command "gf gen dao".

OPTION
    -/--path             directory path for generated files.
    -l, --link           database configuration, please refer to: https://goframe.org/database/gdb/config
    -t, --tables         generate models only for given tables, multiple table names separated with ',' 
    -g, --group          specifying the configuration group name for database,
                         it's not necessary and the default value is "default"
    -c, --config         used to specify the configuration file for database, it's commonly not necessary.
                         If "-l" is not passed, it will search "./config.toml" and "./config/config.toml" 
                         in current working directory in default.
    -p, --prefix         add prefix for all table of specified link/database tables.
    -r, --remove-prefix  remove specified prefix of the table, multiple prefix separated with ',' 
    -m, --mod            module name for generated golang file imports.
                  
CONFIGURATION SUPPORT
    Options are also supported by configuration file. The configuration node name is "gf.gen", which also supports
    multiple databases, for example:
    [gfcli]
        [[gfcli.gen.dao]]
            link   = "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
            tables = "order,products"
        [[gfcli.gen.dao]]
            link   = "mysql:root:12345678@tcp(127.0.0.1:3306)/primary"
            path   = "./my-app"
            prefix = "primary_"
            tables = "user, userDetail"

EXAMPLES
    gf gen dao
        gf gen dao
        gf gen dao -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
        gf gen dao -path ./model -c config.yaml -g user-center -t user,user_detail,user_login
        gf gen dao -r user_
    gf gen model
        gf gen model
        gf gen model -l "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
        gf gen model -path ./model -c config.yaml -g user-center -t user,user_detail,user_login
        gf gen model -r user_

        

DESCRIPTION
    The "gen" command is designed for multiple generating purposes.
    It's currently supporting generating go files for ORM models.
`))
}

func Run() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"path":            true,
		"m,mod":           true,
		"l,link":          true,
		"t,tables":        true,
		"g,group":         true,
		"c,config":        true,
		"p,prefix":        true,
		"r,remove-prefix": true,
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
