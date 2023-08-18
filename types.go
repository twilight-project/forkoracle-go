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

type WatchtowerSender struct {
	Address string
	Amount  uint64
	Txid    string
	vout    uint32
}

type WatchtowerNotification struct {
	Block            string
	Height           uint64
	Receiving        string
	Satoshis         uint64
	Receiving_txid   string
	Sending_txinputs []WatchtowerSender
	Archived         bool
	Receiving_vout   uint64
	Sending          string
	Sending_vout     int32
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

type IndividualTwilightReserveAccount struct {
	TwilightAddress string
	BtcValue        string
}

type SweepProposal struct {
	ReserveAddress                   string
	JudgeAddress                     string
	TotalValue                       string
	IndividualTwilightReserveAccount []IndividualTwilightReserveAccount
	BtcRefundTx                      string
	BtcSweepTx                       string
}

type AttestationSweep struct {
	Observed bool
	Votes    []string
	Height   string
	Proposal SweepProposal
}

type AttestaionBlockSweep struct {
	Attestations []AttestationSweep
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

type DelegateAddress struct {
	ValidatorAddress string
	BtcOracleAddress string
	BtcPublicKey     string
}

type DelegateAddressesResp struct {
	Addresses []DelegateAddress
}

type ErrorResp struct {
	Code    int
	Message string
}

type SweepAddress struct {
	Address        string
	Script         []byte
	Preimage       []byte
	Unlock_height  int64
	Parent_address string
	Signed_refund  bool
	Signed_sweep   bool
}

type Utxo struct {
	Txid   string
	Vout   uint32
	Amount uint64
}

type BtcWithdrawRequestResp struct {
	WithdrawRequest []BtcWithdrawRequest
}

type BtcWithdrawRequest struct {
	WithdrawAddress string
	ReserveAddress  string
	WithdrawAmount  string
	TwilightAddress string
}

type MsgSignSweep struct {
	ReserveAddress   string
	SignerAddress    string
	SweepSignature   string
	BtcOracleAddress string
}

type MsgSignSweepResp struct {
	SignSweepMsg []MsgSignSweep
}

type ReserveAddress struct {
	ReserveScript  string
	ReserveAddress string
	JudgeAddress   string
}

type ReserveAddressResp struct {
	Addresses []ReserveAddress
}

type RegisteredJudge struct {
	Creator          string
	JudgeAddress     string
	ValidatorAddress string
}

type RegisteredJudgeResp struct {
	Judges []RegisteredJudge
}

type MsgSignRefund struct {
	ReserveAddress   string
	SignerAddress    string
	RefundSignature  string
	BtcOracleAddress string
}

type MsgSignRefundResp struct {
	SignRefundMsg []MsgSignRefund
}

type BtcReserveResp struct {
	BtcReserves []BtcReserve
}

type BtcReserve struct {
	ReserveId             uint8
	ReserveAddress        string
	JudgeAddress          string
	BtcRelayCapacityValue uint8
	TotalValue            uint8
	PrivatePoolValue      uint8
	PublicValue           uint8
	FeePool               uint8
}

type UnsignedTxSweepResp struct {
	UnsignedTxSweepMsgs []UnsignedTxSweep
}

type UnsignedTxSweep struct {
	TxId               string
	BtcUnsignedSweepTx string
	JudgeAddress       string
}

type UnsignedTxRefundResp struct {
	UnsignedTxRefundMsgs []UnsignedTxRefund
}

type UnsignedTxRefund struct {
	TxId                string
	BtcUnsignedRefundTx string
	JudgeAddress        string
}
