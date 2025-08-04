package ssm

import (
	"fmt"
	"testing"
)

func TestCreateESDSAKeyPair(t *testing.T) {
	key, pubKey, compressPubKey, _ := CreateECDSAKeyPair()
	fmt.Println(key)
	fmt.Println(pubKey)
	fmt.Println(compressPubKey)
	//646448df201c4cdc805a3271de8e5d951ac4cb83ccb88b8cbc31540d1cdd7fd0
	//04a3cdc9ec7962a2531a65196830e5c322239fcfbc36fad2db1a577cc70ae6b90817dad1f288b35f9d8754190c1521bf9b7119217497687f0e62f8adac4c69e7d6
	//02a3cdc9ec7962a2531a65196830e5c322239fcfbc36fad2db1a577cc70ae6b908
}

func TestSignECDSAMessage(t *testing.T) {
	priKey := "646448df201c4cdc805a3271de8e5d951ac4cb83ccb88b8cbc31540d1cdd7fd0"
	txMsg := "0x3e4f9a460233ec33862da1ac3dabf5b32db01400fba166cdec40ad6dc735b4ab"
	signature, err := SignECDSAMessage(priKey, txMsg)
	if err != nil {
		fmt.Println("sign tx fail")
	}
	fmt.Println("signature = ", signature)
	//5084894c9700ac4041603cc7dd61f0dd2203af310fff6cb9af33d982332890e047b0ca113461967ad0ec89fdcb6f6c88baf56359e31ff13ac2f208b16c147b9101
}

func TestVerifyECDSAMessage(t *testing.T) {
	CompressPubKey := "04a3cdc9ec7962a2531a65196830e5c322239fcfbc36fad2db1a577cc70ae6b90817dad1f288b35f9d8754190c1521bf9b7119217497687f0e62f8adac4c69e7d6"
	txHash := "3e4f9a460233ec33862da1ac3dabf5b32db01400fba166cdec40ad6dc735b4ab"
	signature := "5084894c9700ac4041603cc7dd61f0dd2203af310fff6cb9af33d982332890e047b0ca113461967ad0ec89fdcb6f6c88baf56359e31ff13ac2f208b16c147b9101"

	isValid, err := VerifyEcdsaSignature(CompressPubKey, txHash, signature)
	if err != nil {
		t.Error("failed to verify signature")
	}

	if !isValid {
		t.Error("signature is invalid")
	}
}
