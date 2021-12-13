package cmd

import (
	"context"

	"github.com/gogf/gf-cli/v2/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Install = commandInstall{}
)

type commandInstall struct {
	g.Meta `name:"install" brief:"install gf binary to system (might need root/admin permission)"`
}

type commandInstallInput struct {
	g.Meta `name:"install"`
}
type commandInstallOutput struct{}

func (c commandInstall) Index(ctx context.Context, in commandInstallInput) (out *commandInstallOutput, err error) {
	err = service.Install.Run(ctx)
	return
}
