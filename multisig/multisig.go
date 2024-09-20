package multisig

import (
	"database/sql"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/judge"
)

func ProcessMultisigAddressGeneration(AccountName string, judgeAddr string, dbconn *sql.DB, btcPubKey string, clientEthAddress string, fragmentId int, ethAccount accounts.Account) {
	address.GenerateAndRegisterNewBtcMultiSigAddress(dbconn, AccountName, btcPubKey, judgeAddr, fragmentId, clientEthAddress, ethAccount)
}

func ProcessMultisigWithdraw(withdrawBTCAddress string, clientEthAddr string, accountName string, dbconn *sql.DB, ethAccount accounts.Account) {
	judge.GenerateMultisigwithdrawTx(withdrawBTCAddress, clientEthAddr, accountName, dbconn, ethAccount)
}
