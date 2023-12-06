package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"
)

var dbconn *sql.DB
var masterPrivateKey *bip32.Key
var judge bool
var oracleAddr string
var valAddr string

func initialize() {
	initConfigFile()
	btcPubkey := initWallet()
	dbconn = initDB()
	setDelegator(btcPubkey)
}

func main() {

	initialize()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	fmt.Println("account name : ", accountName)
	var forkscanner_host = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	forkscanner_url := url.URL{Scheme: "ws", Host: forkscanner_host, Path: "/"}
	if accountName == "validator-sfo" || accountName == "validator-ams" {
		judge = true
	}

	time.Sleep(30 * time.Second)
	go orchestrator(accountName, forkscanner_url)

	initJudge(accountName)

	time.Sleep(1 * time.Minute)
	if judge == true {
		go startJudge(accountName)
	} else {
		time.Sleep(2 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url)
	startTransactionSigner(accountName)

	// for local testing
	// by, _ := hex.DecodeString("1308ae55ac15828ba3d412458689fae8848bd93f509c9bbb5024a43009cbfd6a")
	// p := hash160(by)

	// hex := hex.EncodeToString(p)

	// fmt.Println(hex)

	// x := "01000000000102f47005f4fd24885296489de53c030d5b3d2d1ea6b0fc2ae0817234ee72bfc39c000000000052760c00409b7df603ee88fa8aa60e0ee77ebb753faf73d803648fa2ce7869f767685b2f000000000052760c000162520000000000002200207ca749ba27919101ca2b9db782579c4f516e266ebaea1a41e8631435d290066c07201308ae55ac15828ba3d412458689fae8848bd93f509c9bbb5024a43009cbfd6a0047304402200f553b074c61a8c13938446bac012c320cdd196cc3b05680e1b79f2688a4020a02204a39bc0293b31774ab62e0f2a357d62117462f6784b2cd690df5626bc23edf3b0147304402206a0e52093403902c8354b8e9101555c5855e0eff1c515443772c70dac2a145b2022057ccd29a3fe89443433105b1f27b3fadebf205aea9deb815be01f6c32a6ca50f01473044022032b13f11fbc28a1126b7178abdfab1ba6e7e35df30ada5440e40d67e49bae5be02203e626b1ca162e5278f1f2401cc99b2b639f40bee3c8ae09c6c0f52d808067eb30148304502210086894133b0117db50d7ac931cd716c72c919ba4a072c06ff5e814b4a9f6159ff022037fda668f22d20d81732a601c0d138913c64217efa2c124ef939e60c044fd30801fd1e010367760cb175542102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f92721038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe156af82012088a9148fee29a758b609247de423a6d454f5b4749f860d87736421033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f927ac640372760cb275686807201308ae55ac15828ba3d412458689fae8848bd93f509c9bbb5024a43009cbfd6a00473044022028bca8b2020976202e1132022e453bacda0bde9cf2d38dc99db96de0b5a11fff02204c6a588be39948b92b667dade7017d422117ee6017a5ac451f39fe39249e32b70147304402203dde1371722d34056c0c26e2169af4fe5d6a9a840edaa246ddaf05828f442aad0220040534f6a7c255b93adf158b636b3c5a55f15987e2103007641faf2274125bb401483045022100dcd3635e74b91800cac17e0282131e129518fc94fb81b2753e9f127ef7d748e60220287b725e52b57cdc9269bf7fe0c3ae321d57ae36dc88638bef48d46ef07d9fb301483045022100b85a0c83e8fa4cfd5e90706f42a8ee8df4682837568df9c3348b39626e79045d022066ec28cddbd2e3b4361c98a787089023ec46ed67c6f71b25c2f409b8ec7a7a1801fd1e010367760cb175542102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f92721038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe156af82012088a9148fee29a758b609247de423a6d454f5b4749f860d87736421033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f927ac640372760cb275686857760c00"

	// tx, _ := createTxFromHex(x)
	// tx.LockTime = uint32(816783)

	// var UnsignedTx bytes.Buffer
	// tx.Serialize(&UnsignedTx)
	// hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	// fmt.Println("transaction UnSigned Sweep: ", hexTx)

}
