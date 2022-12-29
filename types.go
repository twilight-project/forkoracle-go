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
	Block          string
	Height         uint64
	Receiving      string
	Satoshis       uint64
	Receiving_txid string
	Sending        string
	Archived       bool
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
	DepositAddress         string
	TwilightDepositAddress string
}

type QueryDepositAddressResp struct {
	Addresses []DepositAddress
}

type ConfirmDepositMessage struct {
	DepositAddress         string
	DepositAmount          uint64
	Height                 uint64
	Hash                   string
	TwilightDepositAddress string
	BtcOracleAddress       string
}
