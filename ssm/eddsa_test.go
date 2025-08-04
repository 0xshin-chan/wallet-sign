package ssm

import (
	"fmt"
	"testing"
)

func TestCreateEdDSAKeyPair(t *testing.T) {
	priKey, pubKey, _ := CreateEdDSAKeyPair()
	fmt.Println(priKey)
	fmt.Println(pubKey)
	//7131e0b7d9fa7f831acc4a2d6069cb0d874c10c58dd1d0b829de8a2599012ede8f22cbe3681ac36d15be83a37db91fac0d663ab21dce086c8ac090cd0ff5970d
	//8f22cbe3681ac36d15be83a37db91fac0d663ab21dce086c8ac090cd0ff5970d
}
