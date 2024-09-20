package types

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type ChainTip struct {
	Block           string `json:"block"`
	Height          int64  `json:"height"`
	ID              int64  `json:"id"`
	Node            int64  `json:"node"`
	Parent_chaintip *int64 `json:"parent_chaintip,omitempty"`
	Status          string `json:"status"`
}

type BtcFkBlockData struct {
	Method   string     `json:"method"`
	ChainTip []ChainTip `json:"params,omitempty"`
}

type WatchtowerResponse struct {
	Jsonrpc string
	Method  string
	Params  []WatchtowerNotification
}

type WatchtowerTxInput struct {
	Address string
	Amount  uint64
	Txid    string
	Vout    uint32
}

type WatchtowerNotification struct {
	Block            string
	Height           uint64
	Receiving        string
	Satoshis         uint64
	Receiving_txid   string
	Sending_txinputs []WatchtowerTxInput
	Archived         bool
	Receiving_vout   uint64
	Sending          string
	Sending_vout     int32
}

type NyksProposal struct {
	// @Type                string
	Creator             string
	Height              string
	Hash                string
	OrchestratorAddress string
}

type NyksAttestation struct {
	Observed bool
	Votes    []string
	Height   string
	Proposal NyksProposal
}

type NyksAttestaionBlock struct {
	Attestations []NyksAttestation
}

type IndividualTwilightReserveAccount struct {
	TwilightAddress string
	BtcValue        string
}

type NyksSweepProposal struct {
	ReserveAddress                   string
	JudgeAddress                     string
	TotalValue                       string
	IndividualTwilightReserveAccount []IndividualTwilightReserveAccount
	BtcRefundTx                      string
	BtcSweepTx                       string
}

type NyksAttestationSweep struct {
	Observed bool
	Votes    []string
	Height   string
	Proposal NyksSweepProposal
}

type NyksAttestaionBlockSweep struct {
	Attestations []NyksAttestationSweep
}

type DepositAddress struct {
	BtcDepositAddress           string
	TwilightAddress             string
	TwilightStakingAmount       string
	BtcSatoshiTestAmount        string
	IsConfirmed                 bool
	CreationTwilightBlockHeight string
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
	ZkOracleAddress  string
}

type DelegateAddressesResp struct {
	Addresses []DelegateAddress
}

type ErrorResp struct {
	Code    int
	Message string
}

type SweepAddress struct {
	Address         string
	Script          []byte
	Preimage        []byte
	Unlock_height   int64
	Parent_address  string
	Signed_refund   bool
	Signed_sweep    bool
	Archived        bool
	BroadcastSweep  bool
	BroadcastRefund bool
	Owned           bool
}

type MultiSigAddress struct {
	Address    string
	Script     string
	EthAddress string
	Signed     bool
	Archived   bool
}

type SignedTx struct {
	Tx           string
	UnlockHeight int64
}

type UnsignedTx struct {
	Tx        string
	ReserveId int64
	RoundId   int64
}

type Utxo struct {
	Txid   string
	Vout   uint32
	Amount uint64
}

type MsgSignSweep struct {
	SignerPublicKey string
	SweepSignature  []string
	SignerAddress   string
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
	SignerPublicKey string
	RefundSignature []string
	SignerAddress   string
}

type MsgSignRefundResp struct {
	SignRefundMsg []MsgSignRefund
}

type BtcReserveResp struct {
	BtcReserves []BtcReserve
}

type BtcReserve struct {
	ReserveId             string
	ReserveAddress        string
	JudgeAddress          string
	BtcRelayCapacityValue string
	TotalValue            string
	PrivatePoolValue      string
	PublicValue           string
	FeePool               string
	UnlockHeight          string
	RoundId               string
}

type UnsignedTxSweepResp struct {
	UnsignedTxSweepMsg  UnsignedTxSweep
	UnsignedTxSweepMsgs []UnsignedTxSweep
	Code                int
}

type UnsignedTxSweep struct {
	TxId               string
	BtcUnsignedSweepTx string
	JudgeAddress       string
	RoundId            string
	ReserveId          string
}

type UnsignedTxRefundResp struct {
	UnsignedTxRefundMsg  UnsignedTxRefund
	UnsignedTxRefundMsgs []UnsignedTxRefund
	Code                 int
}

type UnsignedTxRefund struct {
	TxId                string
	BtcUnsignedRefundTx string
	JudgeAddress        string
	RoundId             string
	ReserveId           string
}

type WatchedTxs struct {
	Txs []WatchedTx
}

type WatchedTx struct {
	Txid    string
	Address string
	Reserve uint16
	Round   uint16
	Watched bool
}

// type ProposedAddress struct {
// 	Current      string
// 	Proposed     string
// 	UnlockHeight int64
// }

