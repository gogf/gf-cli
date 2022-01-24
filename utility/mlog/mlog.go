package mlog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	headerPrintEnvName = "GF_CLI_MLOG_HEADER"
)

var (
	ctx    = context.TODO()
	logger = glog.New()
)

func init() {
	logger.SetStack(false)
	if genv.Get(headerPrintEnvName).String() == "1" {
		logger.SetHeaderPrint(true)
	} else {
		logger.SetHeaderPrint(false)
	}
	if gcmd.GetOpt("debug") != nil || gcmd.GetOpt("gf.debug") != nil {
		logger.SetDebug(true)
	} else {
		logger.SetDebug(false)
	}
}

// SetHeaderPrint enables/disables header printing to stdout.
func SetHeaderPrint(enabled bool) {
	logger.SetHeaderPrint(enabled)
	if enabled {
		genv.Set(headerPrintEnvName, "1")
	} else {
		genv.Set(headerPrintEnvName, "0")
	}
}

func Print(v ...interface{}) {
	logger.Print(ctx, v...)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(ctx, format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(ctx, append(g.Slice{"ERROR:"}, v...)...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(ctx, "ERROR: "+format, v...)
}

func Debug(v ...interface{}) {
	logger.Debug(ctx, append(g.Slice{"DEBUG:"}, v...)...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(ctx, "DEBUG: "+format, v...)
}
