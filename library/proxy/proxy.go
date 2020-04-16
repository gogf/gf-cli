package proxy

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gtime"
	"math"
	"sync"
	"time"
)

var (
	httpClient = ghttp.NewClient()
)

func init() {
	httpClient.SetTimeout(time.Second)
}

// AutoSet automatically checks and sets the golang proxy.
func AutoSet() {
	SetGoModuleEnabled(true)
	genv.Set("GOPROXY", "https://goproxy.cn")
}

// SetGoModuleEnabled enables/disables the go module feature.
func SetGoModuleEnabled(enabled bool) {
	if enabled {
		mlog.Debug("set GO111MODULE=on")
		genv.Set("GO111MODULE", "on")
	} else {
		mlog.Debug("set GO111MODULE=off")
		genv.Set("GO111MODULE", "off")
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
	start := gtime.TimestampMilli()
	r, err := httpClient.Head(url)
	if err != nil || r.StatusCode != 200 {
		return math.MaxInt32
	}
	defer r.Close()
	return int(gtime.TimestampMilli() - start)
}
