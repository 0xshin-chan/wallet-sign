package ethereum

type Eip1559DynamicFeeTx struct {
	ChainId              string `json:"chain_id"`
	Nonce                uint64 `json:"nonce"`
	FromAddress          string `json:"from_address"`
	ToAddress            string `json:"to_address"`
	GasLimit             uint64 `json:"gas_limit"`
	Gas                  uint64 `json:"gas"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`
	Amount               string `json:"amount"`
	ContractAddress      string `json:"contract_address"`
}

/*
{
    "chain_id": "11155111",
	"nonce": 0,
	"from_address": "0x0749F85b38614DcE2ec02b8F0b118A8A235C300b",
	"to_address": "0x02A2AaB257B5d51CE8207be790BDd6168cFB38B5",
	"gas_limit": 21000,
	"gas": 200000,
	"max_fee_per_gas": "327993150328",
	"max_priority_fee_per_gas": "32799315032",
	"amount": "100000000000000000",
	"contract_address": "0x00"
}
*/

type LegacyFeeTx struct {
	ChainId         string `json:"chain_id"`
	Nonce           uint64 `json:"nonce"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	GasLimit        uint64 `json:"gas_limit"`
	GasPrice        uint64 `json:"gas_price"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
}

type EthereumSchema struct {
	RequestId    string              `json:"request_id"`
	DynamicFeeTx Eip1559DynamicFeeTx `json:"dynamic_fee_tx"`
	ClassicFeeTx LegacyFeeTx         `json:"classic_fee_tx"`
}
