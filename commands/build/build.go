package build

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"regexp"
	"strings"
)

// https://golang.google.cn/doc/install/source
// Here're the most common used platforms and arches,
// Removed:
//    android	arm
//    darwin	arm
//    darwin	arm64
//    plan9     386
//    plan9     amd64
const platforms = `
	darwin    386
	darwin    amd64
	dragonfly amd64
	freebsd   386
	freebsd   amd64
	freebsd   arm
	linux     386
	linux     amd64
	linux     arm
	linux     arm64
	linux     ppc64
	linux     ppc64le
	linux     mips
	linux     mipsle
	linux     mips64
	linux     mips64le
	netbsd    386
	netbsd    amd64
	netbsd    arm
	openbsd   386
	openbsd   amd64
	openbsd   arm
    solaris   amd64
    windows   386
    windows   amd64
`

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf build [OPTION] FILE

ARGUMENT
    FILE  building file path.

OPTION
    -n, --name       output binary name
    -v, --version    output binary version
    -a, --arch       output binary architecture, multiple arch separated with ','
    -o, --os         output binary system, multiple os separated with ','
    -p, --path       output binary directory path, default is './bin'

PLATFORMS
	darwin    386
	darwin    amd64
	dragonfly amd64
	freebsd   386
	freebsd   amd64
	freebsd   arm
	linux     386
	linux     amd64
	linux     arm
	linux     arm64
	linux     ppc64
	linux     ppc64le
	linux     mips
	linux     mipsle
	linux     mips64
	linux     mips64le
	netbsd    386
	netbsd    amd64
	netbsd    arm
	openbsd   386
	openbsd   amd64
	openbsd   arm
    solaris   amd64
    windows   386
    windows   amd64
`))
}

func Run() {
	file := gcmd.Value.Get(2)
	if len(file) < 1 {
		mlog.Fatal("file path cannot be empty")
	}
	path := gcmd.Option.Get("path", gcmd.Option.Get("p", "./bin"))
	name := gcmd.Option.Get("name", gcmd.Option.Get("n", gfile.Name(file)))
	if len(name) < 1 || name == "*" {
		mlog.Fatal("name cannot be empty")
	}
	version := gcmd.Option.Get("version", gcmd.Option.Get("v"))
	arches := strings.Split(gcmd.Option.Get("arch", gcmd.Option.Get("a")), ",")
	oses := strings.Split(gcmd.Option.Get("os", gcmd.Option.Get("o")), ",")
	ext := ""
	cmd := ""
	if len(version) > 0 {
		path += "/" + version
	}
	reg := regexp.MustCompile(`\s+`)
	lines := strings.Split(strings.TrimSpace(platforms), "\n")
	mlog.Print("building...")
	genv.Set("CGO_ENABLED", "0")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = reg.ReplaceAllString(line, " ")
		array := strings.Split(line, " ")
		array[0] = strings.TrimSpace(array[0])
		array[1] = strings.TrimSpace(array[1])
		if len(oses) > 0 && oses[0] != "" && !gstr.InArray(oses, array[0]) {
			continue
		}
		if len(arches) > 0 && arches[0] != "" && !gstr.InArray(arches, array[1]) {
			continue
		}
		ext = ""
		if array[0] == "windows" {
			ext = ".exe"
		}
		genv.Set("GOOS", array[0])
		genv.Set("GOARCH", array[1])
		cmd = fmt.Sprintf("go build -o %s/%s/%s%s %s", path, array[0]+"_"+array[1], name, ext, file)
		mlog.Print(cmd)
		_, err := gproc.ShellExec(cmd)
		if err != nil {
			mlog.Fatal("build failed:", cmd)
		}
	}
}