type FeeLimits struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type FeeRate struct {
	Limits   FeeLimits `json:"limits"`
	Regular  int       `json:"regular"`
	Priority int       `json:"priority"`
}

type ProposedAddressesResp struct {
	ProposeSweepAddressMsgs []ProposedAddress
}

type ProposedAddressResp struct {
	ProposeSweepAddressMsg ProposedAddress
}

type ProposedAddress struct {
	BtcAddress   string `json:"btcAddress"`
	BtcScript    string `json:"btcScript"`
	ReserveId    string `json:"reserveId"`
	RoundId      string `json:"roundId"`
	JudgeAddress string `json:"judgeAddress"`
}

type BlockResultsResponse struct {
	Result *BlockResult `json:"result,omitempty"`
	Error  *RPCError    `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type BlockResult struct {
	Height           string     `json:"height"`
	TxsResults       []TxResult `json:"txs_results"`
	BeginBlockEvents []Event    `json:"begin_block_events"`
	EndBlockEvents   []Event    `json:"end_block_events"`
}

type TxResult struct {
	Events []Event `json:"events"`
}

type Event struct {
	Type string `json:"type"`
}

type RefundAccount struct {
	Amount                      string `json:"Amount"`
	BtcDepositAddress           string `json:"BtcDepositAddress"`
	BtcDepositAddressIdentifier int    `json:"BtcDepositAddressIdentifier"`
}

// RefundTxSnapshot represents the main structure
type RefundTxSnapshot struct {
	ReserveId                string          `json:"ReserveId"`
	RoundId                  string          `json:"RoundId"`
	RefundAccounts           []RefundAccount `json:"refundAccounts"`
	EndBlockerHeightTwilight string          `json:"EndBlockerHeightTwilight"`
}

type RefundTxSnapshotResp struct {
	RefundTxSnapshot RefundTxSnapshot
}

// WithdrawRequest represents a withdrawal request details
type WithdrawRequest struct {
	WithdrawIdentifier int    `json:"withdrawIdentifier"`
	WithdrawAddress    string `json:"withdrawAddress"`
	WithdrawAmount     string `json:"withdrawAmount"`
}

// ReserveWithdrawSnapshot represents the main structure
type ReserveWithdrawSnapshot struct {
	ReserveId                string            `json:"ReserveId"`
	RoundId                  string            `json:"RoundId"`
	WithdrawRequests         []WithdrawRequest `json:"withdrawRequests"`
	EndBlockerHeightTwilight string            `json:"EndBlockerHeightTwilight"`
}

type ReserveWithdrawSnapshotResp struct {
	ReserveWithdrawSnapshot ReserveWithdrawSnapshot
}

type BroadcastRefundMsg struct {
	ReserveId      string `json:"reserveId"`
	RoundId        string `json:"roundId"`
	SignedRefundTx string `json:"signedRefundTx"`
	JudgeAddress   string `json:"judgeAddress"`
}

type BroadcastRefundMsgResp struct {
	BroadcastRefundMsg BroadcastRefundMsg
}

type ProposeSweepAddressMsg struct {
	BtcAddress   string `json:"btcAddress"`
	BtcScript    string `json:"btcScript"`
	ReserveId    string `json:"reserveId"`
	RoundId      string `json:"roundId"`
	JudgeAddress string `json:"judgeAddress"`
}

type ProposeSweepAddressMsgResp struct {
	ProposeSweepAddressMsgs []ProposeSweepAddressMsg `json:"proposeSweepAddressMsgs"`
}

type Fragment struct {
	FragmentId           string   `json:"FragmentId"`
	FragmentStatus       bool     `json:"FragmentStatus"`
	JudgeAddress         string   `json:"JudgeAddress"`
	JudgeStatus          bool     `json:"JudgeStatus"`
	Signers              []Signer `json:"Signers"`
	SignerApplicationFee string   `json:"SignerApplicationFee"`
	Threshold            string   `json:"Threshold"`
	FeePool              string   `json:"FeePool"`
	FragmentFeeBips      string   `json:"FragmentFeeBips"`
	ArbitraryData        string   `json:"arbitraryData"`
	ReserveIds           []string `json:"ReserveIds"`
}

type Signer struct {
	FragmentID           string `json:"FragmentID"`
	SignerAddress        string `json:"SignerAddress"`
	SignerStatus         bool   `json:"SignerStatus"`
	SignerBtcPublicKey   string `json:"SignerBtcPublicKey"`
	SignerApplicationFee string `json:"SignerApplicationFee"`
	SignerFeeBips        string `json:"SignerFeeBips"`
}

type Fragments struct {
	Fragments []Fragment `json:"Fragments"`
}

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("error in pushing to refund tx channel: ", err)
			return
		}
	}
}
