package rpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/huahaiwudi/wallet-sign/hsm"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/protobuf/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync/atomic"
)

const MaxReceivedMessageSize = 1024 * 1024 * 30000

type RpcServiceConfig struct {
	HostName   string
	Port       int
	KeyPath    string
	KeyName    string
	HsmEnabled bool
}

type RpcService struct {
	*RpcServiceConfig
	db        *leveldb.Keys
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

func NewRpcService(db *leveldb.Keys, config *RpcServiceConfig) (*RpcService, error) {
	rpcService := &RpcService{
		RpcServiceConfig: config,
		db:               db,
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
		addr := fmt.Sprintf("%s:%d", s.HostName, s.Port)
		log.Info("start rpc service", "addr:", addr)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("could not start tcp listener", "err", err)
		}
		opt := grpc.MaxRecvMsgSize(MaxReceivedMessageSize)

		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(nil),
		)
		reflection.Register(gs)

		wallet.RegisterWalletServiceServer(gs, s)

		log.Info("Grpc info", "port", s.Port, "addr", listener.Addr())
		if err := gs.Serve(listener); err != nil {
			log.Error("grpc serve fail", "err", err)
		}
	}(s)
	return nil
}
