package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Gen = commandGen{}
)

type commandGen struct {
	g.Meta `name:"gen" brief:"{commandGenUsage}" brief:"{commandGenBrief}" eg:"{commandGenEg}" eg:"{commandGenDc}"`
}
