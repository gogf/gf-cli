package main

import (
	"fmt"

	_ "github.com/gogf/gf-cli/boot"
	"github.com/gogf/gf-cli/commands/build"
	"github.com/gogf/gf-cli/commands/fix"
	"github.com/gogf/gf-cli/commands/gen"
	"github.com/gogf/gf-cli/commands/get"
	"github.com/gogf/gf-cli/commands/initialize"
	"github.com/gogf/gf-cli/commands/install"
	"github.com/gogf/gf-cli/commands/pack"
	"github.com/gogf/gf-cli/commands/update"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/text/gstr"
)

const (
	VERSION = "v0.3.2"
)

var (
	verContent  = fmt.Sprintf("GoFrame CLI Tool Version %s, https://goframe.org", VERSION)
	helpContent = gstr.TrimLeft(`
USAGE
    gf COMMAND [ARGUMENT] [OPTION]

COMMAND
    get        install or update GF to system in default...
    gen        automatically generate go files for ORM models...
    init       initialize an empty GF project at current working directory...
    help       show more information about a specified command
    pack       packing any file/directory to a resource file, or a go file
    build      cross-building go project for lots of platforms...
    update     update current gf binary to latest one (you may need root/admin permission)
    install    install gf binary to system (you may need root/admin permission)
    version    show version info

OPTION
    -?,-h      show this help or detail for specified command
    -v,-i      show version information

ADDITIONAL
    Use 'gf help COMMAND' or 'gf COMMAND -h' for detail about a command, which has '...' in the tail of their comments.
`)
)

func main() {
	command := gcmd.GetArg(1)
	param := gcmd.GetArg(2)
	/*
		Parameter: param
		Optional value: "-vscode"/"-sublime"/"-goland"
		Description: This parameter is used to select whether to keep the. vscode folder to provide goframe snippets for the current folder.
					 Usually used when developing goframe projects in vscode editors.
					 Other parameters are temporarily unavailable.
	*/
	// Help information
	if gcmd.ContainsOpt("h") && command != "" {
		help(command)
		return
	}
	switch command {
	case "help":
		help(gcmd.GetArg(2))
	case "version":
		mlog.Print(verContent)
	case "get":
		get.Run()
	case "gen":
		gen.Run()
	case "fix":
		fix.Run()
	case "init":
		initialize.Run(param)
	case "pack":
		pack.Run()
	case "update":
		update.Run()
	case "install":
		install.Run()
	case "build":
		build.Run()
	default:
		for k := range gcmd.GetOptAll() {
			switch k {
			case "?", "h":
				mlog.Print(helpContent)
				return
			case "i", "v":
				mlog.Print(verContent)
				return
			}
		}
		mlog.Print(helpContent)
	}
}

// help shows more information for specified command.
func help(command string) {
	switch command {
	case "get":
		get.Help()
	case "gen":
		gen.Help()
	case "init":
		initialize.Help()
	case "build":
		build.Help()
	case "pack":
		pack.Help()
	default:
		mlog.Print(helpContent)
	}
}
