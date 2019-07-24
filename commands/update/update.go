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

func Run() {
	checkUrl := `https://goframe.org/cli/check`
	md5, err := gmd5.EncryptFile(gfile.SelfPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err.Error())
	}
	content := ghttp.GetContent(checkUrl, g.Map{
		"md5":  md5,
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	})
	switch content {
	case "0":
		fmt.Fprintln(os.Stdout, "it's the latest version, no need updates")
	case "1":
		downloadUrl := `https://goframe.org/cli/download`
		content := ghttp.GetContent(downloadUrl, g.Map{
			"md5":  md5,
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		})
	default:
		fmt.Fprintln(os.Stderr, "ERROR:", content)
	}
}
