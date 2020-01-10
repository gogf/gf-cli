package main

import (
	"fmt"
	"github.com/gogf/gf-cli/commands/docker"
	"github.com/gogf/gf-cli/library/proxy"
	"strings"

	_ "github.com/gogf/gf-cli/boot"
	"github.com/gogf/gf-cli/commands/build"
	"github.com/gogf/gf-cli/commands/fix"
	"github.com/gogf/gf-cli/commands/gen"
	"github.com/gogf/gf-cli/commands/get"
	"github.com/gogf/gf-cli/commands/initialize"
	"github.com/gogf/gf-cli/commands/install"
	"github.com/gogf/gf-cli/commands/pack"
	"github.com/gogf/gf-cli/commands/run"
	"github.com/gogf/gf-cli/commands/update"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gbuild"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/text/gstr"
)

const (
	// VERSION of gf-cli
	VERSION = "v0.5.1"
)

func init() {
	// Automatically sets the golang proxy for all commands.
	proxy.AutoSet()
}

var (
	helpContent = gstr.TrimLeft(`
USAGE
    gf COMMAND [ARGUMENT] [OPTION]

COMMAND
    get        install or update GF to system in default...
    gen        automatically generate go files for ORM models...
    run        running go codes with hot-compiled-like feature...
    init       initialize an empty GF project at current working directory...
    help       show more information about a specified command
    pack       packing any file/directory to a resource file, or a go file
    build      cross-building go project for lots of platforms...
    docker     create a docker image for current GF project...
    update     update current gf binary to latest one (might need root/admin permission)
    install    install gf binary to system (might need root/admin permission)
    version    show current binary version info

OPTION
    -?,-h      show this help or detail for specified command
    -v,-i      show version information

ADDITIONAL
    Use 'gf help COMMAND' or 'gf COMMAND -h' for detail about a command, which has '...' 
    in the tail of their comments.
`)
)

func main() {
	command := gcmd.GetArg(1)
	// Help information
	if gcmd.ContainsOpt("h") && command != "" {
		help(command)
		return
	}
	switch command {
	case "help":
		help(gcmd.GetArg(2))
	case "version":
		version()
	case "get":
		get.Run()
	case "gen":
		gen.Run()
	case "fix":
		fix.Run()
	case "init":
		initialize.Run()
	case "pack":
		pack.Run()
	case "docker":
		docker.Run()
	case "update":
		update.Run()
	case "install":
		install.Run()
	case "build":
		build.Run()
	case "run":
		run.Run()
	default:
		for k := range gcmd.GetOptAll() {
			switch k {
			case "?", "h":
				mlog.Print(helpContent)
				return
			case "i", "v":
				version()
				return
			}
		}
		// No argument or option, do installation checks.
		if !install.IsInstalled() {
			s := gcmd.Scanf("do you want to install gf binary to your system (%s)? [y/n]: ", install.GetInstallFolderPath())
			if strings.EqualFold(s, "y") {
				install.Run()
				gcmd.Scan("press <Enter> to exit...")
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
	case "docker":
		docker.Help()
	case "build":
		build.Help()
	case "pack":
		pack.Help()
	case "run":
		run.Help()
	default:
		mlog.Print(helpContent)
	}
}

// version prints the version information of the cli tool.
func version() {
	info := gbuild.Info()
	if info["git"] == "" {
		info["git"] = "none"
	}
	content := fmt.Sprintf(`
GoFrame CLI Tool %s, https://goframe.org
Built Detail:
  Go Version:  %s
  GF Version:  %s
  Git Commit:  %s
  Built Time:  %s
`, VERSION, info["go"], info["gf"], info["git"], info["time"])
	mlog.Print(gstr.Trim(content))
}
