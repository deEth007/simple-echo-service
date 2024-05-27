package flags

import (
	"github.com/tokamak-network/tokamak-trunks/nmgr"
	"github.com/tokamak-network/tokamak-trunks/utils"

	"github.com/urfave/cli/v2"
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
	L1StandardBrige = &cli.StringFlag{
		Name:    "l1-standard-bridge",
		Usage:   "L1StandardBrige Address",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L1_STANDARD_BRIDGE"),
	}
	L2StandardBrige = &cli.StringFlag{
		Name:    "l2-standard-bridge",
		Usage:   "L2StandardBrige Address",
		EnvVars: utils.PrefixEnvVars(envPrefix, "L2_STANDARD_BRIDGE"),
	}
	OutputFileFlag = &cli.StringFlag{
		Name:    "output-file-name",
		Usage:   "Output result file name",
		EnvVars: utils.PrefixEnvVars(envPrefix, "OUTPUT_FILE_NAME"),
	}
)

var Flags = []cli.Flag{
	NodeManagerEnableFlag,
	L1RPCFlag,
	L2RPCFlag,
	L1ChainIdFlag,
	L2ChainIdFlag,
	L1StandardBrige,
	L2StandardBrige,
	OutputFileFlag,
}

func init() {
	Flags = append(Flags, nmgr.CLIFlags(envPrefix)...)
}
