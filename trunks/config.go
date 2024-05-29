package trunks

import (
	"github.com/tokamak-network/tokamak-trunks/cmd/flags"
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/tokamak-network/tokamak-trunks/reporter"
	"github.com/urfave/cli/v2"
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
		NodeManagerEnable:   ctx.Bool(flags.NodeManagerEnableFlag.Name),
		L1RPC:               ctx.String(flags.L1RPCFlag.Name),
		L2RPC:               ctx.String(flags.L2RPCFlag.Name),
		ScenarioFilePath:    ctx.Path(flags.ScenarioFileFlag.Name),
		L1ChainId:           ctx.Uint64(flags.L1ChainIdFlag.Name),
		L2ChainId:           ctx.Uint64(flags.L2ChainIdFlag.Name),
		L1StandardBrige:     ctx.String(flags.L1StandardBrige.Name),
		L2StandardBrige:     ctx.String(flags.L2StandardBrige.Name),
		L2ToL1MessagePasser: ctx.String(flags.L2ToL1MessagePasser.Name),
		Batcher:             ctx.String(flags.Batcher.Name),
		Proposer:            ctx.String(flags.Proposer.Name),
		SequencerFeeVault:   ctx.String(flags.SequencerFeeVault.Name),
		NodeMgr:             nmgr.ReadCLIConfig(ctx),
		Reporter:            reporter.ReadCLIConfig(ctx),
	}
}
