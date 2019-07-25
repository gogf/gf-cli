package install

import (
	"fmt"
	"github.com/gogf/gf/g/os/gfile"
	"os"
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
			fmt.Fprintf(os.Stderr, "ERROR: install gf binary to '%s' failed: %v\n", dst, err)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stdout, "gf binary is successfully installed to: %s\n", dst)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: '%s' does not exist\n", binPath)
		os.Exit(1)
	}
}
