package boot

import (
	"github.com/gogf/gf/frame/g"
)

func init() {
	g.Config("url").SetFileName("url.toml")
}
