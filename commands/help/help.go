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
    get           : install or update GF to system.
    init          : initialize an empty GF project at current working directory.
        [NAME]    : name for current GF project, not necessary, default name is 'gf-app'.
	update        : update current gf binary to latest one.
    install       : install gf binary to system (you may need root/admin permission).
    upgrade       : upgrade current project from older GF version to newer one if there's any compatibility issue.
        1.9       : upgrade to 1.9.x version, it will automatically change *.go files.
    compile       : cross-compile go project.
        FILE          : compiling file path.
        -n, --name    : binary name.
        -v, --version : binary version.
        -a, --arch    : architecture for compiling, multiple.
            --os      : target operation system for compiling.
`
	fmt.Println(gstr.Trim(help))
}
