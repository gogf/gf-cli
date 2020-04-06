package run

import (
	"github.com/gogf/gf/os/gproc"
	"strings"
)

func newProcess(path string, args string) *gproc.Process {
	a := strings.Split(args, " ")
	return gproc.NewProcess(path, a)
}
