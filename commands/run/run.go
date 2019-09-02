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

// App app
type App struct {
	Name        string
	Path        string
	FullPath    string
	BuildTags   string
	CMD         *exec.Cmd
	Locker      *gmlock.Locker
	Timer       *gtimer.Entry
	WatchList   []string
	UnWatchList []string
}

// Help how to use
func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf run
EXAMPLES
    gf run
`))
}

// New new app
func New() *App {
	return &App{
		Locker:      gmlock.New(),
		WatchList:   []string{"(.go)$"},
		UnWatchList: []string{"(.js|.html|.bat|.txt|.md|.exe|.exe~)$"},
	}
}

// Run run
func Run() {
	app := New()

	// 获取当前目录
	app.Path, _ = os.Getwd()

	app.Name = filepath.Base(app.Path)
	// 监控目录
	_, err := gfsnotify.Add(app.Path, func(event *gfsnotify.Event) {
		// 排除文件
		if app.IsUnwatch(event.Path) {
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

		// 非目标文件不重新编译
		if !app.IsWatch(event.Path) {
			return
		}

		if app.Timer != nil {
			app.Timer.Close()
			app.Timer = nil
		}
		// 使用延时执行，避免短时间内多次文件变动导致异常
		app.Timer = gtimer.AddOnce(time.Second, func() {
			app.Build()
		})
	}, true)

	if err != nil {
		mlog.Fatal("%v", err)
	} else {
		app.Build()
		select {}
	}
}

// IsUnwatch is file unwatch or not
func (app *App) IsUnwatch(filename string) bool {
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

// IsWatch is file watch or not
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

// Build build the app
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
	//buildCmd.Env = append(os.Environ(), "GO111MODULE=auto")
	buildCmd.Stderr = &stderr
	err = buildCmd.Run()
	if err != nil {
		mlog.Print("Build Failed:", stderr.String())
		return
	}
	app.Restart()
}

// Kill kill the app
func (app *App) Kill() {
	defer func() {
		if e := recover(); e != nil {
		}
	}()
	if app.CMD != nil && app.CMD.Process != nil {
		err := app.CMD.Process.Kill()
		if err != nil {
			mlog.Fatal(err)
		}
	}
}

// Restart restart the app
func (app *App) Restart() {
	app.Kill()
	go app.Start()
}

// Start start the app
func (app *App) Start() {
	appname := app.Name
	if !strings.Contains(appname, "./") {
		appname = "./" + appname
	}

	app.CMD = exec.Command(appname)
	app.CMD.Stdout = os.Stdout
	app.CMD.Stderr = os.Stderr
	//cmd.Args = append([]string{appname}, config.Conf.CmdArgs...)
	//cmd.Env = append(os.Environ(), config.Conf.Envs...)

	go app.CMD.Run()
}
