package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/container/gtype"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gfsnotify"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	Run = commandRun{}
)

type commandRun struct {
	g.Meta `name:"run" usage:"{commandRunUsage}" brief:"{commandRunBrief}" eg:"{commandRunEg}" dc:"{commandRunDc}"`
}

type commandRunApp struct {
	File    string // Go run file name.
	Path    string // Directory storing built binary.
	Options string // Extra "go run" options.
	Args    string // Custom arguments.
}

const (
	commandRunUsage = `gf run FILE [OPTION]`
	commandRunBrief = `running go codes with hot-compiled-like feature`
	commandRunEg    = `
gf run main.go
gf run main.go --args "server -p 8080"
gf run main.go -mod=vendor
`
	commandRunDc = `
The "run" command is used for running go codes with hot-compiled-like feature,
which compiles and runs the go codes asynchronously when codes change.
`
	commandRunFileBrief  = `building file path.`
	commandRunPathBrief  = `output directory path for built binary file. it's "manifest/output" in default`
	commandRunExtraBrief = `the same options as "go run"/"go build" except some options as follows defined`
)

var (
	process *gproc.Process
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandRunUsage`:      commandRunUsage,
		`commandRunBrief`:      commandRunBrief,
		`commandRunEg`:         commandRunEg,
		`commandRunDc`:         commandRunDc,
		`commandRunFileBrief`:  commandRunFileBrief,
		`commandRunPathBrief`:  commandRunPathBrief,
		`commandRunExtraBrief`: commandRunExtraBrief,
	})
}

type (
	commandRunInput struct {
		g.Meta `name:"run"`
		File   string `name:"FILE"  arg:"true" brief:"{commandRunFileBrief}" v:"required"`
		Path   string `name:"path"  short:"p"  brief:"{commandRunPathBrief}"`
		Extra  string `name:"extra" short:"e"  brief:"{commandRunExtraBrief}"`
	}
	commandRunOutput struct{}
)

func (c commandRun) Index(ctx context.Context, in commandRunInput) (out *commandRunOutput, err error) {
	app := &commandRunApp{
		File:    in.File,
		Path:    in.Path,
		Options: in.Extra,
	}
	dirty := gtype.NewBool()
	_, err = gfsnotify.Add(gfile.RealPath("."), func(event *gfsnotify.Event) {
		if gfile.ExtName(event.Path) != "go" {
			return
		}
		// Variable `dirty` is used for running the changes only one in one second.
		if !dirty.Cas(false, true) {
			return
		}
		// With some delay in case of multiple code changes in very short interval.
		gtimer.SetTimeout(ctx, 1500*gtime.MS, func(ctx context.Context) {
			defer dirty.Set(false)
			mlog.Printf(`go file changes: %s`, event.String())
			app.Run()
		})
	})
	if err != nil {
		mlog.Fatal(err)
	}
	go app.Run()
	select {}
}

func (app *commandRunApp) Run() {
	// Rebuild and run the codes.
	renamePath := ""
	mlog.Printf("build: %s", app.File)
	outputPath := gfile.Join(app.Path, gfile.Name(app.File))
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
		if gfile.Exists(outputPath) {
			renamePath = outputPath + "~"
			if err := gfile.Rename(outputPath, renamePath); err != nil {
				mlog.Print(err)
			}
		}
	}
	// In case of `pipe: too many open files` error.
	// Build the app.
	buildCommand := fmt.Sprintf(
		`go build -o %s %s %s`,
		outputPath,
		app.Options,
		app.File,
	)
	mlog.Print(buildCommand)
	result, err := gproc.ShellExec(buildCommand)
	if err != nil {
		mlog.Printf("build error: \n%s%s", result, err.Error())
		return
	}
	// Kill the old process if build successfully.
	if process != nil {
		if err := process.Kill(); err != nil {
			mlog.Debugf("kill process error: %s", err.Error())
			//return
		}
	}
	// Run the binary file.
	runCommand := fmt.Sprintf(`%s %s`, outputPath, app.Args)
	mlog.Print(runCommand)
	if runtime.GOOS == "windows" {
		// Special handling for windows platform.
		// DO NOT USE "cmd /c" command.
		process = gproc.NewProcess(runCommand, nil)
	} else {
		process = gproc.NewProcessCmd(runCommand, nil)
	}
	if pid, err := process.Start(); err != nil {
		mlog.Printf("build running error: %s", err.Error())
	} else {
		mlog.Printf("build running pid: %d", pid)
	}
}
