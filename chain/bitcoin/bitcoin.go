package bitcoin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/ethereum/go-ethereum/log"
	"github.com/huahaiwudi/wallet-sign/chain"
	"github.com/huahaiwudi/wallet-sign/config"
	"github.com/huahaiwudi/wallet-sign/hsm"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/protobuf/wallet"
	"github.com/huahaiwudi/wallet-sign/ssm"
)

const ChainName = "Bitcoin"

type ChainAdaptor struct {
	db        *leveldb.Keys
	HsmClient *hsm.HsmClient
	signer    ssm.Signer
}

func NewChainAdaptor(conf *config.Config, db *leveldb.Keys, hsmCli *hsm.HsmClient) (chain.IChainAdaptor, error) {
	return &ChainAdaptor{
		db:        db,
		HsmClient: hsmCli,
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
	var vins []Vin
	vins = append(vins, Vin{
		Hash:   "",
		Index:  0,
		Amount: 0,
	})
	var vouts []Vout
	vouts = append(vouts, Vout{
		Address: "",
		Amount:  0,
		Index:   0,
	})
	bs := BitcoinSchema{
		RequestId: "0",
		Fee:       "0",
		Vins:      vins,
		Vouts:     vouts,
	}
	b, err := json.Marshal(bs)
	if err != nil {
		log.Error("marshal fail", "err", err)
	}
	return &wallet.ChainSchemaResponse{
		Code:    wallet.ReturnCode_SUCCESS,
		Message: "get bitcoin sign schema success",
		Schema:  string(b),
	}, nil
}

func (c ChainAdaptor) CreateKeyPairsExportPublicKeyList(ctx context.Context, request *wallet.CreateKeyPairAndExportPublicKeyRequest) (*wallet.CreateKeyPairAndExportPublicKeyResponse, error) {
	resp := &wallet.CreateKeyPairAndExportPublicKeyResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	if request.KeyNum > 10000 {
		resp.Message = "key num too large"
		return resp, nil
	}
	var keyList []leveldb.Key
	var retKeyList []*wallet.ExportPublicKey
	for counter := 0; counter < int(request.KeyNum); counter++ {
		priKey, pubKey, compressPubKey, err := c.signer.CreateKeyPair()
		if err != nil {
			resp.Message = "Failed to create key pair"
			return resp, nil
		}
		keyList = append(keyList, leveldb.Key{
			PrivateKey: priKey,
			PublicKey:  pubKey,
		})
		retKeyList = append(retKeyList, &wallet.ExportPublicKey{
			PublicKey:         pubKey,
			CompressPublicKey: compressPubKey,
		})
	}
	if isOk := c.db.StoreKeys(keyList); !isOk {
		resp.Message = "Failed to create key pair"
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
		resp.Message = "key num too large"
		return resp, nil
	}
	var keyList []leveldb.Key
	var retKeyListWithAddressList []*wallet.ExportPublicKeyWithAddress

	for counter := 0; counter < int(request.KeyNum); counter++ {
		priKey, pubKey, compressPubKey, err := c.signer.CreateKeyPair()
		if err != nil {
			resp.Message = "Failed to create key pair"
			return resp, nil
		}

		var address string
		compressedPubKeyBytes, _ := hex.DecodeString(compressPubKey)
		pubKeyHash := btcutil.Hash160(compressedPubKeyBytes)
		switch request.AddressFormat {
		case "p2pkh":
			p2pkhAddr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			if err != nil {
				resp.Message = "create p2pkh address fail"
				return resp, nil
			}
			address = p2pkhAddr.EncodeAddress()
			break
		case "p2wpkh":
			witnessAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			if err != nil {
				resp.Message = "create p2wpkh address fail"
				return resp, nil
			}
			address = witnessAddr.EncodeAddress()
			break
		case "p2sh":
			witnessAddr, _ := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			script, err := txscript.PayToAddrScript(witnessAddr)
			if err != nil {
				resp.Message = "create p2sh address fail"
				return resp, nil
			}
			p2shAddr, err := btcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
			if err != nil {
				resp.Message = "create p2sh address fail"
				return resp, nil
			}
			address = p2shAddr.EncodeAddress()
			break
		case "p2tr":
			pubKey, err := btcec.ParsePubKey(compressedPubKeyBytes)
			if err != nil {
				resp.Message = "create p2tr address fail"
				return resp, nil
			}
			taprootPubKdy := schnorr.SerializePubKey(pubKey)
			taprootAddr, err := btcutil.NewAddressTaproot(taprootPubKdy, &chaincfg.MainNetParams)
			if err != nil {
				resp.Message = "create p2tr address fail"
				return resp, nil
			}
			address = taprootAddr.EncodeAddress()
			break
		default:
			resp.Message = "Do not support address type"
			return resp, nil
		}
		keyItem := leveldb.Key{
			PrivateKey: priKey,
			PublicKey:  pubKey,
		}
		pukAddressItem := &wallet.ExportPublicKeyWithAddress{
			PublicKey:         pubKey,
			Address:           address,
			CompressPublicKey: compressPubKey,
		}
		keyList = append(keyList, keyItem)
		retKeyListWithAddressList = append(retKeyListWithAddressList, pukAddressItem)
	}
	if isOK := c.db.StoreKeys(keyList); !isOK {
		resp.Message = "Failed to store key pair"
		return resp, nil
	}
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create keys and address success"
	resp.PublicKeyAddresses = retKeyListWithAddressList
	return resp, nil
}

func (c ChainAdaptor) BuildAndSignTransaction(ctx context.Context, request *wallet.BuildAndSignTransactionRequest) (*wallet.BuildAndSignTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) BuildAndSignBatchTransaction(ctx context.Context, request *wallet.BuildAndSignBatchTransactionRequest) (*wallet.BuildAndSignBatchTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}
