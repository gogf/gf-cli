package cmd

import (
	"context"
	"fmt"
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
	commandInitRepoPrefix = `github.com/gogf/`
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

	// Create project folder and files.
	var (
		templateRepoName string
	)
	if in.Mono {
		templateRepoName = commandInitMonoRepo
	} else {
		templateRepoName = commandInitSingleRepo
	}
	err = gres.Export(templateRepoName, in.Name, gres.ExportOption{
		RemovePrefix: templateRepoName,
	})
	if err != nil {
		return
	}

	// Replace template name to project name.
	err = gfile.ReplaceDir(
		commandInitRepoPrefix+templateRepoName,
		gfile.Basename(gfile.RealPath(in.Name)),
		in.Name,
		"*",
		true,
	)
	if err != nil {
		return
	}

	mlog.Print("initialization done! ")
	if !in.Mono {
		enjoyCommand := `gf run main.go`
		if in.Name != "." {
			enjoyCommand = fmt.Sprintf(`cd %s && %s`, in.Name, enjoyCommand)
		}
		mlog.Printf(`you can now run "%s" to start your journey, enjoy!`, enjoyCommand)
	}
	return
}
