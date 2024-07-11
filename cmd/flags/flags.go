package flags

import (
	"github.com/urfave/cli/v2"

	"github.com/tokamak-network/tokamak-trunks/reporter"
	"github.com/tokamak-network/tokamak-trunks/utils"
)

const envPrefix = "TOKAMAK_TRUNKS"

var (
	NodeManagerEnableFlag = &cli.BoolFlag{
		Name:    "node-manager-enable",
		Usage:   "Active node manager",
		EnvVars: utils.PrefixEnvVars(envPrefix, "NODE_MANAGER_ENABLE"),
		Value:   false,
	}
	L1RPCFlag = &cli.StringFlag{
		Name:    "l1-rpc-url",
		Usage:   "Connect L1 chain",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L1_RPC"),
		Value:   "http://localhost:8545",
	}
	L2RPCFlag = &cli.StringFlag{
		Name:    "l2-rpc-url",
		Usage:   "Connect L2 chain",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L2_RPC"),
		Value:   "http://localhost:9545",
	}
	ScenarioFileFlag = &cli.PathFlag{
		Name:    "scenario-file-path",
		Usage:   "Scenario file path",
		EnvVars: utils.PrefixEnvVars(envPrefix, "SCENARIO_FILE_PATH"),
	}
	L1ChainIdFlag = &cli.Uint64Flag{
		Name:    "l1-chain-id",
		Usage:   "L1 chain id",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L1_CHAIN_ID"),
	}
	L2ChainIdFlag = &cli.Uint64Flag{
		Name:    "l2-chain-id",
		Usage:   "L2 chain id",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L2_CHAIN_ID"),
	}
)

var Flags = []cli.Flag{
	L1RPCFlag,
	L2RPCFlag,
	ScenarioFileFlag,
	L1ChainIdFlag,
	L2ChainIdFlag,
}

func init() {
	Flags = append(Flags, reporter.CLIFlags(envPrefix)...)
}
