package build

import (
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"regexp"
	"runtime"
	"strings"
)

// https://golang.google.cn/doc/install/source
// Here're the most common used platforms and arches.
// Here're the removed:
//    android    arm
//    dragonfly amd64
//    plan9     386
//    plan9     amd64
//    solaris   amd64
const platforms = `
    darwin    386
    darwin    amd64
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
    windows   386
    windows   amd64
`

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE 
    gf build FILE [OPTION]

ARGUMENT
    FILE  building file path.

OPTION
    -n, --name       output binary name
    -v, --version    output binary version
    -a, --arch       output binary architecture, multiple arch separated with ','
    -s, --os         output binary system, multiple os separated with ','
    -o, --output     output binary path, used when building single binary file
    -p, --path       output binary directory path, default is './bin'
	-e, --extra      extra 'go build' options

EXAMPLES
    gf build main.go
    gf build main.go -n=my-app -a=all -o=all
    gf build main.go -n=my-app -a=amd64,386 -o=linux -p=.
    gf build main.go -n=my-app -v=1.0 -a=amd64,386 -o=linux,windows,darwin -p=./bin

PLATFORMS
    darwin    386
    darwin    amd64
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
    windows   386
    windows   amd64
`))
}

func Run() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"n,name":    true,
		"v,version": true,
		"a,arch":    true,
		"s,os":      true,
		"o,output":  true,
		"p,path":    true,
		"e,extra":   true,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	file := parser.GetArg(2)
	if len(file) < 1 {
		mlog.Fatal("file path cannot be empty")
	}
	path := parser.GetOpt("path", "./bin")
	name := parser.GetOpt("name", gfile.Name(file))
	if len(name) < 1 || name == "*" {
		mlog.Fatal("name cannot be empty")
	}
	extra := parser.GetOpt("extra")
	version := parser.GetOpt("version")
	outputPath := parser.GetOpt("output")
	osOption := parser.GetOpt("os")
	archOption := parser.GetOpt("arch")
	oses := strings.Split(osOption, ",")
	arches := strings.Split(archOption, ",")
	ext := ""
	cmd := ""
	if len(version) > 0 {
		path += "/" + version
	}

	// git commit if present
	gitCommit := ""
	if s, _ := gproc.ShellExec("git rev-list -1 HEAD"); s != "" {
		if !gstr.Contains(s, " ") && !gstr.Contains(s, "fatal") {
			gitCommit = gstr.Trim(s)
		}
	}
	// injected information.
	ldFlagsMap := g.Map{
		"github.com/gogf/gf/debug/gdebug.buildTime":      gtime.Now().String(),
		"github.com/gogf/gf/debug/gdebug.buildGoVersion": runtime.Version(),
		"github.com/gogf/gf/debug/gdebug.buildGitCommit": gitCommit,
	}
	ldFlags := ""
	for k, v := range ldFlagsMap {
		if len(ldFlags) > 1 {
			ldFlags += " "
		}
		ldFlags += fmt.Sprintf(`-X '%s=%v'`, k, v)
	}
	// start building
	mlog.Print("start building...")
	genv.Set("CGO_ENABLED", "0")
	reg := regexp.MustCompile(`\s+`)
	lines := strings.Split(strings.TrimSpace(platforms), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = reg.ReplaceAllString(line, " ")
		array := strings.Split(line, " ")
		array[0] = strings.TrimSpace(array[0])
		array[1] = strings.TrimSpace(array[1])
		if len(oses) > 0 && oses[0] != "" && oses[0] != "all" && !gstr.InArray(oses, array[0]) {
			continue
		}
		if len(arches) > 0 && arches[0] != "" && arches[0] != "all" && !gstr.InArray(arches, array[1]) {
			continue
		}
		if len(osOption) == 0 && len(archOption) == 0 {
			// single binary building.
			output := ""
			if len(outputPath) > 0 {
				output = " -o " + outputPath
			}
			cmd = fmt.Sprintf(`go build%s -ldflags "%s" %s %s`, output, ldFlags, extra, file)
		} else {
			// cross-building.
			ext = ""
			if array[0] == "windows" {
				ext = ".exe"
			}
			genv.Set("GOOS", array[0])
			genv.Set("GOARCH", array[1])
			cmd = fmt.Sprintf(
				`go build -o %s/%s/%s%s -ldflags "%s" %s %s`,
				path, array[0]+"_"+array[1], name, ext, ldFlags, extra, file,
			)
		}
		cmdShow, _ := gregex.ReplaceString(`\s+(-ldflags ".+?")\s+`, " ", cmd)
		mlog.Print(cmdShow)
		if _, err := gproc.ShellExec(cmd); err != nil {
			mlog.Fatal("build failed:", cmd)
		}
		// single binary building.
		if len(osOption) == 0 && len(archOption) == 0 {
			break
		}
	}
}
