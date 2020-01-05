package run

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/text/gstr"
	"math"
	"runtime"
	"sync"
	"time"
)

type App struct {
	File string
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
	genv.Set("GOPROXY", getProxy())
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
	err := gproc.ShellRun(fmt.Sprintf(`go build -o %s %s`, outputPath, app.File))
	if err != nil {
		mlog.Printf("build error: %s", err.Error())
	}
	process = gproc.NewProcess(outputPath, nil)
	process.Start()
}

// getProxy returns the proper proxy for 'go get'.
func getProxy() string {
	if p := genv.Get("GOPROXY"); p != "" {
		return p
	}
	wg := sync.WaitGroup{}
	checkMap := gmap.NewIntStrMap(true)
	for _, proxy := range g.Config("url").GetStrings("proxy.urls") {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			checkMap.Set(checkProxyLatency(proxy), url)
		}(proxy)
	}
	wg.Wait()

	url := ""
	latency := math.MaxInt32
	checkMap.Iterator(func(k int, v string) bool {
		if k < latency {
			url = v
			latency = k
		}
		return true
	})
	return url
}

// checkProxyLatency checks the latency for specified url.
func checkProxyLatency(url string) int {
	start := gtime.TimestampMilli()
	r, err := httpClient.Head(url)
	if err != nil || r.StatusCode != 200 {
		return math.MaxInt32
	}
	defer r.Close()
	return int(gtime.TimestampMilli() - start)
}
