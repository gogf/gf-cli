package initialize

import (
	"fmt"
	"github.com/gogf/gf/g/encoding/gcompress"
	"github.com/gogf/gf/g/net/ghttp"
	"github.com/gogf/gf/g/os/gcmd"
	"github.com/gogf/gf/g/os/gfile"
	"os"
)

const (
	emptyProject       = "github.com/gogf/gf-empty"
	emptyProjectName   = "gf-empty"
	emptyProjectZipUrl = "https://github.com/gogf/gf-empty/archive/master.zip"
	defaultProjectName = "gf-app"
)

func Run() {
	name := gcmd.Value.Get(2, defaultProjectName)
	fmt.Fprintln(os.Stdout, "initializing...")
	data := ghttp.GetBytes(emptyProjectZipUrl)
	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR:", "got empty project zip data, please tray again later")
		os.Exit(1)
	}
	err := gcompress.UnZipContent(data, ".", emptyProjectName+"-master")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err.Error())
		os.Exit(1)
	}
	if err = gfile.Replace(emptyProject, name, ".", "*.*", true); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err.Error())
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "initialization done! ")
	fmt.Fprintln(os.Stdout, "you can now run 'go run main.go' to start your journey, enjoy!")
}
