package run

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
	"runtime"
	"time"
)

type App struct {
	File string
}

var (
	process *gproc.Process
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE
    gf run FILE

ARGUMENT
    FILE  building file path.

EXAMPLES
    gf run main.go

DESCRIPTION
    The "run" command is used for running go codes with hot-compiled-like feature,
    which compiles and runs the go codes asynchronously when codes change.
`))
}

func New(file string) *App {
	return &App{
		File: file,
	}
}

func Run() {
	file := gcmd.GetArg(2)
	if len(file) < 1 {
		mlog.Fatal("file path cannot be empty")
	}
	app := New(file)
	dirty := gtype.NewBool()
	_, err := gfsnotify.Add(gfile.RealPath("."), func(event *gfsnotify.Event) {
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		// Print the event.
		mlog.Print(event)
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
		mlog.Fatal(err)
	}
	go app.Run()
	select {}
}

func (app *App) Run() {
	// Kill the old process.
	if process != nil {
		process.Kill()
	}
	// Rebuild and run the codes.
	mlog.Printf("build: %s", app.File)
	outputPath := gfile.Join(gfile.TempDir(), "gf-cli", gfile.Name(app.File))
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	result, err := gproc.ShellExec(fmt.Sprintf(`go build -o %s %s`, outputPath, app.File))
	if err != nil {
		mlog.Printf("build error: %s, %s", err.Error(), result)
	}
	process = gproc.NewProcess(outputPath, nil)
	process.Start()
}
