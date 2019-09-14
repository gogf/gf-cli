package update

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"runtime"
)

var (
	cdnUrl  = g.Config("url").GetString("cdn.url")
	homeUrl = g.Config("url").GetString("home.url")
)

func init() {
	if cdnUrl == "" {
		mlog.Fatal("CDN configuration cannot be empty")
	}
	if homeUrl == "" {
		mlog.Fatal("Home configuration cannot be empty")
	}
}

func Run() {
	mlog.Print("checking...")
	md5Url := homeUrl + `/cli/binary/md5`
	latestMd5 := ghttp.GetContent(md5Url, g.Map{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	})
	if latestMd5 == "" {
		mlog.Fatal("get the latest binary md5 failed")
	}
	localMd5, err := gmd5.EncryptFile(gfile.SelfPath())
	if err != nil {
		mlog.Fatal("calculate local binary md5 failed,", err.Error())
	}
	if localMd5 != latestMd5 {
		mlog.Print("downloading...")
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		downloadUrl := fmt.Sprintf(
			`%s/cli/binary/%s_%s/gf%s?%s`,
			cdnUrl,
			runtime.GOOS,
			runtime.GOARCH,
			ext,
			latestMd5,
		)
		data := ghttp.GetBytes(downloadUrl)
		if len(data) == 0 {
			mlog.Fatal("downloading failed for", runtime.GOOS, runtime.GOARCH)
		}
		mlog.Print("installing...")
		if err := gfile.PutBytes(gfile.SelfPath(), data); err != nil {
			mlog.Fatal("installing binary failed,", err.Error())
		}
		mlog.Print("gf binary is now updated to the latest version")
	} else {
		mlog.Print("it's the latest version, no need updates")
	}
}
