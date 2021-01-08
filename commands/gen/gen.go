package gen

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/text/gstr"
)

func Help() {
	switch gcmd.GetArg(2) {
	case "dao":
		HelpDao()
	case "model":
		HelpModel()
	default:
		mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen TYPE [OPTION]

TYPE
    dao     generate dao and model files.
    model   generate model files, note that these generated model files are different from model files 
            of command "gf gen dao".

DESCRIPTION
    The "gen" command is designed for multiple generating purposes. 
    It's currently supporting generating go files for ORM models.
    Please use "gf gen dao -h" or "gf gen model -h" for specified type help.
`))
	}
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
		return
	}
	switch genType {
	case "model":
		doGenModel(parser)

	case "dao":
		doGenDao(parser)
	}
}
