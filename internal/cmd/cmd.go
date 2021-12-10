package cmd

import (
	"context"
	"fmt"

	"github.com/gogf/gf-cli/v2/internal/consts"
	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gbuild"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	GF = commandGF{}
)

type commandGF struct {
	g.Meta `name:"gf" usage:"{commandGFUsage}" ad:"{commandGFAd}"`
}

const (
	commandGFUsage = `gf COMMAND [ARGUMENT] [OPTION]`
	commandGFAd    = `
ADDITIONAL
    Use "gf COMMAND -h" for details about a command.
`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandGFUsage`: commandGFUsage,
		`commandGFAd`:    commandGFAd,
	})
}

type commandGFInput struct {
	g.Meta  `name:"gf"`
	Yes     bool `short:"y" name:"yes"     brief:"all yes for all command without prompt ask"   orphan:"true"`
	Version bool `short:"v" name:"version" brief:"show version information"                     orphan:"true"`
	Debug   bool `short:"d" name:"debug"   brief:"show internal detailed debugging information" orphan:"true"`
}
type commandGFOutput struct{}

func (c commandGF) Index(ctx context.Context, in commandGFInput) (out *commandGFOutput, err error) {
	if in.Version {
		c.printVersion()
		return
	}
	gcmd.CommandFromCtx(ctx).Print()
	// No argument or option, do installation checks.
	//if !install.IsInstalled() {
	//	mlog.Print("hi, it seams it's the first time you installing gf cli.")
	//	s := gcmd.Scanf("do you want to install gf binary to your system? [y/n]: ")
	//	if strings.EqualFold(s, "y") {
	//		install.Run()
	//		gcmd.Scan("press `Enter` to exit...")
	//		return
	//	}
	//}
	return
}

// version prints the version information of the cli tool.
func (c commandGF) printVersion() {
	info := gbuild.Info()
	if info["git"] == "" {
		info["git"] = "none"
	}
	mlog.Printf(`GoFrame CLI Tool %s, https://goframe.org`, consts.Version)
	gfVersion, err := c.getGFVersionOfCurrentProject()
	if err != nil {
		gfVersion = err.Error()
	} else {
		gfVersion = gfVersion + " in current go.mod"
	}
	mlog.Printf(`GoFrame Version: %s`, gfVersion)
	mlog.Printf(`CLI Installed At: %s`, gfile.SelfPath())
	if info["gf"] == "" {
		mlog.Print(`Current is a custom installed version, no installation information.`)
		return
	}

	mlog.Print(gstr.Trim(fmt.Sprintf(`
CLI Built Detail:
  Go Version:  %s
  Git Commit:  %s
  Build Time:  %s
`, info["go"], info["git"], info["time"])))
}

// getGFVersionOfCurrentProject checks and returns the GoFrame version current project using.
func (c commandGF) getGFVersionOfCurrentProject() (string, error) {
	goModPath := gfile.Join(gfile.Pwd(), "go.mod")
	if gfile.Exists(goModPath) {
		match, err := gregex.MatchString(`github.com/gogf/gf\s+([\w\d\.]+)`, gfile.GetContents(goModPath))
		if err != nil {
			return "", err
		}
		if len(match) > 1 {
			return match[1], nil
		}
		return "", gerror.New("cannot find goframe requirement in go.mod")
	} else {
		return "", gerror.New("cannot find go.mod")
	}
}
