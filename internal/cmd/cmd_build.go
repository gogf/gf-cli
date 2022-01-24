package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	Build = cBuild{
		nodeNameInConfigFile: "gfcli.build",
		packedGoFileName:     "build_pack_data.go",
	}
)

type cBuild struct {
	g.Meta               `name:"build" brief:"{cBuildBrief}" dc:"{cBuildDc}" eg:"{cBuildEg}" ad:"{cBuildAd}"`
	nodeNameInConfigFile string // nodeNameInConfigFile is the node name for compiler configurations in configuration file.
	packedGoFileName     string // packedGoFileName specifies the file name for packing common folders into one single go file.
}

const (
	cBuildBrief = `cross-building go project for lots of platforms`
	cBuildEg    = `
gf build main.go
gf build main.go --pack public,template
gf build main.go --cgo
gf build main.go -m none 
gf build main.go -n my-app -a all -s all
gf build main.go -n my-app -a amd64,386 -s linux -p .
gf build main.go -n my-app -v 1.0 -a amd64,386 -s linux,windows,darwin -p ./docker/bin
`
	cBuildDc = `
The "build" command is most commonly used command, which is designed as a powerful wrapper for 
"go build" command for convenience cross-compiling usage. 
It provides much more features for building binary:
1. Cross-Compiling for many platforms and architectures.
2. Configuration file support for compiling.
3. Build-In Variables.
`
	cBuildAd = `
PLATFORMS
    darwin    amd64,arm64
    freebsd   386,amd64,arm
    linux     386,amd64,arm,arm64,ppc64,ppc64le,mips,mipsle,mips64,mips64le
    netbsd    386,amd64,arm
    openbsd   386,amd64,arm
    windows   386,amd64
`
	// https://golang.google.cn/doc/install/source
	cBuildPlatforms = `
darwin    amd64
darwin    arm64
ios       amd64
ios       arm64
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
android   arm
dragonfly amd64
plan9     386
plan9     amd64
solaris   amd64
`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`cBuildBrief`: cBuildBrief,
		`cBuildDc`:    cBuildDc,
		`cBuildEg`:    cBuildEg,
		`cBuildAd`:    cBuildAd,
	})
}

type cBuildInput struct {
	g.Meta  `name:"build" config:"gfcli.build"`
	File    string `name:"FILE" arg:"true"   brief:"building file path"`
	Name    string `short:"n" name:"name"    brief:"output binary name"`
	Version string `short:"v" name:"version" brief:"output binary version"`
	Arch    string `short:"a" name:"arch"    brief:"output binary architecture, multiple arch separated with ','"`
	System  string `short:"s" name:"system"  brief:"output binary system, multiple os separated with ','"`
	Output  string `short:"o" name:"output"  brief:"output binary path, used when building single binary file"`
	Path    string `short:"p" name:"path"    brief:"output binary directory path, default is './bin'" d:"./bin"`
	Extra   string `short:"e" name:"extra"   brief:"extra custom \"go build\" options"`
	Mod     string `short:"m" name:"mod"     brief:"like \"-mod\" option of \"go build\", use \"-m none\" to disable go module"`
	Cgo     bool   `short:"c" name:"cgo"     brief:"enable or disable cgo feature, it's disabled in default" orphan:"true"`
	VarMap  g.Map  `short:"r" name:"varMap"  brief:"custom built embedded variable into binary"`
	Pack    string `name:"pack" brief:"pack specified folder into temporary go file before building and removes it after built"`
}
type cBuildOutput struct{}

