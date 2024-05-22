package nmgr

import (
	"github.com/tokamak-network/tokamak-trunks/utils"
	"github.com/urfave/cli/v2"
)

const (
	DockerComposeFileDirPathName = "docker-compose-file-dir-path"
	L1GenesisFilePathName        = "l1-genesis-file-path"
	L2GenesisFilePathName        = "l2-genesis-file-path"
	RollupConfigFilePathName     = "rollup-config-file-path"
	AddressFilePathName          = "address-file-path"
	JwtSecretFilePathName        = "jwt-file-path"
)

type CLIConfig struct {
	DockerComposeFileDirPath string
	L1GenesisFilePath        string
	L2GenesisFilePath        string
	RollupConfigFilePath     string
	AddressFilePath          string
	JwtSecretFilePath        string
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    DockerComposeFileDirPathName,
			Usage:   "Docker compose file dir path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "DOCKER_COMPOSE_FILE_DIR_PATH"),
		},
		&cli.StringFlag{
			Name:    L1GenesisFilePathName,
			Usage:   "L1 genesis file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "L1_GENESIS_FILE_PATH"),
		},
		&cli.StringFlag{
			Name:    L2GenesisFilePathName,
			Usage:   "L2 genesis file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "L2_GENESIS_FILE_PATH"),
		},
		&cli.StringFlag{
			Name:    RollupConfigFilePathName,
			Usage:   "Rollup config file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "ROLLUP_CONFIG_FILE_PATH"),
		},
		&cli.StringFlag{
			Name:    AddressFilePathName,
			Usage:   "Address file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "ADDRESS_FILE_PATH"),
		},
		&cli.StringFlag{
			Name:    JwtSecretFilePathName,
			Usage:   "Jwt Secret file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "JWT_SECRET_FILE_PATH"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		DockerComposeFileDirPath: ctx.String(DockerComposeFileDirPathName),
		L1GenesisFilePath:        ctx.String(L1GenesisFilePathName),
		L2GenesisFilePath:        ctx.String(L2GenesisFilePathName),
		RollupConfigFilePath:     ctx.String(RollupConfigFilePathName),
		AddressFilePath:          ctx.String(AddressFilePathName),
		JwtSecretFilePath:        ctx.String(JwtSecretFilePathName),
	}
}
