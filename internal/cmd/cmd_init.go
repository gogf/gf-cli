package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf-cli/v2/utility/allyes"
	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gres"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	Init = commandInit{}
)

type commandInit struct {
	g.Meta `name:"init" brief:"{commandInitBrief}" eg:"{commandInitEg}"`
}

const (
	commandInitMonoRepo   = `template-mono`
	commandInitSingleRepo = `template-single`
	commandInitBrief      = `create and initialize an empty GoFrame project`
	commandInitEg         = `
gf init my-project
gf init my-mono-repo -m
`
	commandInitNameBrief = `
name for the project. It will create a folder with NAME in current directory.
The NAME will also be the module name for the project.
`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandInitBrief`:     commandInitBrief,
		`commandInitEg`:        commandInitEg,
		`commandInitNameBrief`: commandInitNameBrief,
	})
}

type commandInitInput struct {
	g.Meta `name:"init"`
	Name   string `name:"NAME" arg:"true" v:"required" brief:"{commandInitNameBrief}"`
	Mono   bool   `name:"mono" short:"m" brief:"initialize a mono-repo instead a single-repo" orphan:"true"`
}
type commandInitOutput struct{}

func (c commandInit) Index(ctx context.Context, in commandInitInput) (out *commandInitOutput, err error) {
	if !gfile.IsEmpty(in.Name) && !allyes.Check() {
		s := gcmd.Scanf(`the folder "%s" is not empty, files might be overwrote, continue? [y/n]: `, in.Name)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	mlog.Print("initializing...")
	if in.Mono {
		err = gres.Export(commandInitMonoRepo, in.Name, gres.ExportOption{
			RemovePrefix: commandInitMonoRepo,
		})
	} else {
		err = gres.Export(commandInitSingleRepo, in.Name, gres.ExportOption{
			RemovePrefix: commandInitSingleRepo,
		})
	}
	if err != nil {
		return
	}
	mlog.Print("initialization done! ")
	if !in.Mono {
		mlog.Print("you can now run 'gf run main.go' to start your journey, enjoy!")
	}
	return
}
