package get

import (
	"github.com/gogf/gf/g/os/genv"
	"github.com/gogf/gf/g/os/gproc"
)

func Run() {
	if p := genv.Get("GOPROXY"); p == "" {
		genv.Set("GOPROXY", "https://mirrors.aliyun.com/goproxy/")
	}
	gproc.ShellRun("go clean -modcache")
	gproc.ShellRun("go get -u github.com/gogf/gf")
}
