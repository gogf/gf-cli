package main

import (
	"github.com/gogf/gf-cli/commands/help"
	"github.com/gogf/gf-cli/commands/initialize"
	"github.com/gogf/gf-cli/commands/install"
	"github.com/gogf/gf-cli/commands/version"
	"github.com/gogf/gf/g/os/gcmd"
)

const (
	VERSION = "v0.0.1"
)

func main() {
	switch gcmd.Value.Get(1) {
	case "help":
		help.Run()
	case "info", "version":
		version.Run()
	case "init":
		initialize.Run()
	case "install":
		install.Run()
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
