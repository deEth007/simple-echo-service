package trunks

import (
	"github.com/urfave/cli/v2"

	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/tokamak-network/tokamak-trunks/reporter"
)

type CLIConfig struct {
	NodeManagerEnable   bool
	L1RPC               string
	L2RPC               string
	ScenarioFilePath    string
	L1ChainId           uint64
	L2ChainId           uint64
	L2BlockTime         uint64
	L1StandardBrige     string
	L2StandardBrige     string
	L2ToL1MessagePasser string
	Batcher             string
	Proposer            string
	SequencerFeeVault   string

	NodeMgr  nmgr.CLIConfig
	Reporter reporter.CLIConfig
}

func NewCLIConfig(ctx *cli.Context) *CLIConfig {
	return &CLIConfig{
		NodeManagerEnable: ctx.Bool(flags.NodeManagerEnableFlag.Name),
		L1RPC:             ctx.String(flags.L1RPCFlag.Name),
		L2RPC:             ctx.String(flags.L2RPCFlag.Name),
		ScenarioFilePath:  ctx.Path(flags.ScenarioFileFlag.Name),
		L1ChainId:         ctx.Uint64(flags.L1ChainIdFlag.Name),
		L2ChainId:         ctx.Uint64(flags.L2ChainIdFlag.Name),
		Reporter:          reporter.ReadCLIConfig(ctx),
	}
}
