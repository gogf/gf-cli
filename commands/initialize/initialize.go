package initialize

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/g"
	"github.com/gogf/gf/g/encoding/gcompress"
	"github.com/gogf/gf/g/net/ghttp"
	"github.com/gogf/gf/g/os/gcmd"
	"github.com/gogf/gf/g/os/gfile"
)

const (
	emptyProject       = "github.com/gogf/gf-empty"
	emptyProjectName   = "gf-empty"
	defaultProjectName = "gf-app"
)

var (
	cdnUrl  = g.Config().GetString("cdn.url")
	homeUrl = g.Config().GetString("home.url")
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
	mlog.Print("initializing...")
	remoteMd5 := ghttp.GetContent(homeUrl + "/project/md5")
	if remoteMd5 == "" {
		mlog.Fatal("get the project zip md5 failed")
	}
	name := gcmd.Value.Get(2, defaultProjectName)
	zipUrl := cdnUrl + "/project/zip?" + remoteMd5
	data := ghttp.GetBytes(zipUrl)
	if len(data) == 0 {
		mlog.Fatal("got empty project zip data, please tray again later")
	}
	err := gcompress.UnZipContent(data, ".", emptyProjectName+"-master")
	if err != nil {
		mlog.Fatal("unzip project data failed,", err.Error())
	}
	if err = gfile.Replace(emptyProject, name, ".", "*.*", true); err != nil {
		mlog.Fatal("content replacing failed,", err.Error())
	}
	mlog.Print("initialization done! ")
	mlog.Print("you can now run 'go run main.go' to start your journey, enjoy!")
}
