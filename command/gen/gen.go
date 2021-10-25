package gen

import (
	"github.com/gogf/gf-cli/v2/library/mlog"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/text/gstr"
)

func Help() {
	switch gcmd.GetArg(2).String() {
	case "dao":
		HelpDao()

	case "pb":
		HelpPb()

	case "pbentity":
		HelpPbEntity()

	default:
		mlog.Print(gstr.TrimLeft(`
USAGE 
    gf gen TYPE [OPTION]

TYPE
    dao        generate dao and model files.
    pb         parse proto files and generate protobuf go files.
    pbentity   generate entity message files in protobuf3 format.

DESCRIPTION
    The "gen" command is designed for multiple generating purposes. 
    It's currently supporting generating go files for ORM models, protobuf and protobuf entity files.
    Please use "gf gen dao -h" or "gf gen model -h" for specified type help.
`))
	}
}

func Run() {
	genType := gcmd.GetArg(2).String()
	if genType == "" {
		mlog.Print("generating type cannot be empty")
		return
	}
	switch genType {
	case "dao":
		doGenDao()

	case "pb":
		doGenPb()

	case "pbentity":
		doGenPbEntity()
	}
}
