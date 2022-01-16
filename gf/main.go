package main

import (
	_ "github.com/gogf/gf-cli/v2/internal/packed"

	"github.com/gogf/gf-cli/v2/internal/cmd"
	"github.com/gogf/gf-cli/v2/utility/allyes"
	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
)

func main() {
	defer func() {
		if exception := recover(); exception != nil {
			if err, ok := exception.(error); ok {
				mlog.Print(err.Error())
			} else {
				panic(exception)
			}
		}
	}()

	// zsh alias "git fetch" conflicts checks.
	handleZshAlias()

	// -y option checks.
	allyes.Init()

	var (
		ctx = gctx.New()
	)
	command, err := gcmd.NewFromObject(cmd.GF)
	if err != nil {
		panic(err)
	}
	err = command.AddObject(
		cmd.Env,
		cmd.Run,
		cmd.Gen,
		cmd.Tpl,
		cmd.Init,
		cmd.Pack,
		cmd.Build,
		cmd.Docker,
		cmd.Install,
		cmd.Version,
	)
	if err != nil {
		panic(err)
	}
	command.Run(ctx)
}

// zsh alias "git fetch" conflicts checks.
func handleZshAlias() {
	if home, err := gfile.Home(); err == nil {
		zshPath := gfile.Join(home, ".zshrc")
		if gfile.Exists(zshPath) {
			var (
				aliasCommand = `alias gf=gf`
				content      = gfile.GetContents(zshPath)
			)
			if !gstr.Contains(content, aliasCommand) {
				_ = gfile.PutContentsAppend(zshPath, "\n"+aliasCommand+"\n")
			}
		}
	}
}
