package boot

import (
	_ "github.com/gogf/gf-cli/packed"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
)

func init() {
	g.Config("url").SetFileName("url.toml")
	handleZshAlias()
}

// zsh alias "git fetch" conflicts checks.
func handleZshAlias() {
	home, err := gfile.Home()
	if err == nil {
		zshPath := gfile.Join(home, ".zshrc")
		if gfile.Exists(zshPath) {
			aliasCommand := `alias gf=gf`
			content := gfile.GetContents(zshPath)
			if !gstr.Contains(content, aliasCommand) {
				_ = gfile.PutContentsAppend(zshPath, "\n"+aliasCommand)
			}
		}
	}
}
