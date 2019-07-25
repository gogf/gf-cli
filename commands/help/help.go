package help

import (
	"fmt"
	"github.com/gogf/gf/g/text/gstr"
)

func Run() {
	help := `
Usage   : gf [COMMAND] [OPTION]
Commands:
    ?,-?,-h,help  : this help.
    -v,-i,version : show version info.
    get           : install or update GF to system in default.
        [PKG]     : install specified golang package, it also supports 'go get' options.
    init          : initialize an empty GF project at current working directory in default.
        [NAME]    : name for current GF project, not necessary, default name is 'gf-app'.
    build         : cross-building go project.
        FILE          : building file path.
        -n, --name    : output binary name.
        -v, --version : output binary version.
        -a, --arch    : output binary architecture, multiple arch separated with ','.
        -o, --os      : output binary system, multiple os separated with ','.
        -p, --path    : output binary directory path, default is './bin'.
    update        : update current gf binary to latest one (you may need root/admin permission).
    install       : install gf binary to system (you may need root/admin permission).
    upgrade       : upgrade current project from older GF version to newer one if there's any compatibility issue.
        1.9       : upgrade to 1.9.x version, it recursively updates *.go files in current working directory.
`
	fmt.Println(gstr.Trim(help))
}
