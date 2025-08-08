package ethereum

import (
	"encoding/hex"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

func BuildErc20Data(toAddress common.Address, amount *big.Int) []byte {
	var data []byte

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodId := hash[:4]
	dataAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataAmount := common.LeftPadBytes(amount.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataAddress...)
	data = append(data, dataAmount...)

	return data
}

func BuildErc721Data(fromAddress, toAddress common.Address, tokenId *big.Int) []byte {
	var data []byte

	transferFnSignature := []byte("safeTransferFrom(address,address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodId := hash[:4]

	dataFromAddress := common.LeftPadBytes(fromAddress.Bytes(), 32)
	dataToAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataTokenId := common.LeftPadBytes(tokenId.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataFromAddress...)
	data = append(data, dataToAddress...)
	data = append(data, dataTokenId...)

	return data
}

func CreateLegacyUnSignTx(txData *types.LegacyTx, chainId *big.Int) string {
	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)
	txHash := signer.Hash(tx)
	return txHash.String()
}

func CreateEip1559UnSignTx(txData *types.DynamicFeeTx, chainId *big.Int) (string, error) {
	// 构建新的 EIP-1559 类型交易对象，如果时legacy会自己选择不需要改代码
	tx := types.NewTx(txData)
	// 签名者，根据链 ID 获取使用的签名规则（EIP-1559）
	signer := types.LatestSignerForChainID(chainId)
	// 获取未签名交易的 hash，对交易进行编码 + 加入 EIP-155 签名规则，计算未签名的哈希
	txHash := signer.Hash(tx)
	return txHash.String(), nil
}

func CreateLegacySignedTx(txData *types.LegacyTx, signature []byte, chainId *big.Int) (string, string, error) {
	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return "", "", errors.New("tx with signature fail")
	}
	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", "", errors.New("encode tx to byte fail")
	}
	return "0x" + hex.EncodeToString(signedTxData), signedTx.Hash().String(), nil
}

func CreateEip1559SignedTx(txData *types.DynamicFeeTx, signature []byte, chainId *big.Int) (types.Signer, *types.Transaction, string, string, error) {
	// 创建新的交易对象
	tx := types.NewTx(txData)
	// 获取签名器（根据链 ID）
	signer := types.LatestSignerForChainID(chainId)
	// 把签名（65 字节）附加到交易中，得到一个已签名的交易对象
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, nil, "", "", errors.New("tx with signature fail")
	}
	// 将已签名交易进行 RLP 编码（以太坊要求的交易格式）
	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, nil, "", "", errors.New("encode tx to byte fail")
	}
	// 返回签名器、已签名交易、最终交易的 RLP 编码（去掉前 2 个字节）和 tx hash
	return signer, signedTx, "0x" + hex.EncodeToString(signedTxData)[4:], signedTx.Hash().String(), nil
}
