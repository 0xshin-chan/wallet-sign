package ssm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/log"
)

func CreateEdDSAKeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error("create key pair fail", "err", err)
		return EmptyHexString, EmptyHexString, err
	}
	return hex.EncodeToString(privateKey), hex.EncodeToString(publicKey), err
}

func SignEdDSAMessage(priKey string, txMsg string) (string, error) {
	priKeyByte, err := hex.DecodeString(priKey)
	if err != nil {
		log.Error("decode private key fail", "err", err)
		return "", err
	}
	txMsgByte, err := hex.DecodeString(txMsg)
	if err != nil {
		log.Error("decode tx message fail", "err", err)
		return "", err
	}
	signMsg := ed25519.Sign(priKeyByte, txMsgByte)
	return hex.EncodeToString(signMsg), nil
}

func VerifyEdDSASign(pubKey string, txMsg string, signature string) bool {
	pubKeyByte, _ := hex.DecodeString(pubKey)
	txMsgByte, _ := hex.DecodeString(txMsg)
	signByte, _ := hex.DecodeString(signature)
	return ed25519.Verify(pubKeyByte, txMsgByte, signByte)
}
