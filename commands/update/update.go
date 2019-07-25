package update

import (
	"fmt"
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/crypto/gmd5"
	"github.com/gogf/gf/g/net/ghttp"
	"github.com/gogf/gf/g/os/gfile"
	"os"
	"runtime"
)

const (
	gCLI_URL = `https://goframe.org/cli/`
)

func Run() {
	checkUrl := gCLI_URL + `check`
	md5, err := gmd5.EncryptFile(gfile.SelfPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err.Error())
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "checking...")
	content := ghttp.GetContent(checkUrl, g.Map{
		"md5":  md5,
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	})
	switch content {
	case "0":
		fmt.Fprintln(os.Stdout, "it's the latest version, no need updates")
	case "1":
		fmt.Fprintln(os.Stdout, "downloading...")
		downloadUrl := gCLI_URL + `download`
		data := ghttp.GetBytes(downloadUrl, g.Map{
			"md5":  md5,
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		})
		if len(data) == 0 {
			fmt.Fprintln(os.Stderr, "ERROR:", "downloading failed, please try again later")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, "installing...")
		if err := gfile.PutBytes(gfile.SelfPath(), data); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err.Error())
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, "gf binary is now updated to the latest version")
	default:
		fmt.Fprintln(os.Stderr, "ERROR:", content)
	}
}
