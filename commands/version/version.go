package version

import (
	"github.com/gogf/gf-cli/library/mlog"
)

const (
	VERSION = "v0.1.0"
)

func Run() {
	mlog.Print("GoFrame CLI Tool Version", VERSION)
}
