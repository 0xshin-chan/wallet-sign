package rpc

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/0xshin-chan/wallet-sign/chaindispatcher"
	"github.com/0xshin-chan/wallet-sign/config"
	"github.com/0xshin-chan/wallet-sign/hsm"
	"github.com/0xshin-chan/wallet-sign/protobuf/wallet"
)

const MaxReceivedMessageSize = 1024 * 1024 * 30000

type RpcService struct {
	conf      *config.Config
	HsmClient *hsm.HsmClient
	wallet.UnimplementedWalletServiceServer
	stopped atomic.Bool
}

func (s *RpcService) Stop(ctx context.Context) error {
	s.stopped.Store(true)
	return nil
}

func (s *RpcService) Stopped() bool {
	return s.stopped.Load()
}

func NewRpcService(config *config.Config) (*RpcService, error) {
	rpcService := &RpcService{
		conf: config,
	}
	var hsmCli *hsm.HsmClient
	var hsmErr error
	if config.HsmEnabled {
		hsmCli, hsmErr = hsm.NewHSMClient(context.Background(), config.KeyPath, config.KeyName)
		if hsmErr != nil {
			log.Error("new hsm client fail", "hsmErr", hsmErr)
			return nil, hsmErr
		}
		rpcService.HsmClient = hsmCli
	}
	return rpcService, nil
}

func (s *RpcService) Start(ctx context.Context) error {
	go func(s *RpcService) {
		addr := fmt.Sprintf("%s:%d", s.conf.RpcServer.Host, s.conf.RpcServer.Port)
		log.Info("start rpc service", "addr:", addr)

		opt := grpc.MaxRecvMsgSize(MaxReceivedMessageSize)

		dispatcher, _ := chaindispatcher.NewChainDispatcher(s.conf)

		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(dispatcher.Interceptor),
		)

		defer gs.GracefulStop()

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("could not start tcp listener", "err", err)
		}

		reflection.Register(gs)

		wallet.RegisterWalletServiceServer(gs, dispatcher)

		log.Info("Grpc info", "port", s.conf.RpcServer.Port, "addr", listener.Addr())
		if err := gs.Serve(listener); err != nil {
			log.Error("grpc serve fail", "err", err)
		}
	}(s)
	return nil
}
