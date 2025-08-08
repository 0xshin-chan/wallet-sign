package solana

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	//"github.com/gagliardetto/solana-go"

	"github.com/cosmos/btcutil/base58"
)

func PubKeyToAddress(publicKey *ed25519.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key is nil")
	}
	return base58.Encode(*publicKey), nil
}

func PubKeyHexToPubKey(publicKeyHex string) (*ed25519.PublicKey, error) {
	pubKeyByteList, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key hex: %w", err)
	}
	pubKey := ed25519.PublicKey(pubKeyByteList)
	return &pubKey, nil
}

func PubKeyHexToAddress(publicKeyHex string) (string, error) {
	pubKey, err := PubKeyHexToPubKey(publicKeyHex)
	if err != nil {
		return "", err
	}
	return PubKeyToAddress(pubKey)
}
