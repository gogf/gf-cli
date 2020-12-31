package main

import (
	"fmt"
	"strings"

	"github.com/gogf/gf-cli/commands/env"
	"github.com/gogf/gf-cli/commands/mod"

	_ "github.com/gogf/gf-cli/boot"
	"github.com/gogf/gf-cli/commands/build"
	"github.com/gogf/gf-cli/commands/docker"
	"github.com/gogf/gf-cli/commands/fix"
	"github.com/gogf/gf-cli/commands/gen"
	"github.com/gogf/gf-cli/commands/get"
	"github.com/gogf/gf-cli/commands/initialize"
	"github.com/gogf/gf-cli/commands/install"
	"github.com/gogf/gf-cli/commands/pack"
	"github.com/gogf/gf-cli/commands/run"
	"github.com/gogf/gf-cli/commands/swagger"
	"github.com/gogf/gf-cli/commands/update"
	"github.com/gogf/gf-cli/common"
	"github.com/gogf/gf-cli/library/allyes"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf-cli/library/proxy"
	"github.com/gogf/gf/os/gbuild"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
)

func init() {
	// Automatically sets the golang proxy for all commands.
	proxy.AutoSet()
}

func main() {
	allyes.Init()

	command := gcmd.GetArg(1)
	// Help information
	if gcmd.ContainsOpt("h") && command != "" {
		help(command)
		return
	}
	switch command {
	case common.Help:
		help(gcmd.GetArg(2))
	case common.Version:
		version()
	case common.Env:
		env.Run()
	case common.Get:
		get.Run()
	case common.Gen:
		gen.Run()
	case common.Fix:
		fix.Run()
	case common.Mod:
		mod.Run()
	case common.Init:
		initialize.Run()
	case common.Pack:
		pack.Run()
	case common.Docker:
		docker.Run()
	case common.Swagger:
		swagger.Run()
	case common.Update:
		update.Run()
	case common.Install:
		install.Run()
	case common.Build:
		build.Run()
	case common.Run:
		run.Run()
	default:
		for k := range gcmd.GetOptAll() {
			switch k {
			case common.AsQues, common.AsHelp:
				mlog.Print(common.HelpContent)
				return
			case common.AsInfo, common.AsVersion:
				version()
				return
			}
		}
		// No argument or option, do installation checks.
		if !install.IsInstalled() {
			mlog.Print("hi, it seams it's the first time you installing gf cli.")
			s := gcmd.Scanf("do you want to install gf binary to your system? [y/n]: ")
			if strings.EqualFold(s, common.AsYes) {
				install.Run()
				gcmd.Scan("press <Enter> to exit...")
				return
			}
		}
		mlog.Print(common.HelpContent)
	}
}

// help shows more information for specified command.
func help(command string) {
	switch command {
	case common.Get:
		get.Help()
	case common.Gen:
		gen.Help()
	case common.Init:
		initialize.Help()
	case common.Docker:
		docker.Help()
	case common.Swagger:
		swagger.Help()
	case common.Build:
		build.Help()
	case common.Pack:
		pack.Help()
	case common.Run:
		run.Help()
	case common.Mod:
		mod.Help()
	default:
		mlog.Print(common.HelpContent)
	}
}

// version prints the version information of the cli tool.
func version() {
	info := gbuild.Info()
	if info["git"] == "" {
		info["git"] = "none"
	}
	mlog.Printf(`GoFrame CLI Tool %s, %s`, common.VERSION, common.Host)
	mlog.Printf(`Install Path: %s`, gfile.SelfPath())
	if info["gf"] == "" {
		mlog.Print(`Current is a custom installed version, no installation info.`)
		return
	}

	mlog.Print(gstr.Trim(fmt.Sprintf(`
Build Detail:
  Go Version:  %s
  GF Version:  %s
  Git Commit:  %s
  Build Time:  %s
`, info["go"], info["gf"], info["git"], info["time"])))
}
