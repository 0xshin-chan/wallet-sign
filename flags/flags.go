package flags

import "github.com/urfave/cli/v2"

const envVarPrefix = "SIGNATURE"

func prefixEnvVars(name string) []string {
	return []string{envVarPrefix + "_" + name}
}

var (
	// RpcHostFlag RPC Service
	RpcHostFlag = &cli.StringFlag{
		Name:     "rpc-host",
		Usage:    "The host of the rpc",
		EnvVars:  prefixEnvVars("RPC_HOST"),
		Required: true,
	}
	RpcPortFlag = &cli.StringFlag{
		Name:     "rpc-port",
		Usage:    "The port of the rpc",
		EnvVars:  prefixEnvVars("RPC_PORT"),
		Required: true,
	}
	// LevelDbPathFlag Database
	LevelDbPathFlag = &cli.StringFlag{
		Name:    "master-db-host",
		Usage:   "The path of the leveldb",
		EnvVars: prefixEnvVars("LEVEL_DB_PATH"),
		Value:   "./",
	}
	CredentialsFileFlag = &cli.StringFlag{
		Name:    "credentials-file",
		Usage:   "The credentials file of cloud hsm",
		EnvVars: prefixEnvVars("CREDENTIALS_FILE"),
	}
	KeyNameFlag = &cli.StringFlag{
		Name:    "key-name",
		Usage:   "The name of cloud hsm",
		EnvVars: prefixEnvVars("KEY_NAME"),
	}
	HsmEnable = &cli.BoolFlag{
		Name:    "hsm-enable",
		Usage:   "hsm enable",
		EnvVars: prefixEnvVars("HSM_ENABLE"),
		Value:   false,
	}
)

var requiredFlags = []cli.Flag{
	RpcHostFlag,
	RpcPortFlag,
	LevelDbPathFlag,
}

var optionalFlags = []cli.Flag{
	CredentialsFileFlag,
	KeyNameFlag,
	HsmEnable,
}

var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}
