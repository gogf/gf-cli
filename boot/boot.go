package boot

import (
	_ "github.com/gogf/gf-cli/packed"
	"github.com/gogf/gf/os/genv"

	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
)

func init() {
	// Force using configuration file in current working directory.
	// In case of source environment.
	genv.Set("GF_GCFG_PATH", gfile.Pwd())
	handleZshAlias()
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
