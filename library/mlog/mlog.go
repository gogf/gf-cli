package mlog

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/glog"
)

var (
	logger = glog.New()
)

func init() {
	logger.SetStack(false)
	logger.SetDebug(false)
	logger.SetHeaderPrint(false)
	if gcmd.ContainsOpt("debug") {
		logger.SetDebug(true)
	}
}

func Print(v ...interface{}) {
	logger.Print(v...)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(append(g.Slice{"Error:"}, v...)...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf("Error: "+format, v...)
}

func Debug(v ...interface{}) {
	logger.Debug(append(g.Slice{"Debug:"}, v...)...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf("Debug: "+format, v...)
}
