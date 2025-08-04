package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/huahaiwudi/wallet-sign/common/cliapp"
	"github.com/huahaiwudi/wallet-sign/config"
	flags2 "github.com/huahaiwudi/wallet-sign/flags"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/services/rpc"
	"github.com/urfave/cli/v2"
)

func runRpc(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	fmt.Println("running grpc services....")
	cfg := config.NewConfig(ctx)
	grpcServerCfg := &rpc.RpcServiceConfig{
		HostName:   cfg.RpcServer.Host,
		Port:       cfg.RpcServer.Port,
		KeyName:    cfg.KeyName,
		KeyPath:    cfg.CredentialsFile,
		HsmEnabled: cfg.HsmEnabled,
	}
	db, err := leveldb.NewKeyStore(cfg.LevelDbPath)
	if err != nil {
		log.Error("new key store level db", "err", err)
	}
	return rpc.NewRpcService(db, grpcServerCfg)
}

func NewCli() *cli.App {
	flags := flags2.Flags
	return &cli.App{
		Version:              "0.0.1",
		Description:          "wallet sign service",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "rpc",
				Flags:       flags,
				Description: "Run rpc services",
				Action:      cliapp.LifecycleCmd(runRpc),
			},
			{
				Name:        "version",
				Description: "Show project version",
				Action: func(ctx *cli.Context) error {
					cli.ShowVersion(ctx)
					return nil
				},
			},
		},
	}
}
