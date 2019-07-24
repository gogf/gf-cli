package compile

import (
	"fmt"
	"github.com/gogf/gf/g/os/gcmd"
	"github.com/gogf/gf/g/os/genv"
	"github.com/gogf/gf/g/os/gfile"
	"github.com/gogf/gf/g/os/gproc"
	"github.com/gogf/gf/g/text/gstr"
	"os"
	"regexp"
	"strings"
)

const platforms = `
    android   arm
    darwin    386
    darwin    amd64
    darwin    arm
    darwin    arm64
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
    plan9     386
    plan9     amd64
    solaris   amd64
    windows   386
    windows   amd64
`

func Run() {
	file := gcmd.Value.Get(2)
	if len(file) < 1 {
		fmt.Fprintln(os.Stderr, "ERROR: file path cannot be empty.")
		os.Exit(1)
	}
	name := gcmd.Option.Get("name", gcmd.Option.Get("n", gfile.Name(file)))
	if len(name) < 1 || name == "*" {
		fmt.Println("ERROR: name cannot be empty")
		return
	}
	version := gcmd.Option.Get("version", gcmd.Option.Get("v"))
	arches := strings.Split(gcmd.Option.Get("arch", gcmd.Option.Get("a")), ",")
	oses := strings.Split(gcmd.Option.Get("os"), ",")
	ext := ""
	cmd := ""
	path := "./bin"
	if len(version) > 0 {
		path += "/" + version
	}
	reg := regexp.MustCompile(`\s+`)
	lines := strings.Split(strings.TrimSpace(platforms), "\n")
	fmt.Println("compiling...")
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
		if _, err := gproc.ShellExec(cmd); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: build failed:", cmd)
			return
		} else {
			fmt.Fprintln(os.Stdout, cmd)
		}
	}
}
