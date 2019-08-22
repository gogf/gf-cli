package install

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gfile"
	"runtime"
)

func Run() {
	binPath := "/usr/local/bin"
	if "windows" == runtime.GOOS {
		binPath = "C:\\Windows"
	}
	if gfile.Exists(binPath) {
		dst := binPath + gfile.Separator + "gf" + gfile.Ext(gfile.SelfPath())
		err := gfile.CopyFile(gfile.SelfPath(), dst)
		if err != nil {
			mlog.Fatalf("install gf binary to '%s' failed: %v", dst, err)
		} else {
			mlog.Printf("gf binary is successfully installed to: %s", dst)
		}
	} else {
		mlog.Fatal("'%s' does not exist", binPath)
	}
}
