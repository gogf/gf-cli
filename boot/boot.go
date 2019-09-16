package boot

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcfg"
)

var (
	urlConfig = `
[cdn]
    url = "https://gf.cdn.johng.cn"

[home]
    url = "https://goframe.org"

[proxy]
    urls = [
		"https://mirrors.aliyun.com/goproxy/", 
		"https://goproxy.io/", 
		"https://goproxy.cn/"
	]
`
)

// DO NOT overwrites the default configuration!
func init() {
	gcfg.SetContent(urlConfig, "url.toml")
	g.Config("url").SetFileName("url.toml")
}
