package ethereum

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/huahaiwudi/wallet-sign/chain"
	"github.com/huahaiwudi/wallet-sign/config"
	"github.com/huahaiwudi/wallet-sign/hsm"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/protobuf/wallet"
	"github.com/huahaiwudi/wallet-sign/ssm"
)

const ChainName = "Ethereum"

type ChainAdaptor struct {
	db        *leveldb.Keys
	HsmClient *hsm.HsmClient
	signer    *ssm.ECDSASigner
}

func NewChainAdaptor(conf *config.Config, db *leveldb.Keys, hsmCli *hsm.HsmClient) (chain.IChainAdaptor, error) {
	return &ChainAdaptor{
		db:        db,
		HsmClient: hsmCli,
		signer:    &ssm.ECDSASigner{},
	}, nil
}

func (c ChainAdaptor) GetChainSignMethod(ctx context.Context, request *wallet.ChainSignMethodRequest) (*wallet.ChainSignMethodResponse, error) {
	return &wallet.ChainSignMethodResponse{
		Code:       wallet.ReturnCode_SUCCESS,
		Message:    "get sign method success",
		SignMethod: "ecdsa",
	}, nil
}

func (c ChainAdaptor) GetChainSchema(ctx context.Context, request *wallet.ChainSchemaRequest) (*wallet.ChainSchemaResponse, error) {
	es := EthereumSchema{
		RequestId: "0",
		DynamicFeeTx: Eip1559DynamicFeeTx{
			ChainId:              "",
			Nonce:                0,
			FromAddress:          common.Address{}.String(),
			ToAddress:            common.Address{}.String(),
			GasLimit:             0,
			Gas:                  0,
			MaxFeePerGas:         "0",
			MaxPriorityFeePerGas: "0",
			Amount:               "0",
			ContractAddress:      "",
		},
		ClassicFeeTx: LegacyFeeTx{
			ChainId:         "0",
			Nonce:           0,
			FromAddress:     common.Address{}.String(),
			ToAddress:       common.Address{}.String(),
			GasLimit:        0,
			GasPrice:        0,
			Amount:          "0",
			ContractAddress: "",
		},
	}
	b, err := json.Marshal(es)
	if err != nil {
		log.Error("marshal fail", "err", err)
	}
	return &wallet.ChainSchemaResponse{
		Code:    wallet.ReturnCode_SUCCESS,
		Message: "get ethereum sign schema success",
		Schema:  string(b),
	}, nil
}

func (c ChainAdaptor) CreateKeyPairsExportPublicKeyList(ctx context.Context, request *wallet.CreateKeyPairAndExportPublicKeyRequest) (*wallet.CreateKeyPairAndExportPublicKeyResponse, error) {

	resp := &wallet.CreateKeyPairAndExportPublicKeyResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	if request.KeyNum > 10000 {
		resp.Message = "Number must be less than 10000"
		return resp, nil
	}

	var keyList []leveldb.Key
	var retKeyList []*wallet.ExportPublicKey

	for counter := 0; counter < int(request.KeyNum); counter++ {
		priKey, pubKey, compressPubKey, err := c.signer.CreateKeyPair()
		if err != nil {
			resp.Message = "create key pair fail"
			return resp, nil
		}
		keyItem := leveldb.Key{
			PrivateKey: priKey,
			PublicKey:  pubKey,
		}
		pukItem := &wallet.ExportPublicKey{
			PublicKey:         pubKey,
			CompressPublicKey: compressPubKey,
		}
		retKeyList = append(retKeyList, pukItem)
		keyList = append(keyList, keyItem)
	}
	isOk := c.db.StoreKeys(keyList)
	if !isOk {
		log.Error("store keys fail", "isOk", isOk)
		return nil, errors.New("store keys fail")
	}
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create key pair success"
	resp.PublicKeyList = retKeyList
	return resp, nil
}

