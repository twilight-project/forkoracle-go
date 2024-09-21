package eventhandler

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/rpc"

	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/json"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/comms"
	"github.com/twilight-project/forkoracle-go/judge"
	"github.com/twilight-project/forkoracle-go/multisig"
	"github.com/twilight-project/forkoracle-go/store"
	"github.com/twilight-project/forkoracle-go/transaction_signer"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
)

var (
	ContractAddress string
	Dbconn          *sql.DB
	AccountName     string
	JudgeAddr       string
	EthAccount      accounts.Account
)

func NyksEventListener(event string, accountName string, functionCall string, dbconn *sql.DB,
	oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub, latestRefundTxHash *prometheus.GaugeVec) {
	headers := make(map[string][]string)
	headers["Content-Type"] = []string{"application/json"}
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_socket_url"))
	conn, _, err := websocket.DefaultDialer.Dial(nyksd_url, headers)
	if err != nil {
		fmt.Println("nyks event listerner dial:", err)
	}
	defer conn.Close()

	// Set up ping/pong connection health check
	pingPeriod := 30 * time.Second
	pongWait := 60 * time.Second
	stopChan := make(chan struct{}) // Create the stop channel

	err = conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		fmt.Println("error setting read deadline: ", err)
	}
	conn.SetPongHandler(func(string) error {
		err = conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			fmt.Println("error setting read deadline: ", err)
		}
		return nil
	})

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-stopChan: // Listen to the stop channel
				return
			}
		}
	}()

	payload := `{
        "jsonrpc": "2.0",
        "method": "subscribe",
        "id": 0,
        "params": {
            "query": "tm.event='Tx' AND message.action='%s'"
        }
    }`
	payload = fmt.Sprintf(payload, event)

	err = conn.WriteMessage(websocket.TextMessage, []byte(payload))
	if err != nil {
		fmt.Println("error in nyks event handler: ", err)
		stopChan <- struct{}{} // Signal goroutine to stop
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error in nyks event handler: ", err)
			stopChan <- struct{}{} // Signal goroutine to stop
			return
		}

		// var event Event
		// err = json.Unmarshal(message, &event)
		// if err != nil {
		// 	fmt.Println("error unmarshalling event: ", err)
		// 	continue
		// }

		// fmt.Print("event : ", event)
		// fmt.Print("event : ", message)

		// if event.Method == "subscribe" && event.Params.Query == fmt.Sprintf("tm.event='Tx' AND message.action='%s'", event) {
		// 	continue
		// }

		switch functionCall {
		case "signed_sweep_process":
			go judge.ProcessSignedSweep(accountName, oracleAddr, dbconn)
		case "refund_process":
			go judge.ProcessRefund(accountName, oracleAddr, dbconn)
		case "signed_refund_process":
			go judge.ProcessSignedRefund(accountName, oracleAddr, dbconn, WsHub, latestRefundTxHash)
		case "register_res_addr_validators":
			go address.RegisterAddressOnValidators(dbconn)
		case "register_res_addr_signers":
			go address.RegisterAddressOnSigners(dbconn)
		case "signing_sweep":
			go transaction_signer.ProcessTxSigningSweep(accountName, dbconn, oracleAddr)
		case "signing_refund":
			go transaction_signer.ProcessTxSigningRefund(accountName, dbconn, oracleAddr)
		case "sweep_process":
			go judge.ProcessSweep(accountName, dbconn, oracleAddr)
		default:
			log.Println("Unknown function :", functionCall)
		}
	}
}

