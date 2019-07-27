package boot

import (
	"github.com/gogf/gf/g/os/gcfg"
)

var (
	config = `
[cdn]
    url = "https://gf.cdn.johng.cn"

[home]
    url = "https://goframe.org"

[proxy]
    urls = [
		"https://mirrors.aliyun.com/goproxy/", 
		"https://goproxy.io/", 
		"https://goproxy.cn"
	]

`
)

func init() {
	gcfg.SetContent(config)
}
