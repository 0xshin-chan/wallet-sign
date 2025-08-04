package ssm

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
)

func CreateECDSAKeyPair() (priKey string, pubKey string, compressPubKey string, err error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Error("generate key fail", "err", err)
		return EmptyHexString, EmptyHexString, EmptyHexString, err
	}
	privateKeyStr := hex.EncodeToString(crypto.FromECDSA(privateKey))
	publicKeyStr := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	compressPublicKeyStr := hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey))
	return privateKeyStr, publicKeyStr, compressPublicKeyStr, nil
}

func SignECDSAMessage(priKey string, txMsg string) (string, error) {
	hash := common.HexToHash(txMsg)
	priByte, err := hex.DecodeString(priKey)
	if err != nil {
		log.Error("decode private key fail", "err", err)
		return EmptyHexString, err
	}
	priKeyEcdsa, err := crypto.ToECDSA(priByte)
	if err != nil {
		log.Error("Byte private key to ecdsa key fail", "err", err)
		return EmptyHexString, err
	}
	signatureByte, err := crypto.Sign(hash[:], priKeyEcdsa)
	if err != nil {
		log.Error("Sign txMsg fail", "err", err)
		return EmptyHexString, err
	}
	return hex.EncodeToString(signatureByte), nil
}

func VerifyEcdsaSignature(publicKey, txMsg, signature string) (bool, error) {
	pubKeyByte, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Error("decode public key fail", "err", err)
		return false, err
	}
	txMsgByte, err := hex.DecodeString(txMsg)
	if err != nil {
		log.Error("decode txMsg fail", "err", err)
		return false, err
	}
	sigByte, err := hex.DecodeString(signature)
	if err != nil {
		log.Error("decode signature fail", "err", err)
		return false, err
	}
	return crypto.VerifySignature(pubKeyByte, txMsgByte, sigByte[:64]), nil
}
