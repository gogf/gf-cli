package run

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
	"time"
)

type App struct {
	File      string
	WatchList []string
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
		File:      file,
		WatchList: []string{"(.go)$"},
	}
}

func Run() {
	file := gcmd.GetArg(2)
	if len(file) < 1 {
		mlog.Fatal("file path cannot be empty")
	}
	app := New(file)
	dirty := gtype.NewBool()
	_, err := gfsnotify.Add(app.File, func(event *gfsnotify.Event) {
		g.Log().Debug(event)
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		dirty.Set(true)
		// With some delay in case of multiple code changes in very short interval.
		gtimer.SetTimeout(time.Second, func() {
			// Variable <dirty> is used for running the changes only one in one second.
			if dirty.Cas(true, false) {
				app.Run()
			}
		})

	})
	if err != nil {
		mlog.Fatal(err)
	}
	app.Run()
	select {}
}

func (app *App) Run() {
	// TODO Check the codes using 'go build'.

	// Kill the old process.
	if process != nil {
		process.Kill()
	}
	// Wait until the old process exits.
	time.Sleep(time.Second)
	// Running the codes.
	mlog.Printf("Build: %s", app.File)
	newProcess := gproc.NewProcessCmd("go run " + app.File)
	if err := newProcess.Run(); err != nil {
		mlog.Print("Build failed:", err)
	}
	process = newProcess
}
