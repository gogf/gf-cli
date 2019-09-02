package get

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"math"
	"sync"
	"time"
)

const (
	gPROXY_CHECK_TIMEOUT = time.Second
)

var (
	httpClient = ghttp.NewClient()
)

func init() {
	httpClient.SetTimeOut(gPROXY_CHECK_TIMEOUT)
}

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE    
    gf get [ARGUMENT]

ARGUMENT 
    [PACKAGE]  remote golang package path, eg: github.com/gogf/gf
               it's optional, it updates GF version in default
`))
}

func Run() {
	genv.Set("GOPROXY", getProxy())
	mlog.Print("cleaning cache...")
	gproc.ShellRun("go clean -modcache")
	if value := gcmd.Value.Get(2); value != "" {
		options := gcmd.Option.Build("-")
		if options == "" {
			options = "-u"
		}
		gproc.ShellRun(fmt.Sprintf(`go get %s %s`, options, value))
	} else {
		mlog.Print("downloading the latest version of GF...")
		gproc.ShellRun("go get -u github.com/gogf/gf")
	}
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
	start := gtime.Millisecond()
	r, err := httpClient.Head(url)
	if err != nil || r.StatusCode != 200 {
		return math.MaxInt32
	}
	defer r.Close()
	return int(gtime.Millisecond() - start)
}