func (c ChainAdaptor) CreateKeyPairsWithAddresses(ctx context.Context, request *wallet.CreateKeyPairsWithAddressesRequest) (*wallet.CreateKeyPairsWithAddressesResponse, error) {
	resp := &wallet.CreateKeyPairsWithAddressesResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	if request.KeyNum > 10000 {
		resp.Message = "Number must be less than 10000"
		return resp, nil
	}

	var keyList []leveldb.Key
	var retKeyWithAddrList []*wallet.ExportPublicKeyWithAddress
	for counter := 0; counter < int(request.KeyNum); counter++ {
		priKey, pubKey, compressPubKey, err := c.signer.CreateKeyPair()
		if err != nil {
			resp.Message = "create key pair fail"
			return resp, nil
		}
		keyItem := leveldb.Key{
			PrivateKey: priKey,
			PublicKey:  pubKey,
		}

		publicKeyByte, err := hex.DecodeString(pubKey)
		pukAddrItem := &wallet.ExportPublicKeyWithAddress{
			PublicKey:         pubKey,
			CompressPublicKey: compressPubKey,
			Address:           common.BytesToAddress(crypto.Keccak256(publicKeyByte[1:])[12:]).String(),
		}
		retKeyWithAddrList = append(retKeyWithAddrList, pukAddrItem)
		keyList = append(keyList, keyItem)
	}
	isOk := c.db.StoreKeys(keyList)
	if !isOk {
		log.Error("store keys fail", "isOk", isOk)
		return nil, errors.New("store keys fail")
	}
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create key pairs with address success"
	resp.PublicKeyAddresses = retKeyWithAddrList
	return resp, nil
}

func (c ChainAdaptor) SignTransactionMessage(ctx context.Context, request *wallet.SignTransactionMessageRequest) (*wallet.SignTransactionMessageResponse, error) {
	resp := &wallet.SignTransactionMessageResponse{
		Code: wallet.ReturnCode_ERROR,
	}

	privKey, isOk := c.db.GetPrivKey(request.PublicKey)
	if !isOk {
		return nil, errors.New("get private key fail")
	}

	signature, err := c.signer.SignMessage(privKey, request.MessageHash)
	if err != nil {
		log.Error("sign message fail", "err", err)
	}

	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "sign message success"
	resp.Signature = signature
	return resp, nil
}

func (c ChainAdaptor) BuildAndSignTransaction(ctx context.Context, request *wallet.BuildAndSignTransactionRequest) (*wallet.BuildAndSignTransactionResponse, error) {

	resp := &wallet.BuildAndSignTransactionResponse{
		Code: wallet.ReturnCode_ERROR,
	}

	dFeeTx, _, err := c.buildDynamicFeeTx(request.TxBase64Body)
	if err != nil {
		return nil, err
	}

	rawTx, err := CreateEip1559UnSignTx(dFeeTx, dFeeTx.ChainID)
	if err != nil {
		log.Error("create un sign tx fail", "err", err)
		resp.Message = "get un sign tx fail"
		return resp, nil
	}

	privKey, isOk := c.db.GetPrivKey(request.PublicKey)
	log.Error("priKey ==== ", privKey, "pubKey ==== ", request.PublicKey)

	if !isOk {
		log.Error("get private key by public key fail", "err", err)
		resp.Message = "get private key by public key fail"
		return resp, nil
	}

	signature, err := c.signer.SignMessage(privKey, rawTx)
	if err != nil {
		log.Error("sign transaction fail", "err", err)
		resp.Message = "sign transaction fail"
		return resp, nil
	}

	inputSignatureByteList, err := hex.DecodeString(signature)
	if err != nil {
		log.Error("decode signature failed", "err", err)
		resp.Message = "decode signature failed"
		return resp, nil
	}

	eip1559Signer, signedTx, signAndHandledTx, txHash, err := CreateEip1559SignedTx(dFeeTx, inputSignatureByteList, dFeeTx.ChainID)
	if err != nil {
		log.Error("create signed tx fail", "err", err)
		resp.Message = "create signed tx fail"
		return resp, nil
	}
	log.Info("sign transaction success",
		"eip1559Signer", eip1559Signer,
		"signedTx", signedTx,
		"signAndHandledTx", signAndHandledTx,
		"txHash", txHash,
	)
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "sign whole transaction success"
	resp.SignedTx = signAndHandledTx
	resp.TxHash = txHash
	resp.TxMessageHash = rawTx
	return resp, nil
}

