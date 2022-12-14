package main

type ChainTip struct {
	Block           string `json:"block"`
	Height          int64  `json:"height"`
	ID              int64  `json:"id"`
	Node            int64  `json:"node"`
	Parent_chaintip *int64 `json:"parent_chaintip,omitempty"`
	Status          string `json:"status"`
}

type BlockData struct {
	Method   string     `json:"method"`
	ChainTip []ChainTip `json:"params,omitempty"`
}

type WatchtowerResponse struct {
	Jsonrpc string
	Method  string
	Params  []WatchtowerNotification
}

type WatchtowerNotification struct {
	Block     string
	Receiving string
	Satoshis  uint64
	Height    uint64
	Txid      string
	archived  bool
}

type Proposal struct {
	// @Type                string
	Creator             string
	Height              string
	Hash                string
	OrchestratorAddress string
}

type Attestation struct {
	Observed bool
	Votes    []string
	Height   string
	Proposal Proposal
}

type AttestaionBlock struct {
	Attestations []Attestation
}

type DepositAddress struct {
	depositAddress         string
	twilightDepositAddress string
}

type QueryDepositAddressResp struct {
	addresses []DepositAddress
}

type ConfirmDepositMessage struct {
	DepositAddress         string
	DepositAmount          uint64
	Height                 uint64
	Hash                   string
	TwilightDepositAddress string
	BtcOracleAddress       string
}
