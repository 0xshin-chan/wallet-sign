package rpc

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/log"
	"github.com/huahaiwudi/wallet-sign/leveldb"
	"github.com/huahaiwudi/wallet-sign/protobuf"
	"github.com/huahaiwudi/wallet-sign/protobuf/wallet"
	"github.com/huahaiwudi/wallet-sign/ssm"
)

const BearerToken = "slim"

func (s *RpcService) GetSupportSignType(ctx context.Context, in *wallet.SupportSignRequest) (*wallet.SupportSignResponse, error) {
	if in.ConsumerToken != BearerToken {
		return &wallet.SupportSignResponse{
			Code:        wallet.ReturnCode_ERROR,
			Message:     "bearer token is error",
			SignWayList: nil,
		}, nil
	}
	var signWay []*wallet.SignWay
	signWay = append(signWay, &wallet.SignWay{Name: "ecdsa"})
	signWay = append(signWay, &wallet.SignWay{Name: "eddsa"})
	return &wallet.SupportSignResponse{
		Code:        wallet.ReturnCode_SUCCESS,
		Message:     "bearer token is success",
		SignWayList: signWay,
	}, nil
}

//			{
//	     "public_key": "0494bffaa512710c724560dfeec756888b973540b6e79463e91185dc362fdf175b7e5831fce5c3a25fc41937258b393a7f9e3187cca2c36f601b03e9cb41cd311c",
//	     "compress_public_key": "0294bffaa512710c724560dfeec756888b973540b6e79463e91185dc362fdf175b"
//	   },
//	   {
//	     "public_key": "0492a9f83eea39c395f3c1e256220e89ebeb84b2e727b0f2a6b3ad8dc9aec1304b44fa3bcf3b08f016eb51b7b977601bb5279f332d4edc999fc85ee829271b1934",
//	     "compress_public_key": "0292a9f83eea39c395f3c1e256220e89ebeb84b2e727b0f2a6b3ad8dc9aec1304b"
//	   },
//	   {
//	     "public_key": "046ded7acb964ddd33c3db7658ee7ce882ac795bf9977ff5c4cefde1e45d59ee3a3ebf322c7971b7fb1bfabbb6a229ca29b1a30fd313ea3c2831eb5724a99e376b",
//	     "compress_public_key": "036ded7acb964ddd33c3db7658ee7ce882ac795bf9977ff5c4cefde1e45d59ee3a"
//	   }
func (s *RpcService) CreateKeyPairsExportPublicKeyList(ctx context.Context, in *wallet.CreateKeyPairAndExportPublicKeyRequest) (*wallet.CreateKeyPairAndExportPublicKeyResponse, error) {
	resp := &wallet.CreateKeyPairAndExportPublicKeyResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	if in.ConsumerToken != BearerToken {
		return resp, nil
	}
	cryptoType, err := protobuf.ParseTransactionType(in.SignType)
	if err != nil {
		resp.Message = "input sign type error"
		return resp, nil
	}

	if in.KeyNum > 20000 {
		resp.Message = "key num must be less than 20000"
		return resp, nil
	}

	var keyList []leveldb.Key
	var exportPublicKeyList []*wallet.ExportPublicKey

	for counter := 0; counter < int(in.KeyNum); counter++ {
		var priKeyStr, pubKeyStr, compressPubKeyStr string
		var err error

		switch cryptoType {
		case protobuf.ECDSA:
			priKeyStr, pubKeyStr, compressPubKeyStr, err = ssm.CreateECDSAKeyPair()
		case protobuf.EDDSA:
			priKeyStr, pubKeyStr, err = ssm.CreateEdDSAKeyPair()
		default:
			return nil, errors.New("unsupported key type")
		}
		if err != nil {
			log.Error("create key pair fail", "err", err)
			return nil, err
		}

		keyItem := leveldb.Key{
			PrivateKey: priKeyStr,
			PublicKey:  pubKeyStr,
		}
		pubKeyItem := &wallet.ExportPublicKey{
			CompressPublicKey: compressPubKeyStr,
			PublicKey:         pubKeyStr,
		}
		exportPublicKeyList = append(exportPublicKeyList, pubKeyItem)
		keyList = append(keyList, keyItem)
	}
	isOK := s.db.StoreKeys(keyList)
	if !isOK {
		log.Error("store keys fail", "isOK", isOK)
		return nil, errors.New("store keys fail")
	}
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.Message = "create keys success"
	resp.PublicKeyList = exportPublicKeyList
	return resp, nil
}

func (s *RpcService) SignMessageSignature(ctx context.Context, in *wallet.SignMessageSignatureRequest) (*wallet.SignMessageSignatureResponse, error) {
	resp := &wallet.SignMessageSignatureResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	cryptoType, err := protobuf.ParseTransactionType(in.SignType)
	if err != nil {
		resp.Message = "input sign type error"
	}

	priKey, isOK := s.db.GetPrivKey(in.PublicKey)
	if !isOK {
		return nil, errors.New("get private key fail")
	}
	var signature string
	var err2 error

	switch cryptoType {
	case protobuf.ECDSA:
		signature, err2 = ssm.SignECDSAMessage(priKey, in.TxMessageHash)
	case protobuf.EDDSA:
		signature, err2 = ssm.SignEdDSAMessage(priKey, in.TxMessageHash)
	default:
		return nil, errors.New("unsupported key type")
	}
	if err2 != nil {
		log.Error("sign tx message fail", "err", err2)
	}
	resp.Message = "sign tx message success"
	resp.Signature = signature
	resp.Code = wallet.ReturnCode_SUCCESS
	return resp, nil
}

func (s *RpcService) SignBatchMessageSignature(ctx context.Context, in *wallet.SignBatchMessageSignatureRequest) (*wallet.SignBatchMessageSignatureResponse, error) {
	resp := &wallet.SignBatchMessageSignatureResponse{
		Code: wallet.ReturnCode_ERROR,
	}
	var msgSignatureList []*wallet.MessageSignature
	for _, msgHash := range in.MessageHashs {
		cryptoType, err := protobuf.ParseTransactionType(msgHash.SignType)
		if err != nil {
			resp.Message = "input sign type error"
		}
		priKey, isOK := s.db.GetPrivKey(msgHash.PublicKey)
		if !isOK {
			log.Error("get private key fail", "err", isOK)
		}
		var signature string
		var err2 error
		switch cryptoType {
		case protobuf.ECDSA:
			signature, err2 = ssm.SignECDSAMessage(priKey, msgHash.TxMessageHash)
		case protobuf.EDDSA:
			signature, err2 = ssm.SignEdDSAMessage(priKey, msgHash.TxMessageHash)
		default:
			return nil, errors.New("unsupported key type")
		}
		if err2 != nil {
			log.Error("sign tx message fail", "err", err2)
			continue
		}
		signItem := &wallet.MessageSignature{
			TxMessageHash: msgHash.TxMessageHash,
			Signature:     signature,
		}
		msgSignatureList = append(msgSignatureList, signItem)
	}
	resp.Message = "sign tx message success"
	resp.Code = wallet.ReturnCode_SUCCESS
	resp.MessageSignatures = msgSignatureList
	return resp, nil
}
