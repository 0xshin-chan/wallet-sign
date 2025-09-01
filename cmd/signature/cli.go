package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	"github.com/0xshin-chan/wallet-sign/common/cliapp"
	"github.com/0xshin-chan/wallet-sign/config"
	flags2 "github.com/0xshin-chan/wallet-sign/flags"
	"github.com/0xshin-chan/wallet-sign/services/rpc"
)

func runRpc(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	fmt.Println("running grpc services....")
	var f = flag.String("c", "config.yml", "config path")
	flag.Parse()
	cfg, err1 := config.NewConfig(*f)
	if err1 != nil {
		log.Error("new config failed", "err", err1)
		return nil, err1
	}
	return rpc.NewRpcService(cfg)
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
