package ssm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/log"
)

type EdDSASigner struct{}

func (eddsa *EdDSASigner) CreateKeyPair() (string, string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error("create key pair fail", "err", err)
		return EmptyHexString, EmptyHexString, EmptyHexString, err
	}
	return hex.EncodeToString(privateKey), hex.EncodeToString(publicKey), hex.EncodeToString(publicKey), err
}

func (eddsa *EdDSASigner) SignMessage(priKey string, txMsg string) (string, error) {
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

func (eddsa *EdDSASigner) VerifySignature(pubKey string, txMsg string, signature string) (bool, error) {
	pubKeyByte, _ := hex.DecodeString(pubKey)
	txMsgByte, _ := hex.DecodeString(txMsg)
	signByte, _ := hex.DecodeString(signature)
	return ed25519.Verify(pubKeyByte, txMsgByte, signByte), nil
}
