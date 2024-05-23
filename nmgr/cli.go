package nmgr

import (
	"github.com/tokamak-network/tokamak-trunks/utils"
	"github.com/urfave/cli/v2"
)

const (
	DockerComposeDirPathName = "docker-compose-dir-path"
	DeployConfigFilePathName = "deploy-config-file-path"
	AddressFilePathName      = "address-file-path"
	AllocFilePathName        = "alloc-file-path"
)

type CLIConfig struct {
	DockerComposeDirPath string
	DeployConfigFilePath string
	AddressFilePath      string
	AllocFilePath        string
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.PathFlag{
			Name:    DockerComposeDirPathName,
			Usage:   "Docker compose file dir path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "DOCKER_COMPOSE_DIR_PATH"),
		},
		&cli.PathFlag{
			Name:    DeployConfigFilePathName,
			Usage:   "DevConfig file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "DEPLOY_CONFIG_FILE_PATH"),
		},
		&cli.PathFlag{
			Name:    AddressFilePathName,
			Usage:   "Address file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "ADDRESS_FILE_PATH"),
		},
		&cli.PathFlag{
			Name:    AllocFilePathName,
			Usage:   "Alloc file path",
			EnvVars: utils.PrefixEnvVars(envPrefix, "ALLOC_FILE_PATH"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		DockerComposeDirPath: ctx.Path(DockerComposeDirPathName),
		DeployConfigFilePath: ctx.Path(DeployConfigFilePathName),
		AddressFilePath:      ctx.Path(AddressFilePathName),
		AllocFilePath:        ctx.Path(AllocFilePathName),
	}
}
