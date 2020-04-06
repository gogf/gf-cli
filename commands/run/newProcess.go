// +build !windows

package run

import (
	"fmt"
	"github.com/gogf/gf/os/gproc"
)

func newProcess(path string, args string) *gproc.Process {
	command := fmt.Sprintf(`%s %s`, path, args)
	return gproc.NewProcessCmd(command, nil)
}
