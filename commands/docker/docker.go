package docker

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"os"
	"strings"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE    
    gf docker [FILE] [OPTIONS]

ARGUMENT
    FILE     file path for "gf build", it's "main.go" in default.
    OPTIONS  the same options as "docker build".

EXAMPLES
    gf docker 
    gf docker -t image
    gf docker main.go
    gf docker main.go -t image
    gf docker main.go -t registry.cn-hangzhou.aliyuncs.com/john/image:tag

DESCRIPTION
    The "docker" command builds the GF project to a docker images. It runs "docker build" 
    command automatically, so you should have docker command first.
    There must be a Dockerfile in the root of the project.

`))
}

func Run() {
	file := "main.go"
	extraOptions := ""
	if len(os.Args) > 2 {
		if gfile.ExtName(os.Args[2]) == "go" {
			file = os.Args[2]
			if len(os.Args) > 3 {
				extraOptions = strings.Join(os.Args[3:], " ")
			}
		} else {
			extraOptions = strings.Join(os.Args[2:], " ")
		}
	}
	err := gproc.ShellRun(fmt.Sprintf(`gf build %s -a amd64 -s linux`, file))
	if err != nil {
		return
	}
	gproc.ShellRun(fmt.Sprintf(`docker build . %s`, extraOptions))
}
