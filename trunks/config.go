package trunks

import (
	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/urfave/cli/v2"
)

type CLIConfig struct {
	NodeManagerEnable bool
	L1RPC             string
	L2RPC             string
	L1ChainId         uint64
	L2ChainId         uint64
	L1StandardBrige   string
	L2StandardBrige   string

	NodeMgr nmgr.CLIConfig
}

func NewCLIConfig(ctx *cli.Context) *CLIConfig {
	return &CLIConfig{
		NodeManagerEnable: ctx.Bool(flags.NodeManagerEnableFlag.Name),
		L1RPC:             ctx.String(flags.L1RPCFlag.Name),
		L2RPC:             ctx.String(flags.L2RPCFlag.Name),
		L1ChainId:         ctx.Uint64(flags.L1ChainIdFlag.Name),
		L2ChainId:         ctx.Uint64(flags.L2ChainIdFlag.Name),
		L1StandardBrige:   ctx.String(flags.L1StandardBrige.Name),
		L2StandardBrige:   ctx.String(flags.L2StandardBrige.Name),
		NodeMgr:           nmgr.ReadCLIConfig(ctx),
	}
}
