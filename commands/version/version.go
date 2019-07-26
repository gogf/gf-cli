package version

import (
	"github.com/gogf/gf-cli/library/mlog"
)

const (
	VERSION = "v0.2.0"
)

func Run() {
	mlog.Printf("GoFrame CLI Tool Version %s, https://goframe.org", VERSION)
}
