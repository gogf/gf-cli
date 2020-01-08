package get

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"os"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE    
    gf get [PACKAGE]

ARGUMENT 
    [PACKAGE]  remote golang package path, eg: github.com/gogf/gf
               it's optional, it updates GF version for current project in default
EXAMPLES
    gf get github.com/gogf/gf
    gf get github.com/gogf/gf@latest
    gf get github.com/gogf/gf@master
    gf get golang.org/x/sys

`))
}

func Run() {
	if len(os.Args) > 2 && os.Args[2] != "" {
		gproc.ShellRun(fmt.Sprintf(`go get -u %s`, os.Args[2]))
	} else {
		mlog.Print("downloading the latest version of GF...")
		gproc.ShellRun("go get -u github.com/gogf/gf")
	}
}
