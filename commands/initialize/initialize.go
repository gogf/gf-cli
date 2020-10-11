package initialize

import (
	"github.com/gogf/gf-cli/library/allyes"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/encoding/gcompress"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"strings"
)

const (
	emptyProject     = "github.com/gogf/gf-empty"
	emptyProjectName = "gf-empty"
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
    gf init NAME

ARGUMENT 
    NAME  name for the project. It will create a folder with NAME in current directory.
          The NAME will also be the module name for the project.

EXAMPLES
    gf init my-app
    gf init my-project-name
`))
}

func Run() {
	parser, err := gcmd.Parse(nil)
	if err != nil {
		mlog.Fatal(err)
	}
	projectName := parser.GetArg(2)
	if projectName == "" {
		mlog.Fatal("project name should not be empty")
	}
	dirPath := projectName
	if !gfile.IsEmpty(dirPath) && !allyes.Check() {
		s := gcmd.Scanf(`the folder "%s" is not empty, files might be overwrote, continue? [y/n]: `, projectName)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	mlog.Print("initializing...")
	// MD5 retrieving.
	respMd5, err := ghttp.Get(homeUrl + "/cli/project/md5")
	if err != nil {
		mlog.Fatalf("get the project zip md5 failed: %s", err.Error())
	}
	defer respMd5.Close()
	md5DataStr := respMd5.ReadAllString()
	if md5DataStr == "" {
		mlog.Fatal("get the project zip md5 failed: empty md5 value. maybe network issue, try again?")
	}

	// Zip data retrieving.
	respData, err := ghttp.Get(cdnUrl + "/cli/project/zip?" + md5DataStr)
	if err != nil {
		mlog.Fatal("got the project zip data failed: %s", err.Error())
	}
	defer respData.Close()
	zipData := respData.ReadAll()
	if len(zipData) == 0 {
		mlog.Fatal("get the project data failed: empty data value. maybe network issue, try again?")
	}

	// Unzip the zip data.
	if err = gcompress.UnZipContent(zipData, dirPath, emptyProjectName+"-master"); err != nil {
		mlog.Fatal("unzip project data failed,", err.Error())
	}
	// Replace project name.
	if err = gfile.ReplaceDir(emptyProject, projectName, dirPath, "Dockerfile,*.go,*.MD,*.mod", true); err != nil {
		mlog.Fatal("content replacing failed,", err.Error())
	}
	if err = gfile.ReplaceDir(emptyProjectName, projectName, dirPath, "Dockerfile,*.go,*.MD,*.mod", true); err != nil {
		mlog.Fatal("content replacing failed,", err.Error())
	}
	mlog.Print("initialization done! ")
	mlog.Print("you can now run 'gf run main.go' to start your journey, enjoy!")
}
