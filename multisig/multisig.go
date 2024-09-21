package multisig

import (
	"database/sql"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/judge"
)

func ProcessMultisigAddressGeneration(AccountName string, judgeAddr string, dbconn *sql.DB, btcPubKey string, clientEthAddress string, fragmentId int, ethAccount accounts.Account) string {
	return address.GenerateAndRegisterNewBtcMultiSigAddress(dbconn, AccountName, btcPubKey, judgeAddr, fragmentId, clientEthAddress, ethAccount)
}

func ProcessMultisigWithdraw(withDrawBtcAddr string, clientEthAddr string, accountName string, dbconn *sql.DB, ethAccount accounts.Account) string {
	return judge.GenerateMultisigwithdrawTx(withDrawBtcAddr, clientEthAddr, accountName, dbconn, ethAccount)
}
func ProcesSignWithdrawPSBT(psbt string) string {
	return judge.SignMultisigPSBT(psbt)
}
