package trunks

import (
	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/urfave/cli/v2"
)

type CLIConfig struct {
	NodeManagerEnable bool

	NodeMgr nmgr.CLIConfig
}

func NewCLIConfig(ctx *cli.Context) *CLIConfig {
	return &CLIConfig{
		NodeManagerEnable: ctx.Bool(flags.NodeManagerEnableFlag.Name),
		NodeMgr:           nmgr.ReadCLIConfig(ctx),
	}
}
