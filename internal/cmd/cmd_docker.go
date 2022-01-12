package cmd

import (
	"context"
	"fmt"

	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/util/gtag"
)

var (
	Docker = commandDocker{}
)

type commandDocker struct {
	g.Meta `name:"docker" usage:"{commandDockerUsage}" brief:"{commandDockerBrief}" eg:"{commandDockerEg}" dc:"{commandDockerDc}"`
}

const (
	commandDockerUsage = `gf docker [MAIN] [OPTION]`
	commandDockerBrief = `build docker image for current GoFrame project`
	commandDockerEg    = `
gf docker 
gf docker -t hub.docker.com/john/image:tag
gf docker -p -t hub.docker.com/john/image:tag
gf docker main.go
gf docker main.go -t hub.docker.com/john/image:tag
gf docker main.go -t hub.docker.com/john/image:tag
gf docker main.go -p -t hub.docker.com/john/image:tag
`
	commandDockerDc = `
The "docker" command builds the GF project to a docker images.
It runs "gf build" firstly to compile the project to binary file.
It then runs "docker build" command automatically to generate the docker image.
You should have docker installed, and there must be a Dockerfile in the root of the project.
`
	commandDockerMainBrief  = `main golang file path for "gf build", it's "main.go" in default`
	commandDockerFileBrief  = `file path of the Dockerfile. it's "manifest/docker/Dockerfile" in default`
	commandDockerShellBrief = `path of the shell file which is executed before docker build`
	commandDockerPushBrief  = `auto push the docker image to docker registry if "-t" option passed`
	commandDockerTagBrief   = `tag name for this docker, which is usually used for docker push`
	commandDockerExtraBrief = `extra build options passed to "docker image"`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`commandDockerUsage`:      commandDockerUsage,
		`commandDockerBrief`:      commandDockerBrief,
		`commandDockerEg`:         commandDockerEg,
		`commandDockerDc`:         commandDockerDc,
		`commandDockerMainBrief`:  commandDockerMainBrief,
		`commandDockerFileBrief`:  commandDockerFileBrief,
		`commandDockerShellBrief`: commandDockerShellBrief,
		`commandDockerPushBrief`:  commandDockerPushBrief,
		`commandDockerTagBrief`:   commandDockerTagBrief,
		`commandDockerExtraBrief`: commandDockerExtraBrief,
	})
}

type commandDockerInput struct {
	g.Meta `name:"docker" config:"gfcli.docker"`
	Main   string `name:"MAIN"  arg:"true" brief:"{commandDockerMainBrief}"  d:"main.go"`
	File   string `name:"file"  short:"f"  brief:"{commandDockerFileBrief}"  d:"manifest/docker/Dockerfile"`
	Shell  string `name:"shell" short:"s"  brief:"{commandDockerShellBrief}" d:"manifest/docker/docker.sh"`
	Tag    string `name:"tag"   short:"t"  brief:"{commandDockerTagBrief}"`
	Push   bool   `name:"push"  short:"p"  brief:"{commandDockerPushBrief}" orphan:"true"`
	Extra  string `name:"extra" short:"e"  brief:"{commandDockerExtraBrief}"`
}
type commandDockerOutput struct{}

func (c commandDocker) Index(ctx context.Context, in commandDockerInput) (out *commandDockerOutput, err error) {
	// Necessary check.
	if gproc.SearchBinary("docker") == "" {
		mlog.Fatalf(`command "docker" not found in your environment, please install docker first to proceed this command`)
	}

	// Binary build.
	if err = gproc.ShellRun(fmt.Sprintf(`gf build %s -a amd64 -s linux`, in.Main)); err != nil {
		return
	}
	// Shell executing.
	if gfile.Exists(in.Shell) {
		if err = gproc.ShellRun(gfile.GetContents(in.Shell)); err != nil {
			return
		}
	}
	// Docker build.
	dockerBuildOptions := ""
	if in.Tag != "" {
		dockerBuildOptions = fmt.Sprintf(`-t %s`, in.Tag)
	}
	if in.Extra != "" {
		dockerBuildOptions = fmt.Sprintf(`%s %s`, dockerBuildOptions, in.Extra)
	}
	if err = gproc.ShellRun(fmt.Sprintf(`docker build -f %s . %s`, in.File, dockerBuildOptions)); err != nil {
		return
	}
	// Docker push.
	if in.Tag == "" || !in.Push {
		return
	}
	if err = gproc.ShellRun(fmt.Sprintf(`docker push %s`, in.Tag)); err != nil {
		return
	}
	return
}
