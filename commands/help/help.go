package help

import (
	"fmt"
	"github.com/gogf/gf/g/text/gstr"
)

func Run() {
	help := `
Usage   : gf [command] [option]
Commands:
    ?,-?,-h,help        : this help.
    -v,-i,info,version  : show version info.
    init                : initialize an empty GF project at current working directory.
    install             : install gf binary to system (you may need root/admin permission).
`
	fmt.Println(gstr.Trim(help))
}
