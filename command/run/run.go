package run

import (
	"fmt"
	"github.com/gogf/gf-cli/v2/library/mlog"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gtype"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gfsnotify"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/text/gstr"
	"os"
	"runtime"
	"strings"
	"time"
)

type App struct {
	File    string // Go run file name/path.
	Options string // Extra "go run" options.
	Args    string // Custom arguments.
}

const (
	proxyCheckTimeout = time.Second
)

var (
	process    *gproc.Process
	httpClient = ghttp.NewClient()
)

func init() {
	httpClient.SetTimeout(proxyCheckTimeout)
}

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE
    gf run FILE [OPTION]

ARGUMENT
    FILE    building file path.
    OPTION  the same options as "go run"/"go build" except some options as follows defined

OPTION
    -/--args     custom process arguments.

EXAMPLES
    gf run main.go
    gf run main.go --args "server -p 8080"
    gf run main.go -mod=vendor

DESCRIPTION
    The "run" command is used for running go codes with hot-compiled-like feature,
    which compiles and runs the go codes asynchronously when codes change.
`))
}

func Run() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"args": true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	mlog.SetHeaderPrint(true)
	file := gcmd.GetArg(2)
	if len(file) < 1 {
		mlog.Fatal("file path cannot be empty")
	}
	app := &App{
		File: file,
	}
	// ================================================================================
	// This command is very special that it supports options of "go run" and "go build"
	// from the third parameter of os.Args. That means, we should filter any parameter
	// that "go run" and "go build" do not allow.
	// ================================================================================
	// Swagger checks.
	array := garray.NewStrArrayFrom(os.Args)
	// args checks.
	args := parser.GetOpt("args")
	if args != "" {
		app.Args = args
		index := -1
		array.Iterator(func(k int, v string) bool {
			if gstr.Contains(v, "-args") {
				index = k
				return false
			}
			return true
		})
		if index != -1 {
			v, _ := array.Get(index)
			if gstr.Contains(v, "=") {
				array.Remove(index)
			} else {
				array.Remove(index)
				array.Remove(index)
			}
		}
	}
	// -y checks
	array.RemoveValue("-y")
	array.RemoveValue("--y")
	if array.Len() > 3 {
		app.Options = strings.Join(array.SubSlice(3), " ")
	}
	dirty := gtype.NewBool()
	_, err = gfsnotify.Add(gfile.RealPath("."), func(event *gfsnotify.Event) {
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		// Variable `dirty` is used for running the changes only one in one second.
		if !dirty.Cas(false, true) {
			return
		}
		// With some delay in case of multiple code changes in very short interval.
		gtimer.SetTimeout(1500*gtime.MS, func() {
			defer dirty.Set(false)
			mlog.Printf(`go file changes: %s`, event.String())
			app.Run()
		})
	})
	if err != nil {
		mlog.Fatal(err)
	}
	go app.Run()
	select {}
}

func (app *App) Run() {
	// Rebuild and run the codes.
	renamePath := ""
	mlog.Printf("build: %s", app.File)
	outputPath := gfile.Join("bin", gfile.Name(app.File))
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
		if gfile.Exists(outputPath) {
			renamePath = outputPath + "~"
			if err := gfile.Rename(outputPath, renamePath); err != nil {
				mlog.Print(err)
			}
		}
	}
	// In case of `pipe: too many open files` error.
	// Build the app.
	buildCommand := fmt.Sprintf(`go build -o %s %s %s`, outputPath, app.Options, app.File)
	mlog.Print(buildCommand)
	result, err := gproc.ShellExec(buildCommand)
	if err != nil {
		mlog.Printf("build error: \n%s%s", result, err.Error())
		return
	}
	// Kill the old process if build successfully.
	if process != nil {
		if err := process.Kill(); err != nil {
			mlog.Debugf("kill process error: %s", err.Error())
			//return
		}
	}
	// Run the binary file.
	runCommand := fmt.Sprintf(`%s %s`, outputPath, app.Args)
	mlog.Print(runCommand)
	if runtime.GOOS == "windows" {
		// Special handling for windows platform.
		// DO NOT USE "cmd /c" command.
		process = gproc.NewProcess(runCommand, nil)
	} else {
		process = gproc.NewProcessCmd(runCommand, nil)
	}
	if pid, err := process.Start(); err != nil {
		mlog.Printf("build running error: %s", err.Error())
	} else {
		mlog.Printf("build running pid: %d", pid)
	}
}
