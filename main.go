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
	if accountName == "validator-sfo" {
		judge = true
	}

	time.Sleep(30 * time.Second)

	go orchestrator(accountName, forkscanner_url)

	initJudge(accountName)

	time.Sleep(1 * time.Minute)
	if judge == true {
		go startJudge(accountName)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url)
	startTransactionSigner(accountName)

	//for local testing
	// x := "01000000000102295ba36c785e2555c91c44e4bb4d219b8a81dcc5e17f6469d2e1e09cff964d8b0000000000f5ffffff6e367aba70e2c45a3887e90027a2711665d69e41e8f125fa40d30a60f9cb42800000000000f5ffffff01e8fd0000000000002200208b2d3aed5d37ed05f3f78eafcd611f2daa4878c7ad933da45d45cea2cb3e6a4f0720b346bca7b2c3e4cf824a16a71c94a1758a8597a0f11521bd792ee9722893a365004830450221009d799f2f4b7d0c9c11df5e2a89fa588e2bfe49b0d526a9e24dbdaae7aecc40f9022009087b64c3ac497bf32ecaaf5cf04430c55d0522b200db200213c077f65476bc014730440220478defb6b0901b7ca959e6f7e0180b67905e28f00b6e2e2ceb1de5fbd7d22191022071317493e65d3c76c186decad19792455c28e5384fc018c5dea838092cfa50a001483045022100eaade6964ed8fd9010cd859114b97901f141c6dbac64f1cd0b819bd512cf6a4b02206d9e4cde71ceafc410e675a9d36ddca7ed020b7d1297e58e8ebf637dc25e2ec301483045022100d0f20aa1a26f3a4a42e533dbdae907d14a6862f970a92e5b1e95561c027c2ab402204dc7f4fdecbcb6bfec81171e1d08604399fc2f907a2dab385c28846589ee32f601fd1801542102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f92721038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe156ae6382012088a91442fa7057aa1b3a5c11bb0cecae8ec94e0a5326cc87672102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055ac6403e74b0cb27568680720b346bca7b2c3e4cf824a16a71c94a1758a8597a0f11521bd792ee9722893a36500473044022007a4555de5317266240ba042e89cf69358ce6cf2a27adb6badfbd9ce8d14b8aa022065a69590190931db33cc05f8b91aa24ea69ee1c9e241e6c83057cf43e941c3df01483045022100837d2de0d12ffaa41c2147e470acd7530d66524e1b695e48d9b81df97dcd6436022004e69ce4875d2dddfebeebc24dd1949ac0ed617d370ca2500c2eea4c532d101801483045022100f3469e65423b74deb1972523c27a75b45fe2c1e4684c385219997e1a2b98bf71022029e58a23e6666507eaf0d2413da5d14b326f50621f38c6b94bd870c6b6fe7c7a0147304402201ff9e45680a70aa3db5a75e141eafacb377d777cd9e5b3305aac6fbeae0e70b80220191605c81d3d716a0fcdf6a097dc03856e32873613b7694f427205b6779566a401fd1801542102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f92721038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe156ae6382012088a91442fa7057aa1b3a5c11bb0cecae8ec94e0a5326cc87672102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055ac6403e74b0cb275686800000000"
	// tx, _ := createTxFromHex(x)

	// replaceSignatureInWitness(tx, 0, "3045022100c18e4ade8a23871fa915b4e400ca8826b6905ac6e253fb64b60d9ded62533cb1022079400a0b7b4aedb7beff35d69d297ca22fe144b0b6c82ae3d395612c031ebc0801")
	// replaceSignatureInWitness(tx, 1, "3044022019111a0663159c1b345b8b6832e2b522d102810933565494b7b3ae45b347823002205caca6cb8bc45953505a9bfd7d21617bf06b997a2a21bff1e873a07a7d180cf201")

	// var UnsignedTx bytes.Buffer
	// tx.Serialize(&UnsignedTx)
	// hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	// fmt.Println(hexTx)

}

// func replaceSignatureInWitness(tx *wire.MsgTx, inputIndex int, newSignature string) {
// 	// Check if the input index is valid
// 	if inputIndex < 0 || inputIndex >= len(tx.TxIn) {
// 		log.Fatalf("Invalid input index: %d", inputIndex)
// 	}

// 	witness := tx.TxIn[inputIndex].Witness
// 	if len(witness) < 3 { // Assuming P2WSH multisig format: [sig1, sig2, witnessScript]
// 		log.Fatalf("Not enough elements in witness to replace the second signature.")
// 	}

// 	// Replace the second signature
// 	x, _ := hex.DecodeString(newSignature)
// 	witness[5] = witness[4]
// 	witness[4] = witness[3]
// 	witness[3] = witness[2]
// 	witness[2] = x

// 	// Reassign the modified witness
// 	tx.TxIn[inputIndex].Witness = witness
// }
