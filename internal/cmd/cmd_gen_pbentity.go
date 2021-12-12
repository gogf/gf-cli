package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type (
	commandGenPbEntityInput struct {
		g.Meta `name:"pbentity"`
	}
	commandGenPbEntityOutput struct{}
)

func (c commandGen) PbEntity(ctx context.Context, in commandGenPbEntityInput) (out *commandGenPbEntityOutput, err error) {
	return
}
