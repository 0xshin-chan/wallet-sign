package solana

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"github.com/cosmos/btcutil/base58"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"

	"github.com/huahaiwudi/wallet-sign/chain"
	"github.com/huahaiwudi/wallet-sign/config"
	"github.com/huahaiwudi/wallet-sign/hsm"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/protobuf/wallet"
	"github.com/huahaiwudi/wallet-sign/ssm"
)

const ChainName = "Solana"

type ChainAdaptor struct {
	db        *leveldb.Keys
	HsmClient *hsm.HsmClient
	signer    ssm.Signer
}

func NewChainAdaptor(conf *config.Config, db *leveldb.Keys, hsmCli *hsm.HsmClient) (chain.IChainAdaptor, error) {
	return &ChainAdaptor{
		db:        db,
		HsmClient: hsmCli,
		signer:    &ssm.EdDSASigner{},
	}, nil
}

func (c ChainAdaptor) GetChainSignMethod(ctx context.Context, request *wallet.ChainSignMethodRequest) (*wallet.ChainSignMethodResponse, error) {
	return &wallet.ChainSignMethodResponse{
		Code:       wallet.ReturnCode_SUCCESS,
		Message:    "get sign method success",
		SignMethod: "eddsa",
	}, nil
}

func (c ChainAdaptor) GetChainSchema(ctx context.Context, request *wallet.ChainSchemaRequest) (*wallet.ChainSchemaResponse, error) {
	ss := SolanaSchema{
		Nonce:           "",
		GasPrice:        "",
		GasTipCap:       "",
		GasFeeCap:       "",
		Gas:             0,
		ContractAddress: "",
		FromAddress:     "",
		ToAddress:       "",
		TokenId:         "",
		Value:           "",
	}
	b, err := json.Marshal(ss)
	if err != nil {
		log.Error("marshal fail", "err", err)
	}
	return &wallet.ChainSchemaResponse{
		Code:    wallet.ReturnCode_SUCCESS,
		Message: "get solana sign schema success",
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
		priKeyStr, pubKeyStr, compressPubKeyStr, err := c.signer.CreateKeyPair()
		if err != nil {
			log.Error("create key fail", "err", err)
		}
		keyItem := leveldb.Key{
			PrivateKey: priKeyStr,
			PublicKey:  pubKeyStr,
		}
		pubKeyItem := &wallet.ExportPublicKey{
			PublicKey:         pubKeyStr,
			CompressPublicKey: compressPubKeyStr,
		}

		keyList = append(keyList, keyItem)
		retKeyList = append(retKeyList, pubKeyItem)
	}
	if isOk := c.db.StoreKeys(keyList); !isOk {
		resp.Message = "store keys fail"
		return resp, nil
	}

	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create keys success"
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
	var retKeyList []*wallet.ExportPublicKeyWithAddress
	for counter := 0; counter < int(request.KeyNum); counter++ {
		priKeyStr, pubKeyStr, compressPubKeyStr, err := c.signer.CreateKeyPair()
		if err != nil {
			log.Error("create key fail", "err", err)
		}
		keyItem := leveldb.Key{
			PrivateKey: priKeyStr,
			PublicKey:  pubKeyStr,
		}

		address, err := PubKeyHexToAddress(pubKeyStr)
		if err != nil {
			resp.Message = "public key to address fail"
			return resp, nil
		}
		pubKeyItem := &wallet.ExportPublicKeyWithAddress{
			PublicKey:         pubKeyStr,
			CompressPublicKey: compressPubKeyStr,
			Address:           address,
		}

		keyList = append(keyList, keyItem)
		retKeyList = append(retKeyList, pubKeyItem)
	}
	if isOk := c.db.StoreKeys(keyList); !isOk {
		resp.Message = "store keys fail"
		return resp, nil
	}

	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create keys success"
	resp.PublicKeyAddresses = retKeyList
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
	//base64 => byte
	jsonBytes, err := base64.StdEncoding.DecodeString(request.TxBase64Body)
	if err != nil {
		resp.Message = "base64 decode fail"
		return resp, nil
	}
	//byte => solanaSchema
	var data SolanaSchema
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		resp.Message = "json unmarshal fail"
		return resp, nil
	}
	//将value转为十进制，uint64
	value, _ := strconv.ParseUint(data.Value, 10, 64)
	//将from地址从base58转为solana.PublicKey类型
	fromPubKey, err := solana.PublicKeyFromBase58(data.FromAddress)
	if err != nil {
		resp.Message = "Failed to parse public key from base58 by from address"
		return resp, nil
	}
	toPubKey, err := solana.PublicKeyFromBase58(data.ToAddress)
	if err != nil {
		resp.Message = "Failed to parse public key from base58 by to address"
		return resp, nil
	}
	//定义solana交易
	var tx *solana.Transaction
	//判断是否是sol币交易，如果是就直接构建交易，如果不是，处理代币合约地址
	if isSOLTransfer(data.ContractAddress) {
		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					value,
					fromPubKey,
					toPubKey,
				).Build(),
			},
			solana.MustHashFromBase58(data.Nonce),
			solana.TransactionPayer(fromPubKey),
		)
	} else {
		mintPubKey := solana.MustPublicKeyFromBase58(data.ContractAddress)
		fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			fromPubKey,
			mintPubKey,
		)
		if err != nil {
			resp.Message = "Failed to find associated token address"
			return resp, nil
		}
		toTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			toPubKey,
			mintPubKey,
		)
		if err != nil {
			resp.Message = "Failed to find associated token address"
			return resp, nil
		}
		decimals := data.Decimal

		//把value转为float64
		valueFloat, err := strconv.ParseFloat(data.Value, 64)
		if err != nil {
			resp.Message = "Failed to parse value to float"
			return resp, nil
		}
		actualValue := uint64(valueFloat * math.Pow10(int(decimals)))

		transferInstruction := token.NewTransferInstruction(
			actualValue,
			fromTokenAccount,
			toTokenAccount,
			fromPubKey,
			[]solana.PublicKey{},
		).Build()
		//如果处理有错误，或者在交易体中TokenCreate为true（及在交易体中要求给toAddress创建ATA）时创建ATA
		if err != nil || data.TokenCreate {
			createATAInstruction := associatedtokenaccount.NewCreateInstruction(
				fromPubKey,
				toPubKey,
				mintPubKey,
			).Build()
			tx, err = solana.NewTransaction(
				[]solana.Instruction{createATAInstruction, transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
			//否则就直接构建交易
		} else {
			tx, err = solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
		}
	}
	log.Info("Transaction", tx.String())
	//tx =》 bytes
	txm, _ := tx.Message.MarshalBinary()
	//bytes => hex
	signingMessageHex := hex.EncodeToString(txm)

	log.Info("this is we should use sign message hash", "signingMessageHex", signingMessageHex)
	priKey, isOk := c.db.GetPrivKey(request.PublicKey)
	if !isOk {
		resp.Message = "get private key fail"
		return resp, nil
	}
	txSignatures, err := c.signer.SignMessage(priKey, signingMessageHex)
	if err != nil {
		resp.Message = "sign message fail"
		return resp, nil
	}
	if len(txSignatures) == 0 {
		tx.Signatures = make([]solana.Signature, 1)
	}
	if len(txSignatures) != 64 {
		resp.Message = "invalid signature length"
		return resp, nil
	}
	var solanaSig solana.Signature
	copy(solanaSig[:], txSignatures)
	tx.Signatures[0] = solanaSig
	//展示交易，类似logInfo
	spew.Dump(tx)
	if err := tx.VerifySignatures(); err != nil {
		resp.Message = "verify signatures fail"
		return resp, nil
	}
	serializedTx, err := tx.MarshalBinary()
	if err != nil {
		resp.Message = "Failed to serialize transaction"
		return resp, nil
	}
	log.Info("serialized transaction", "serializedTx", serializedTx)
	base58Tx := base58.Encode(serializedTx)
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.SignedTx = base58Tx
	return resp, nil
}

func (c ChainAdaptor) BuildAndSignBatchTransaction(ctx context.Context, request *wallet.BuildAndSignBatchTransactionRequest) (*wallet.BuildAndSignBatchTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func isSOLTransfer(coinAddress string) bool {
	return coinAddress == "" || coinAddress == "So11111111111111111111111111111111111111112"
}