func (c ChainAdaptor) BuildAndSignBatchTransaction(ctx context.Context, request *wallet.BuildAndSignBatchTransactionRequest) (*wallet.BuildAndSignBatchTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) buildDynamicFeeTx(base64Tx string) (*types.DynamicFeeTx, *Eip1559DynamicFeeTx, error) {
	// 1. Decode base64 string
	txReqJsonByte, err := base64.StdEncoding.DecodeString(base64Tx)
	if err != nil {
		log.Error("decode string fail", "err", err)
		return nil, nil, err
	}

	// 2. Unmarshal JSON to struct
	var dynamicFeeTx Eip1559DynamicFeeTx
	if err := json.Unmarshal(txReqJsonByte, &dynamicFeeTx); err != nil {
		log.Error("parse json fail", "err", err)
		return nil, nil, err
	}

	// 3. Convert string values to big.Int
	chainID := new(big.Int)
	maxPriorityFeePerGas := new(big.Int)
	maxFeePerGas := new(big.Int)
	amount := new(big.Int)

	//把 chainId等 从字符串 解析成整数 *big.Int，以便后续构造交易
	if _, ok := chainID.SetString(dynamicFeeTx.ChainId, 10); !ok {
		return nil, nil, fmt.Errorf("invalid chain ID: %s", dynamicFeeTx.ChainId)
	}
	if _, ok := maxPriorityFeePerGas.SetString(dynamicFeeTx.MaxPriorityFeePerGas, 10); !ok {
		return nil, nil, fmt.Errorf("invalid max priority fee: %s", dynamicFeeTx.MaxPriorityFeePerGas)
	}
	if _, ok := maxFeePerGas.SetString(dynamicFeeTx.MaxFeePerGas, 10); !ok {
		return nil, nil, fmt.Errorf("invalid max fee: %s", dynamicFeeTx.MaxFeePerGas)
	}
	if _, ok := amount.SetString(dynamicFeeTx.Amount, 10); !ok {
		return nil, nil, fmt.Errorf("invalid amount: %s", dynamicFeeTx.Amount)
	}

	// 4. Handle addresses and data
	//用于将字符串形式（十六进制）的以太坊地址转换为标准格式的 common.Address 类型（长度固定为 20 字节）
	//以太坊交易中，To 字段必须是一个 common.Address 类型，而不是字符串。 因此，在构建交易时，必须先做类型转换
	toAddress := common.HexToAddress(dynamicFeeTx.ToAddress)
	var finalToAddress common.Address
	var finalAmount *big.Int
	var buildData []byte
	log.Info("contract address check",
		"contractAddress", dynamicFeeTx.ContractAddress,
		"isEthTransfer", isEthTransfer(&dynamicFeeTx),
	)

	// 5. Handle contract interaction vs direct transfer
	if isEthTransfer(&dynamicFeeTx) {
		finalToAddress = toAddress
		finalAmount = amount
	} else {
		contractAddress := common.HexToAddress(dynamicFeeTx.ContractAddress)
		buildData = BuildErc20Data(toAddress, amount)
		finalToAddress = contractAddress
		finalAmount = big.NewInt(0)
	}

	// 6. Create dynamic fee transaction
	dFeeTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     dynamicFeeTx.Nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       dynamicFeeTx.GasLimit,
		To:        &finalToAddress,
		Value:     finalAmount,
		Data:      buildData,
	}

	return dFeeTx, &dynamicFeeTx, nil
}

// 判断你是否为 ETH 转账
func isEthTransfer(tx *Eip1559DynamicFeeTx) bool {
	//合约地址是否为零地址
	if tx.ContractAddress == "" ||
		tx.ContractAddress == "0x0000000000000000000000000000000000000000" ||
		tx.ContractAddress == "0x00" {
		return true
	}
	return false
}

/*
{
    "chain_id": "11155111",
	"nonce": 0,
	"from_address": "0x0749F85b38614DcE2ec02b8F0b118A8A235C300b",
	"to_address": "0x02A2AaB257B5d51CE8207be790BDd6168cFB38B5",
	"gas_Limit": 21000,
	"gas": 200000,
	"max_Fee_Per_Gas": "327993150328",
	"max_Priority_Fee_Per_Gas": "32799315032",
	"amount": "140000000000000000",
	"contract_address": "0x00"
}
*/