func RegistertoEthEvents(contractAddress string, dbconn *sql.DB, accountName string, judgeAddr string, ethAccount accounts.Account) {
	client := comms.GetEthWSSClient()
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		fmt.Println("Failed to subscribe to contract logs: %v", err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(store.StoreABI)))
	if err != nil {
		fmt.Println("Failed to parse contract ABI: %v", err)
	}

	var fragment btcOracleTypes.Fragment
	fragments := comms.GetAllFragments()
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
		}
	}
	fragmentId, _ := strconv.Atoi(fragment.FragmentId)

	for {
		select {
		case vLog := <-logs:
			switch vLog.Topics[0].Hex() {
			case crypto.Keccak256Hash([]byte("AddressRequested(bytes)")).Hex():
				event := new(struct {
					BitcoinPublicKey []byte
				})
				err := contractAbi.UnpackIntoInterface(event, "AddressRequested", vLog.Data)
				if err != nil {
					fmt.Println("Failed to unpack AddressRequested event data: %v", err)
				}
				fmt.Printf("AddressRequested event emitted, BitcoinPublicKey: %s\n", event.BitcoinPublicKey)

				multisig.ProcessMultisigAddressGeneration(accountName, judgeAddr, dbconn, hex.EncodeToString(event.BitcoinPublicKey), vLog.Address.Hex(), fragmentId, ethAccount)

			case crypto.Keccak256Hash([]byte("WithdrawalRequest(string)")).Hex():
				event := new(struct {
					HexAddress string
				})
				err := contractAbi.UnpackIntoInterface(event, "WithdrawalRequest", vLog.Data)
				if err != nil {
					fmt.Println("Failed to unpack WithdrawalRequest event data: %v", err)
				}
				fmt.Printf("WithdrawalRequest event emitted, HexAddress: %s\n", event.HexAddress)
				multisig.ProcessMultisigWithdraw(event.HexAddress, vLog.Address.Hex(), accountName, dbconn, ethAccount)
			}

		case err := <-sub.Err():
			fmt.Println("Received subscription error: %v", err)
		}
	}
}

type Server struct{}

type BtcPubkeyArgs struct {
	BTCPubKey string
	EthAddr   string
}

type GetUnsignedPsbtArgs struct {
	EthAddr         string
	WithdrawBtcAddr string
}

type SubmitSignedPSBT struct {
	Psbt string
}

func RpcServer(contractAddress string, dbconn *sql.DB, accountName string, judgeAddr string, ethAccount accounts.Account) {
	judgeAddr = judgeAddr
	ContractAddress = contractAddress
	Dbconn = dbconn
	AccountName = accountName
	EthAccount = ethAccount

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	rpc := new(Server)
	rpcServer.RegisterService(rpc, "rpc")

	router := mux.NewRouter()
	router.Handle("/rpc", rpcServer)
	fmt.Println("rpc server started")
	http.ListenAndServe("0.0.0.0:1234", router)
}

func (s *Server) SubmitBtcPubkey(r *http.Request, args *BtcPubkeyArgs, reply *string) error {
	fmt.Println("inside submit btc pubkey")
	if args.BTCPubKey == "" {
		*reply = ""
		return nil
	}
	if args.EthAddr == "" {
		*reply = ""
		return nil
	}
	var fragment btcOracleTypes.Fragment
	fragments := comms.GetAllFragments()
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == JudgeAddr {
			fragment = f
		}
	}
	fragmentId, _ := strconv.Atoi(fragment.FragmentId)

	newAddress := multisig.ProcessMultisigAddressGeneration(AccountName, JudgeAddr, Dbconn, args.BTCPubKey, args.EthAddr, fragmentId, EthAccount)
	*reply = newAddress
	return nil
}

func (s *Server) GetUnsignedPsbt(args *GetUnsignedPsbtArgs, reply *string) error {
	// Here you can add your logic to get the unsigned PSBT
	// For now, it just returns an empty string
	fmt.Println("inside get unsigned psbt")
	if args.EthAddr == "" {
		*reply = "no eth address submitted"
	}
	if args.WithdrawBtcAddr == "" {
		*reply = "no withdraw btc address submitted"
	}
	psbt := multisig.ProcessMultisigWithdraw(args.WithdrawBtcAddr, args.EthAddr, AccountName, Dbconn, EthAccount)
	*reply = psbt
	return nil
}

func (s *Server) SubmitSignedPsbt(args *SubmitSignedPSBT, reply *string) error {
	fmt.Println("inside submit signed psbt")
	// Here you can add your logic to process the signed PSBT
	// For now, it just returns true if the PSBT is not empty
	if args.Psbt == "" {
		*reply = "no psbt submitted"
	}
	psbt := multisig.ProcesSignWithdrawPSBT(args.Psbt)
	*reply = psbt
	return nil
}
