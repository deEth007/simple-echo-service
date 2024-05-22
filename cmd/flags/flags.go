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
)

var Flags = []cli.Flag{
	NodeManagerEnableFlag,
}

func init() {
	Flags = append(Flags, nmgr.CLIFlags(envPrefix)...)
}
