package run

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gmlock"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf run
`))
}

type App struct {
	Name        string
	Path        string
	FullPath    string
	BuildTags   string
	debugCmd    *exec.Cmd
	isModule    bool
	Locker      *gmlock.Locker
	Timer       *gtimer.Entry
	WatchList   []string
	UnWatchList []string
}

// New
func New() *App {
	return &App{
		Locker:      gmlock.New(),
		WatchList:   []string{"(.go)$"},
		UnWatchList: []string{"(.js|.html|.bat|.txt|.md|.exe)$"},
	}
}

// Run description
func Run() {
	path, _ := os.Getwd()
	app := New()

	app.Name = filepath.Base(path)

	_, err := gfsnotify.Add(path, func(event *gfsnotify.Event) {
		//app.isModule = isModule
		if app.IsUnWatch(event.Path) {
			return
		}

		switch {
		case event.IsCreate():
			mlog.Print("create file:", event.Path)
		case event.IsWrite():
			mlog.Print("write file:", event.Path)
		case event.IsRemove():
			mlog.Print("remove file:", event.Path)
		case event.IsRename():
			mlog.Print("rename file:", event.Path)
		case event.IsChmod():
			mlog.Print("chmod file:", event.Path)
		default:
			mlog.Print(event)
		}
		if !app.IsWatch(event.Path) {
			return
		}
		if app.Timer != nil {
			app.Timer.Stop()
			app.Timer = nil
		}
		if app.Timer == nil {
			app.Timer = gtimer.AddOnce(time.Second, func() {
				app.Build()
			})
		}

	}, true)
	app.Build()
	if err != nil {
		mlog.Fatal("%v", err)
	} else {
		select {}
	}
}

// IsUnWatch
func (app *App) IsUnWatch(filename string) bool {
	for _, regex := range app.UnWatchList {
		r, err := regexp.Compile(regex)
		if err != nil {
			return false
		}
		if r.MatchString(filename) {
			return true
		}
		continue
	}
	return false
}

// IsWatch
func (app *App) IsWatch(filename string) bool {
	for _, regex := range app.WatchList {
		r, err := regexp.Compile(regex)
		if err != nil {
			return false
		}
		if r.MatchString(filename) {
			return true
		}
		continue
	}
	return false
}

// Build
func (app *App) Build() {
	app.Locker.Lock(app.Name)
	defer app.Locker.Unlock(app.Name)

	var (
		err     error
		stderr  bytes.Buffer
		appname string
	)
	appname = app.Name
	if runtime.GOOS == "windows" {
		appname += ".exe"
	}
	cmdName := "go"
	args := []string{"build"}
	args = append(args, "-o", appname)
	if app.BuildTags != "" {
		args = append(args, "-tags", app.BuildTags)
	}
	buildCmd := exec.Command(cmdName, args...)
	buildCmd.Env = append(os.Environ(), "GOGC=off")
	if app.isModule {
		buildCmd.Env = append(os.Environ(), "GO111MODULE=auto")
	}
	buildCmd.Stderr = &stderr
	err = buildCmd.Run()
	if err != nil {
		mlog.Fatal("Build Failed:", stderr.String())
		return
	}
	app.Restart()
}

// Kill
func (app *App) Kill() {
	defer func() {
		if e := recover(); e != nil {
		}
	}()
	if app.debugCmd != nil && app.debugCmd.Process != nil {
		err := app.debugCmd.Process.Kill()
		if err != nil {
			mlog.Fatal(err)
		}
	}
}

// Restart
func (app *App) Restart() {
	app.Kill()
	go app.Start()
}

// Start
func (app *App) Start() {
	appname := app.Name
	if !strings.Contains(appname, "./") {
		appname = "./" + appname
	}

	app.debugCmd = exec.Command(appname)
	app.debugCmd.Stdout = os.Stdout
	app.debugCmd.Stderr = os.Stderr
	//cmd.Args = append([]string{appname}, config.Conf.CmdArgs...)
	//cmd.Env = append(os.Environ(), config.Conf.Envs...)

	go app.debugCmd.Run()
}
