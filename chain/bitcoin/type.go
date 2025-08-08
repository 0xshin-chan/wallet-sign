package bitcoin

type Vin struct {
	Hash   string `json:"hash"`
	Index  uint64 `json:"index"`
	Amount uint64 `json:"amount"`
}

type Vout struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Index   uint64 `json:"index"`
}

type BitcoinSchema struct {
	RequestId string `json:"requestId"`
	Fee       string `json:"fee"`
	Vins      []Vin  `json:"vins"`
	Vouts     []Vout `json:"vouts"`
}
