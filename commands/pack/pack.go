package pack

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gres"
	"github.com/gogf/gf/text/gstr"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf pack SRC DST

ARGUMENT
    SRC  source path for packing, which can be multiple source paths.
    DST  destination file path for packed file.
         if extension of the filename is ".go" and "-n" option is given, 
         it enables packing SRC to go file, or else it packs SRC into a binary file.

OPTION
    -n, --name      package name for output go file
    -p, --prefix    prefix for each file packed into the resource file

EXAMPLES
    gf pack public data.bin
    gf pack public,template data.bin
    gf pack public,template data/data.go -n=data
    gf pack public,template,config resource/resource.go -n=resource
    gf pack public,template,config resource/resource.go -n=resource -p=/var/www/my-app
    gf pack /var/www/public resource/resource.go -n=resource
`))
}

func Run() {
	srcPath := gcmd.Value.Get(2)
	dstPath := gcmd.Value.Get(3)
	if srcPath == "" {
		mlog.Fatal("SRC path cannot be empty")
	}
	if dstPath == "" {
		mlog.Fatal("DST path cannot be empty")
	}
	name := gcmd.Option.Get("name", gcmd.Option.Get("n"))
	prefix := gcmd.Option.Get("prefix", gcmd.Option.Get("p"))

	mlog.Print("packing...")
	if name != "" {
		if err := gres.PackToGoFile(srcPath, dstPath, name, prefix); err != nil {
			mlog.Fatalf("pack failed: %v", err)
		}
	} else {
		if err := gres.PackToFile(srcPath, dstPath, prefix); err != nil {
			mlog.Fatalf("pack failed: %v", err)
		}
	}
	mlog.Print("done!")
}
