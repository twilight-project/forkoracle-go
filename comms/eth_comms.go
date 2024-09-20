package comms

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
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

func SubmitNewAddress(account accounts.Account, contractAddress string, newAddress string, clientEthAddress string) {
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

	tx, err := contract.ConfirmAddress(auth, newAddress, common.HexToAddress(clientEthAddress))

	fmt.Printf("tx submit new address sent: %s", tx.Hash().Hex())
}

func SubmitSignedPsbt(account accounts.Account, psbt string, clientEthAddress string) {
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

	contractAddress := viper.GetString("eth_contract_address")
	address := common.HexToAddress(contractAddress)
	contract, err := store.NewStore(address, client)
	if err != nil {
		fmt.Println("Failed to instantiate a Store contract: %v", err)
	}

	tx, err := contract.EmitSignedBtcPsbt(auth, psbt, common.HexToAddress(clientEthAddress))

	fmt.Printf("tx signed psbt sent: %s", tx.Hash().Hex())
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

func GetEthWSSClient() *ethclient.Client {
	eth_wss := viper.GetString("eth_wss")
	client, err := ethclient.Dial(eth_wss)
	if err != nil {
		fmt.Println("Failed to connect to the Ethereum client: %v", err)
		return nil
	}
	return client
}
