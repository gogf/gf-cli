package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type (
	commandGenPbInput struct {
		g.Meta `name:"pb"`
	}
	commandGenPbOutput struct{}
)

func (c commandGen) Pb(ctx context.Context, in commandGenPbInput) (out *commandGenPbOutput, err error) {
	return
}
