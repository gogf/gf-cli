package pack

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/text/gstr"
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf pack SRC DST

ARGUMENT
    SRC  source path for packing
    DST  destination file path for packed file,
         if extension of the filename is '.go', it then outputs to a go file

OPTION
    -n, --name      package name for output go file
    -p, --prefix    prefix for each file packed into the resource file

EXAMPLES
    gf pack ./public ./data.bin
    gf pack ./public ./data/data.go -n=data
    gf pack ./public ./resource/resource.go -n=resource -p=/var/www/public
`))
}

func Run() {

}
