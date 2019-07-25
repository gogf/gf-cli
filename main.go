package main

import (
	"github.com/gogf/gf-cli/commands/build"
	"github.com/gogf/gf-cli/commands/get"
	"github.com/gogf/gf-cli/commands/help"
	"github.com/gogf/gf-cli/commands/initialize"
	"github.com/gogf/gf-cli/commands/install"
	"github.com/gogf/gf-cli/commands/update"
	"github.com/gogf/gf-cli/commands/upgrade"
	"github.com/gogf/gf-cli/commands/version"
	"github.com/gogf/gf/g/os/gcmd"
)

func main() {
	switch gcmd.Value.Get(1) {
	case "help":
		help.Run()
	case "version":
		version.Run()
	case "get":
		get.Run()
	case "init":
		initialize.Run()
	case "update":
		update.Run()
	case "install":
		install.Run()
	case "build":
		build.Run()
	case "upgrade":
		upgrade.Run()
	default:
		for k, _ := range gcmd.Option.GetAll() {
			switch k {
			case "?", "h":
				help.Run()
				return
			case "i", "v":
				version.Run()
				return
			}
		}
		help.Run()
	}
}
