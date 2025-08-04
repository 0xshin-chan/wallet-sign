package config

import (
	"github.com/huahaiwudi/wallet-sign/flags"
	"github.com/urfave/cli/v2"
)

type ServerConfig struct {
	Host string
	Port int
}

type Config struct {
	LevelDbPath     string
	RpcServer       ServerConfig
	CredentialsFile string
	KeyName         string
	HsmEnabled      bool
}

func NewConfig(ctx *cli.Context) Config {
	return Config{
		LevelDbPath:     ctx.String(flags.LevelDbPathFlag.Name),
		CredentialsFile: ctx.String(flags.CredentialsFileFlag.Name),
		KeyName:         ctx.String(flags.KeyNameFlag.Name),
		HsmEnabled:      ctx.Bool(flags.HsmEnable.Name),
		RpcServer: ServerConfig{
			Host: ctx.String(flags.RpcHostFlag.Name),
			Port: ctx.Int(flags.RpcPortFlag.Name),
		},
	}
}
