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
	Pack = commandPack{}
)

type commandPack struct {
	g.Meta `name:"pack" brief:"{commandPackUsage}" brief:"{commandPackBrief}" eg:"{commandPackEg}"`
}

const (
	commandPackUsage = `gf pack SRC DST`
	commandPackBrief = `packing any file/directory to a resource file, or a go file`
	commandPackEg    = `
gf pack public data.bin
gf pack public,template data.bin
gf pack public,template packed/data.go
gf pack public,template,config packed/data.go
gf pack public,template,config packed/data.go -n=packed -p=/var/www/my-app
gf pack /var/www/public packed/data.go -n=packed
`
	commandPackSrcBrief = `source path for packing, which can be multiple source paths.`
	commandPackDstBrief = `
destination file path for packed file. if extension of the filename is ".go" and "-n" option is given, 
it enables packing SRC to go file, or else it packs SRC into a binary file.
`
	commandPackNameBrief   = `package name for output go file, it's set as its directory name if no name passed`
	commandPackPrefixBrief = `prefix for each file packed into the resource file`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandPackUsage`:       commandPackUsage,
		`commandPackBrief`:       commandPackBrief,
		`commandPackEg`:          commandPackEg,
		`commandPackSrcBrief`:    commandPackSrcBrief,
		`commandPackDstBrief`:    commandPackDstBrief,
		`commandPackNameBrief`:   commandPackNameBrief,
		`commandPackPrefixBrief`: commandPackPrefixBrief,
	})
}

type commandPackInput struct {
	g.Meta `name:"pack"`
	Src    string `name:"SRC" arg:"true" v:"required" brief:"{commandPackSrcBrief}"`
	Dst    string `name:"DST" arg:"true" v:"required" brief:"{commandPackDstBrief}"`
	Name   string `name:"name"   short:"n" brief:"{commandPackNameBrief}"`
	Prefix string `name:"prefix" short:"p" brief:"{commandPackPrefixBrief}"`
}
type commandPackOutput struct{}

func (c commandPack) Index(ctx context.Context, in commandPackInput) (out *commandPackOutput, err error) {
	if gfile.Exists(in.Dst) && gfile.IsDir(in.Dst) {
		mlog.Fatalf("DST path '%s' cannot be a directory", in.Dst)
	}
	if !gfile.IsEmpty(in.Dst) && !allyes.Check() {
		s := gcmd.Scanf("path '%s' is not empty, files might be overwrote, continue? [y/n]: ", in.Dst)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	if in.Name == "" && gfile.ExtName(in.Dst) == "go" {
		in.Name = gfile.Basename(gfile.Dir(in.Dst))
	}
	if in.Name != "" {
		if err = gres.PackToGoFile(in.Src, in.Dst, in.Name, in.Prefix); err != nil {
			mlog.Fatalf("pack failed: %v", err)
		}
	} else {
		if err = gres.PackToFile(in.Src, in.Dst, in.Prefix); err != nil {
			mlog.Fatalf("pack failed: %v", err)
		}
	}
	mlog.Print("done!")
	return
}
