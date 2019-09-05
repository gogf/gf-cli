package initialize

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/encoding/gcompress"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
)

const (
	emptyProject       = "github.com/gogf/gf-empty"
	emptyProjectName   = "gf-empty"
	defaultProjectName = "gf-app"
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

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE    
    gf init [NAME]

ARGUMENT 
    NAME  name for current project, not necessary, default name is 'gf-app'

EXAMPLES
    gf init
    gf init my-project-name
`))
}

func Run() {
	parser, err := gcmd.Parse(nil)
	if err != nil {
		mlog.Fatal(err)
	}
	mlog.Print("initializing...")
	remoteMd5 := ghttp.GetContent(homeUrl + "/cli/project/md5")
	if remoteMd5 == "" {
		mlog.Fatal("get the project zip md5 failed")
	}
	name := parser.GetArg(2, defaultProjectName)
	zipUrl := cdnUrl + "/cli/project/zip?" + remoteMd5
	data := ghttp.GetBytes(zipUrl)
	if len(data) == 0 {
		mlog.Fatal("got empty project zip data, please tray again later")
	}
	if err = gcompress.UnZipContent(data, ".", emptyProjectName+"-master"); err != nil {
		mlog.Fatal("unzip project data failed,", err.Error())
	}
	if err = gfile.Replace(emptyProject, name, ".", "*.*", true); err != nil {
		mlog.Fatal("content replacing failed,", err.Error())
	}
	mlog.Print("initialization done! ")
	mlog.Print("you can now run 'go run main.go' to start your journey, enjoy!")
}
