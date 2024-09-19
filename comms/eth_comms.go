package comms

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"github.com/twilight-project/forkoracle-go/store"
)

func getEthClient() *ethclient.Client {
	ethUrl := viper.GetString("eth_url")
	client, err := ethclient.Dial(ethUrl)
	if err != nil {
		fmt.Println("Failed to connect to the Ethereum client: %v", err)
		return nil
	}
	return client
}

func CallContractFunc(account accounts.Account, contractAddress string) {
	client := getEthClient()

	// The path to the keystore directory
	keystoreDir := "keystore"

	// Load the keystore
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// Ask the user for the password
	fmt.Print("Enter the password for the account: ")
	password := viper.GetString("eth_keystore_password")

	// Unlock the account with the password
	if err := ks.Unlock(account, password); err != nil {
		log.Fatalf("Failed to unlock account: %v", err)
	}

	fromAddress := account.Address
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	chainId := big.NewInt(viper.GetInt64("eth_chain_id"))

	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, account, chainId)
	if err != nil {
		fmt.Println("Failed to create authorized transactor: %v", err)
		return
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(contractAddress)
	contract, err := store.NewStore(address, client)
	if err != nil {
		fmt.Println("Failed to instantiate a Store contract: %v", err)
	}

	num := big.NewInt(888)
	tx, err := contract.Store(auth, num)

	// retrieved_num, err := contract.Retrieve(&bind.CallOpts{})
	// if err != nil {
	// 	fmt.Println("Failed to retrieve the value from the contract: %v", err)
	// }

	// fmt.Printf("Retrieved value from contract: %s", retrieved_num.String())

	fmt.Printf("tx sent: %s", tx.Hash().Hex())
}

func DeployEthContract(account accounts.Account) string {
	client := getEthClient()
	chainId := big.NewInt(viper.GetInt64("eth_chain_id"))
	ks := keystore.NewKeyStore("keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	// Load the private key, public key, and address

	password := viper.GetString("eth_keystore_password")
	if err := ks.Unlock(account, strings.TrimSpace(password)); err != nil {
		log.Fatalf("Failed to unlock account: %v", err)
	}

	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, account, chainId)
	if err != nil {
		fmt.Println("Failed to create authorized transactor: %v", err)
	}

	// Deploy the contract
	address, tx, _, err := store.DeployStore(auth, client)
	if err != nil {
		fmt.Println("Failed to deploy new Store contract: %v", err)
	}

	fmt.Println("Contract deployed to address %s", address.Hex())
	fmt.Println("Transaction hash: %s", tx.Hash().Hex())

	// Wait for the transaction to be mined
	contractAddress, err := bind.WaitDeployed(context.Background(), client, tx)
	if err != nil {
		fmt.Println("Failed to wait for contract deployment: %v", err)
	}

	fmt.Println(contractAddress.Hex())
	fmt.Println("Contract deployed successfully")
	return contractAddress.Hex()
}

func GenerateEthKeyPair() accounts.Account {
	// Generate a new random private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	password := viper.GetString("eth_keystore_password")
	ks := keystore.NewKeyStore("keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(privateKey, password)
	if err != nil {
		log.Fatalf("Failed to import private key: %v", err)
	}

	return account
}

func getEthWSSClient() *ethclient.Client {
	eth_wss := viper.GetString("eth_wss")
	client, err := ethclient.Dial(eth_wss)
	if err != nil {
		fmt.Println("Failed to connect to the Ethereum client: %v", err)
		return nil
	}
	return client
}

func RegistertoEvent(contractAddress string) {
	client := getEthWSSClient()
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
	fmt.Println(logs)
	fmt.Println(contractAbi)
	for {
		select {

		case vLog := <-logs:
			event := new(store.StoreNumberStored)
			err := contractAbi.UnpackIntoInterface(event, "NumberStored", vLog.Data)
			if err != nil {
				fmt.Println("Failed to unpack event data: %v", err)
			}

			fmt.Printf("New number stored: %s\n", event.NewNumber.String())
		case err := <-sub.Err():
			fmt.Println("Received subscription error: %v", err)
		}
		fmt.Println("waiting for event")
	}
}
