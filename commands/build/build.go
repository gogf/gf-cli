package build

import (
	"encoding/json"
	"fmt"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/encoding/gbase64"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gutil"
	"regexp"
	"strings"
)

// https://golang.google.cn/doc/install/source
// Here're the most commonly used platforms and arches,
// but some are removed:
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

// nodeNameInConfigFile is the node name for compiler configurations in configuration file.
const nodeNameInConfigFile = "compiler"

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
    -s, --system     output binary system, multiple os separated with ','
    -o, --output     output binary path, used when building single binary file
    -p, --path       output binary directory path, default is './bin'
	-e, --extra      extra custom 'go build' options

EXAMPLES
    gf build main.go
    gf build main.go -n my-app -a all -s all
    gf build main.go -n my-app -a amd64,386 -s linux -p .
    gf build main.go -n my-app -v 1.0 -a amd64,386 -s linux,windows,darwin -p ./dockerfiles/bin

DESCRIPTION
    The "build" command is most commonly used command, which is designed as a powerful wrapper for 
    "go build" command for convenience cross-compiling usage. 
    It provides much more features for building binary:
    1. Cross-Compiling for many platforms and architectures.
    2. Configuration file support for compiling.
    3. Build-In Variables.

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
		"s,system":  true,
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
	path := getOption(parser, "path", "./bin")
	name := getOption(parser, "name", gfile.Name(file))
	if len(name) < 1 || name == "*" {
		mlog.Fatal("name cannot be empty")
	}
	extra := getOption(parser, "extra")
	version := getOption(parser, "version")
	outputPath := getOption(parser, "output")
	archOption := getOption(parser, "arch")
	systemOption := getOption(parser, "system")
	arches := strings.Split(archOption, ",")
	systems := strings.Split(systemOption, ",")
	if len(version) > 0 {
		path += "/" + version
	}

	// injected information.
	ldFlags := fmt.Sprintf(`-X 'github.com/gogf/gf/os/gbuild.builtInVarStr=%v'`, getBuildInVarStr())

	// start building
	mlog.Print("start building...")
	genv.Set("CGO_ENABLED", "0")
	cmd := ""
	reg := regexp.MustCompile(`\s+`)
	lines := strings.Split(strings.TrimSpace(platforms), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = reg.ReplaceAllString(line, " ")
		array := strings.Split(line, " ")
		array[0] = strings.TrimSpace(array[0])
		array[1] = strings.TrimSpace(array[1])
		if len(systems) > 0 && systems[0] != "" && systems[0] != "all" && !gstr.InArray(systems, array[0]) {
			continue
		}
		if len(arches) > 0 && arches[0] != "" && arches[0] != "all" && !gstr.InArray(arches, array[1]) {
			continue
		}
		if len(systemOption) == 0 && len(archOption) == 0 {
			// Single binary building, output the binary to current working folder.
			output := ""
			if len(outputPath) > 0 {
				output = "-o " + outputPath
			} else {
				output = "-o " + name
			}
			cmd = fmt.Sprintf(`go build %s -ldflags "%s" %s %s`, output, ldFlags, extra, file)
		} else {
			// Cross-building, output the compiled binary to specified path.
			if array[0] == "windows" {
				name += ".exe"
			}
			genv.Set("GOOS", array[0])
			genv.Set("GOARCH", array[1])
			cmd = fmt.Sprintf(
				`go build -o %s/%s/%s -ldflags "%s" %s %s`,
				path, array[0]+"_"+array[1], name, ldFlags, extra, file,
			)
		}
		// It's not necessary printing the complete command string.
		//cmdShow, _ := gregex.ReplaceString(`\s+(-ldflags ".+?")\s+`, " ", cmd)
		mlog.Print(cmd)
		if _, err := gproc.ShellExec(cmd); err != nil {
			mlog.Fatal("build failed:", cmd)
		}
		// single binary building.
		if len(systemOption) == 0 && len(archOption) == 0 {
			break
		}
	}
}

// getOption retrieves option value from parser and configuration file.
// It returns the default value specified by parameter <value> is no value found.
func getOption(parser *gcmd.Parser, name string, value ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" {
		result = g.Config().GetString(nodeNameInConfigFile + "." + name)
	}
	if result == "" && len(value) > 0 {
		result = value[0]
	}
	return
}

// getBuildInVarMapJson retrieves and returns the custom build-in variables in configuration
// file as json.
func getBuildInVarStr() string {
	buildInVarMap := g.Map{}
	configMap := g.Config().GetMap(nodeNameInConfigFile)
	if len(configMap) > 0 {
		_, v := gutil.MapPossibleItemByKey(configMap, "VarMap")
		if v != nil {
			buildInVarMap = gconv.Map(v)
		}
	}
	buildInVarMap["builtGit"] = getGitCommit()
	buildInVarMap["builtTime"] = gtime.Now().String()
	b, err := json.Marshal(buildInVarMap)
	if err != nil {
		mlog.Fatal(err)
	}
	return gbase64.EncodeToString(b)
}

// getGitCommit retrieves and returns the latest git commit hash string if present.
func getGitCommit() string {
	if gproc.SearchBinary("git") == "" {
		return ""
	}
	if s, _ := gproc.ShellExec("git rev-list -1 HEAD"); s != "" {
		if !gstr.Contains(s, " ") && !gstr.Contains(s, "fatal") {
			return gstr.Trim(s)
		}
	}
	return ""
}
