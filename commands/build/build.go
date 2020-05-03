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
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gutil"
	"regexp"
	"runtime"
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
    -e, --extra      extra custom "go build" options
    -m, --mod        like "-mod" option of "go build", use "-m none" to disable go module
    --swagger        auto parse and pack swagger into boot/data-swagger.go before building. 
    --pack           auto pack config,public,template folder into boot/data-packed.go before building.

EXAMPLES
    gf build main.go
    gf build main.go --swagger
    gf build main.go --pack
    gf build main.go -m none --pack
    gf build main.go -n my-app -a all -s all
    gf build main.go -n my-app -a amd64,386 -s linux -p .
    gf build main.go -n my-app -v 1.0 -a amd64,386 -s linux,windows,darwin -p ./docker/bin

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
	mlog.SetHeaderPrint(true)
	parser, err := gcmd.Parse(g.MapStrBool{
		"n,name":    true,
		"v,version": true,
		"a,arch":    true,
		"s,system":  true,
		"o,output":  true,
		"p,path":    true,
		"e,extra":   true,
		"m,mod":     true,
		"swagger":   false,
		"pack":      false,
	})
	if err != nil {
		mlog.Fatal(err)
	}
	file := parser.GetArg(2)
	if len(file) < 1 {
		// Check and use the main.go file.
		if gfile.Exists("main.go") {
			file = "main.go"
		} else {
			mlog.Fatal("build file path cannot be empty")
		}
	}
	path := getOption(parser, "path", "./bin")
	name := getOption(parser, "name", gfile.Name(file))
	if len(name) < 1 || name == "*" {
		mlog.Fatal("name cannot be empty")
	}
	var (
		mod   = getOption(parser, "mod")
		extra = getOption(parser, "extra")
	)
	if mod != "" && mod != "none" {
		mlog.Debugf(`mod is %s`, mod)
		if extra == "" {
			extra = fmt.Sprintf(`-mod=%s`, mod)
		} else {
			extra = fmt.Sprintf(`-mod=%s %s`, mod, extra)
		}
	}
	var (
		version      = getOption(parser, "version")
		outputPath   = getOption(parser, "output")
		archOption   = getOption(parser, "arch")
		systemOption = getOption(parser, "system")
		arches       = strings.Split(archOption, ",")
		systems      = strings.Split(systemOption, ",")
	)
	if len(version) > 0 {
		path += "/" + version
	}

	// Auto swagger.
	if containsOption(parser, "swagger") {
		if err := gproc.ShellRun(`gf swagger`); err != nil {
			return
		}
		if gfile.Exists("swagger") {
			packCmd := fmt.Sprintf(`gf pack %s boot/data-swagger.go -n boot`, "swagger")
			mlog.Print(packCmd)
			if err := gproc.ShellRun(packCmd); err != nil {
				return
			}
		}
	}

	// Auto packing
	if containsOption(parser, "pack") {
		packFolderStr := ""
		if gfile.Exists("config") {
			packFolderStr += "config,"
		}
		if gfile.Exists("public") {
			packFolderStr += "public,"
		}
		if gfile.Exists("template") {
			packFolderStr += "template,"
		}
		packFolderStr = gstr.Trim(packFolderStr, ",")
		if len(packFolderStr) > 0 {
			packCmd := fmt.Sprintf(`gf pack %s boot/data-packed.go -n boot`, packFolderStr)
			mlog.Print(packCmd)
			gproc.ShellRun(packCmd)
		}
	}

	// injected information.
	ldFlags := fmt.Sprintf(`-X 'github.com/gogf/gf/os/gbuild.builtInVarStr=%v'`, getBuildInVarStr())

	// start building
	mlog.Print("start building...")
	genv.Set("CGO_ENABLED", "0")
	var (
		cmd   = ""
		ext   = ""
		reg   = regexp.MustCompile(`\s+`)
		lines = strings.Split(strings.TrimSpace(platforms), "\n")
	)
	for _, line := range lines {
		cmd = ""
		ext = ""
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
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}
			// Single binary building, output the binary to current working folder.
			output := ""
			if len(outputPath) > 0 {
				output = "-o " + outputPath + ext
			} else {
				output = "-o " + name + ext
			}
			cmd = fmt.Sprintf(`go build %s -ldflags "%s" %s %s`, output, ldFlags, extra, file)
		} else {
			// Cross-building, output the compiled binary to specified path.
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
		// It's not necessary printing the complete command string.
		cmdShow, _ := gregex.ReplaceString(`\s+(-ldflags ".+?")\s+`, " ", cmd)
		mlog.Print(cmdShow)
		if result, err := gproc.ShellExec(cmd); err != nil {
			mlog.Fatalf("build failed: %s%s", result, err.Error())
		}
		// single binary building.
		if len(systemOption) == 0 && len(archOption) == 0 {
			break
		}
	}
	mlog.Print("done!")
}

// getOption retrieves option value from parser and configuration file.
// It returns the default value specified by parameter <value> is no value found.
func getOption(parser *gcmd.Parser, name string, value ...string) (result string) {
	result = parser.GetOpt(name)
	if result == "" && g.Config().Available() {
		result = g.Config().GetString(nodeNameInConfigFile + "." + name)
	}
	if result == "" && len(value) > 0 {
		result = value[0]
	}
	return
}

// containsOption checks whether the command option or the configuration file containing
// given option name.
func containsOption(parser *gcmd.Parser, name string) bool {
	result := parser.ContainsOpt(name)
	if !result && g.Config().Available() {
		result = g.Config().Contains(nodeNameInConfigFile + "." + name)
	}
	return result
}

// getBuildInVarMapJson retrieves and returns the custom build-in variables in configuration
// file as json.
func getBuildInVarStr() string {
	buildInVarMap := g.Map{}
	if g.Config().Available() {
		configMap := g.Config().GetMap(nodeNameInConfigFile)
		if len(configMap) > 0 {
			_, v := gutil.MapPossibleItemByKey(configMap, "VarMap")
			if v != nil {
				buildInVarMap = gconv.Map(v)
			}
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
