package run

import (
	"fmt"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
	"os"
	"runtime"
	"strings"
	"time"
)

type App struct {
	File    string // Go run file name/path.
	Options string // Extra "go run" options.
}

const (
	gPROXY_CHECK_TIMEOUT = time.Second
)

var (
	process    *gproc.Process
	httpClient = ghttp.NewClient()
)

func init() {
	httpClient.SetTimeOut(gPROXY_CHECK_TIMEOUT)
}

func Help() {
	g.Log().Print(gstr.TrimLeft(`
USAGE
    gf run FILE [OPTIONS]

ARGUMENT
    FILE     building file path.
    OPTIONS  the same options as "go run" or "go build"

EXAMPLES
    gf run main.go
    gf run main.go -mod=vendor

DESCRIPTION
    The "run" command is used for running go codes with hot-compiled-like feature,
    which compiles and runs the go codes asynchronously when codes change.
`))
}

func New(file string) *App {
	app := &App{
		File: file,
	}
	if len(os.Args) > 3 {
		app.Options = strings.Join(os.Args[3:], " ")
	}
	return app
}

func Run() {
	file := gcmd.GetArg(2)
	if len(file) < 1 {
		g.Log().Fatal("file path cannot be empty")
	}
	app := New(file)
	dirty := gtype.NewBool()
	_, err := gfsnotify.Add(gfile.RealPath("."), func(event *gfsnotify.Event) {
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		// Print the event.
		g.Log().Print(event)
		// Variable <dirty> is used for running the changes only one in one second.
		if !dirty.Cas(false, true) {
			return
		}
		// With some delay in case of multiple code changes in very short interval.
		gtimer.SetTimeout(time.Second, func() {
			app.Run()
			dirty.Set(false)
		})

	})
	if err != nil {
		g.Log().Fatal(err)
	}
	go app.Run()
	select {}
}

func (app *App) Run() {
	// Rebuild and run the codes.
	g.Log().Printf("build: %s", app.File)
	outputPath := gfile.Join(gfile.TempDir(), "gf-cli", gfile.Name(app.File))
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	// Build the app.
	result, err := gproc.ShellExec(fmt.Sprintf(`go build -o %s %s %s`, outputPath, app.File, app.Options))
	if err != nil {
		g.Log().Printf("build error: \n%s%s", result, err.Error())
		return
	}
	// Kill the old process if build successfully.
	if process != nil {
		process.Kill()
	}
	process = gproc.NewProcess(outputPath, nil)
	if pid, err := process.Start(); err != nil {
		g.Log().Printf("build running error: %s", err.Error())
	} else {
		g.Log().Printf("build running pid: %d", pid)
	}
}