func (c cBuild) Index(ctx context.Context, in cBuildInput) (out *cBuildOutput, err error) {
	mlog.SetHeaderPrint(true)

	mlog.Debugf(`build input: %+v`, in)
	// Necessary check.
	if gproc.SearchBinary("go") == "" {
		mlog.Fatalf(`command "go" not found in your environment, please install golang first to proceed this command`)
	}

	var (
		parser = gcmd.ParserFromCtx(ctx)
		file   = parser.GetArg(2).String()
	)
	if len(file) < 1 {
		// Check and use the main.go file.
		if gfile.Exists("main.go") {
			file = "main.go"
		} else {
			mlog.Fatal("build file path cannot be empty")
		}
	}
	if in.Name == "" {
		in.Name = gfile.Name(file)
	}
	if len(in.Name) < 1 || in.Name == "*" {
		mlog.Fatal("name cannot be empty")
	}
	if in.Mod != "" && in.Mod != "none" {
		mlog.Debugf(`mod is %s`, in.Mod)
		if in.Extra == "" {
			in.Extra = fmt.Sprintf(`-mod=%s`, in.Mod)
		} else {
			in.Extra = fmt.Sprintf(`-mod=%s %s`, in.Mod, in.Extra)
		}
	}
	if in.Extra != "" {
		in.Extra += " "
	}
	var (
		customSystems = gstr.SplitAndTrim(in.System, ",")
		customArches  = gstr.SplitAndTrim(in.Arch, ",")
	)
	if len(in.Version) > 0 {
		in.Path += "/" + in.Version
	}
	// System and arch checks.
	var (
		spaceRegex  = regexp.MustCompile(`\s+`)
		platformMap = make(map[string]map[string]bool)
	)
	for _, line := range strings.Split(strings.TrimSpace(cBuildPlatforms), "\n") {
		line = gstr.Trim(line)
		line = spaceRegex.ReplaceAllString(line, " ")
		var (
			array  = strings.Split(line, " ")
			system = strings.TrimSpace(array[0])
			arch   = strings.TrimSpace(array[1])
		)
		if platformMap[system] == nil {
			platformMap[system] = make(map[string]bool)
		}
		platformMap[system][arch] = true
	}
	// Auto packing.
	if len(in.Pack) > 0 {
		dataFilePath := fmt.Sprintf(`packed/%s`, c.packedGoFileName)
		if !gfile.Exists(dataFilePath) {
			// Remove the go file that is automatically packed resource.
			defer func() {
				_ = gfile.Remove(dataFilePath)
				mlog.Printf(`remove the automatically generated resource go file: %s`, dataFilePath)
			}()
		}
		packCmd := fmt.Sprintf(`gf pack %s %s`, in.Pack, dataFilePath)
		mlog.Print(packCmd)
		gproc.MustShellRun(packCmd)
	}

	// Injected information by building flags.
	ldFlags := fmt.Sprintf(`-X 'github.com/gogf/gf/v2/os/gbuild.builtInVarStr=%v'`, c.getBuildInVarStr(in))

	// start building
	mlog.Print("start building...")
	if in.Cgo {
		genv.MustSet("CGO_ENABLED", "1")
	} else {
		genv.MustSet("CGO_ENABLED", "0")
	}
	var (
		cmd = ""
		ext = ""
	)
	for system, item := range platformMap {
		cmd = ""
		ext = ""
		if len(customSystems) > 0 && customSystems[0] != "all" && !gstr.InArray(customSystems, system) {
			continue
		}
		for arch, _ := range item {
			if len(customArches) > 0 && customArches[0] != "all" && !gstr.InArray(customArches, arch) {
				continue
			}
			if len(customSystems) == 0 && len(customArches) == 0 {
				if runtime.GOOS == "windows" {
					ext = ".exe"
				}
				// Single binary building, output the binary to current working folder.
				output := ""
				if len(in.Output) > 0 {
					output = "-o " + in.Output + ext
				} else {
					output = "-o " + in.Name + ext
				}
				cmd = fmt.Sprintf(`go build %s -ldflags "%s" %s %s`, output, ldFlags, in.Extra, file)
			} else {
				// Cross-building, output the compiled binary to specified path.
				if system == "windows" {
					ext = ".exe"
				}
				genv.MustSet("GOOS", system)
				genv.MustSet("GOARCH", arch)
				cmd = fmt.Sprintf(
					`go build -o %s/%s/%s%s -ldflags "%s" %s%s`,
					in.Path, system+"_"+arch, in.Name, ext, ldFlags, in.Extra, file,
				)
			}
			mlog.Debug(cmd)
			// It's not necessary printing the complete command string.
			cmdShow, _ := gregex.ReplaceString(`\s+(-ldflags ".+?")\s+`, " ", cmd)
			mlog.Print(cmdShow)
			if result, err := gproc.ShellExec(cmd); err != nil {
				mlog.Printf("failed to build, os:%s, arch:%s, error:\n%s\n", system, arch, gstr.Trim(result))
			} else {
				mlog.Debug(gstr.Trim(result))
			}
			// single binary building.
			if len(customSystems) == 0 && len(customArches) == 0 {
				goto buildDone
			}
		}
	}
buildDone:
	mlog.Print("done!")
	return
}

// getBuildInVarMapJson retrieves and returns the custom build-in variables in configuration
// file as json.
func (c cBuild) getBuildInVarStr(in cBuildInput) string {
	buildInVarMap := in.VarMap
	if buildInVarMap == nil {
		buildInVarMap = make(g.Map)
	}
	buildInVarMap["builtGit"] = c.getGitCommit()
	buildInVarMap["builtTime"] = gtime.Now().String()
	b, err := json.Marshal(buildInVarMap)
	if err != nil {
		mlog.Fatal(err)
	}
	return gbase64.EncodeToString(b)
}

// getGitCommit retrieves and returns the latest git commit hash string if present.
func (c cBuild) getGitCommit() string {
	if gproc.SearchBinary("git") == "" {
		return ""
	}
	var (
		cmd  = `git log -1 --format="%cd %H" --date=format:"%Y-%m-%d %H:%M:%S"`
		s, _ = gproc.ShellExec(cmd)
	)
	mlog.Debug(cmd)
	if s != "" {
		if !gstr.Contains(s, "fatal") {
			return gstr.Trim(s)
		}
	}
	return ""
}
