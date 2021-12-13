package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	Gen = commandGen{}
)

type commandGen struct {
	g.Meta `name:"gen" brief:"{commandGenBrief}" dc:"{commandGenDc}"`
}

const (
	commandGenBrief = `automatically generate go files for dao/dto/entity/pb/pbentity`
	commandGenDc    = `
The "gen" command is designed for multiple generating purposes. 
It's currently supporting generating go files for ORM models, protobuf and protobuf entity files.
Please use "gf gen dao -h" for specified type help.
`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandGenBrief`: commandGenBrief,
		`commandGenDc`:    commandGenDc,
	})
}
