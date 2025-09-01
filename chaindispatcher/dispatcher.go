package chaindispatcher

import (
	"context"
	"encoding/base64"
	"github.com/status-im/keycard-go/hexutils"
	"runtime/debug"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xshin-chan/wallet-sign/chain"
	"github.com/0xshin-chan/wallet-sign/chain/bitcoin"
	"github.com/0xshin-chan/wallet-sign/chain/ethereum"
	"github.com/0xshin-chan/wallet-sign/chain/solana"
	"github.com/0xshin-chan/wallet-sign/config"
	"github.com/0xshin-chan/wallet-sign/hsm"
	"github.com/0xshin-chan/wallet-sign/leveldb"
	"github.com/0xshin-chan/wallet-sign/protobuf/wallet"
)

const (
	AccessToken = "slim"
	WalletKey   = "wallet key 111"
	RiskKey     = "risk key 111"
)

type CommonRequest interface {
	GetConsumerToken() string
	GetChainName() string
}

type CommonReply = wallet.ChainSignMethodResponse

type ChainDispatcher struct {
	registry map[string]chain.IChainAdaptor
}

func NewChainDispatcher(conf *config.Config) (*ChainDispatcher, error) {
	dispatcher := ChainDispatcher{
		registry: make(map[string]chain.IChainAdaptor),
	}

	chainAdaptorFactoryMap := map[string]func(conf *config.Config, db *leveldb.Keys, hsmCli *hsm.HsmClient) (chain.IChainAdaptor, error){
		bitcoin.ChainName:  bitcoin.NewChainAdaptor,
		ethereum.ChainName: ethereum.NewChainAdaptor,
		solana.ChainName:   solana.NewChainAdaptor,
	}
	supportChains := []string{
		bitcoin.ChainName,
		ethereum.ChainName,
		solana.ChainName,
	}

	db, err := leveldb.NewKeyStore(conf.LevelDbPath)
	if err != nil {
		log.Error("new key store level db", "err", err)
	}
	var hsmClient *hsm.HsmClient
	var errHsmCli error

	if conf.HsmEnabled {
		hsmClient, errHsmCli = hsm.NewHSMClient(context.Background(), conf.KeyPath, conf.KeyName)
		if errHsmCli != nil {
			log.Error("new hsm client fail", "err", errHsmCli)
			return nil, errHsmCli
		}
	}
	for _, chainName := range conf.Chains {
		if factory, ok := chainAdaptorFactoryMap[chainName]; ok {
			adaptor, err := factory(conf, db, hsmClient)
			if err != nil {
				log.Error("failed setup chain", "chain", chainName, "err", err)
			}
			dispatcher.registry[chainName] = adaptor
		} else {
			log.Error("unsupported chain", "chain", chainName, "supportChains", supportChains)
		}
	}
	return &dispatcher, nil
}

func (d *ChainDispatcher) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("panic error", "err", e)
			log.Debug(string(debug.Stack()))
			err = status.Errorf(codes.Internal, "panic error: %v", e)
		}
	}()

	pos := strings.LastIndex(info.FullMethod, "/")
	method := info.FullMethod[pos+1:]

	chainName := req.(CommonRequest).GetChainName()
	log.Info(method, "chain", chainName, "req", req)

	resp, err = handler(ctx, req)
	log.Debug("finish handling", "resp", resp, "err", err)
	return
}

func (d *ChainDispatcher) preHandler(req interface{}) (resp *CommonReply) {
	// proto 生成的 Go struct 已经实现了接口，因为生成的代码里自带了 GetConsumerToken() 和 GetChainName() 方法。
	consumerToken := req.(CommonRequest).GetConsumerToken()
	log.Debug("consumer token", "consumerToken", consumerToken, "req", req)
	if consumerToken != AccessToken {
		return &CommonReply{
			Code:    wallet.ReturnCode_ERROR,
			Message: "invalid consumer token",
		}
	}

	chainName := req.(CommonRequest).GetChainName()
	log.Debug("chain", chainName, "req", req)
	if _, ok := d.registry[chainName]; !ok {
		return &CommonReply{
			Code:    wallet.ReturnCode_ERROR,
			Message: "unsupported chain",
		}
	}
	return nil
}

func (d *ChainDispatcher) GetChainSignMethod(ctx context.Context, request *wallet.ChainSignMethodRequest) (*wallet.ChainSignMethodResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.ChainSignMethodResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].GetChainSignMethod(ctx, request)
}

func (d *ChainDispatcher) GetChainSchema(ctx context.Context, request *wallet.ChainSchemaRequest) (*wallet.ChainSchemaResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.ChainSchemaResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].GetChainSchema(ctx, request)
}

func (d *ChainDispatcher) CreateKeyPairsExportPublicKeyList(ctx context.Context, request *wallet.CreateKeyPairAndExportPublicKeyRequest) (*wallet.CreateKeyPairAndExportPublicKeyResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.CreateKeyPairAndExportPublicKeyResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].CreateKeyPairsExportPublicKeyList(ctx, request)
}

func (d *ChainDispatcher) CreateKeyPairsWithAddresses(ctx context.Context, request *wallet.CreateKeyPairsWithAddressesRequest) (*wallet.CreateKeyPairsWithAddressesResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.CreateKeyPairsWithAddressesResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].CreateKeyPairsWithAddresses(ctx, request)
}

func (d *ChainDispatcher) SignTransactionMessage(ctx context.Context, request *wallet.SignTransactionMessageRequest) (*wallet.SignTransactionMessageResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.SignTransactionMessageResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].SignTransactionMessage(ctx, request)
}

func (d *ChainDispatcher) BuildAndSignTransaction(ctx context.Context, request *wallet.BuildAndSignTransactionRequest) (*wallet.BuildAndSignTransactionResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.BuildAndSignTransactionResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	//验证 walletKey 和 riskKey
	txReqJsonByte, err := base64.StdEncoding.DecodeString(request.TxBase64Body)
	if err != nil {
		return &wallet.BuildAndSignTransactionResponse{
			Code:    wallet.ReturnCode_ERROR,
			Message: "decode base64 string fail",
		}, nil
	}
	RiskKeyHash := crypto.Keccak256(append(txReqJsonByte, []byte(RiskKey)...))
	RiskKeyHashStr := hexutils.BytesToHex(RiskKeyHash)
	if RiskKeyHashStr != request.RiskKeyHash {
		return &wallet.BuildAndSignTransactionResponse{
			Code:    wallet.ReturnCode_ERROR,
			Message: "riskKey hash check fail",
		}, nil
	}
	WalletKeyHash := crypto.Keccak256(append(txReqJsonByte, []byte(WalletKey)...))
	WalletKeyHashStr := hexutils.BytesToHex(WalletKeyHash)
	if WalletKeyHashStr != request.WalletKeyHash {
		return &wallet.BuildAndSignTransactionResponse{
			Code:    wallet.ReturnCode_ERROR,
			Message: "wallet hash check fail",
		}, nil
	}
	return d.registry[request.ChainName].BuildAndSignTransaction(ctx, request)
}

func (d *ChainDispatcher) BuildAndSignBatchTransaction(ctx context.Context, request *wallet.BuildAndSignBatchTransactionRequest) (*wallet.BuildAndSignBatchTransactionResponse, error) {
	resp := d.preHandler(request)
	if resp != nil {
		return &wallet.BuildAndSignBatchTransactionResponse{
			Code:    resp.Code,
			Message: resp.Message,
		}, nil
	}
	return d.registry[request.ChainName].BuildAndSignBatchTransaction(ctx, request)
}
