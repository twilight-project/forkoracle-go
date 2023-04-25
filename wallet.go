package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	_ "github.com/lib/pq"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func create_wallet(passphrase string) (*bip32.Key, error) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		fmt.Println("Error generating mnemonic:", err)
		return nil, err
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	encryptedSeed, err := Encrypt(seed, passphrase)
	if err != nil {
		fmt.Println("Error encrypting seed:", err)
	}
	write_to_file("seed.txt", encryptedSeed)

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}
	privatekey, err := masterKey.NewChildKey(0)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}

	// Display mnemonic and keys
	fmt.Println("please save the mnemonic phrase in a safe place")
	fmt.Println("Mnemonic: ", mnemonic)

	return privatekey, nil
}

func load_wallet(passphrase string) (*bip32.Key, error) {
	encrypted_seed, err := read_from_file("seed.txt")
	if err != nil {
		fmt.Println("Error reading seed file:", err)
	}
	seed, err := Decrypt(encrypted_seed, passphrase)
	if err != nil {
		fmt.Println("Error decrypting seed:", err)
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}
	privatekey, err := masterKey.NewChildKey(0)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}

	return privatekey, nil
}

func create_wallet_from_mnemonic(mnemonic string, passphrase string) (*bip32.Key, error) {
	seed := bip39.NewSeed(mnemonic, passphrase)
	encryptedSeed, err := Encrypt(seed, passphrase)
	if err != nil {
		fmt.Println("Error encrypting seed:", err)
	}
	write_to_file("seed.txt", encryptedSeed)

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}
	privatekey, err := masterKey.NewChildKey(0)
	if err != nil {
		fmt.Println("Error :", err)
		panic(err)
	}

	return privatekey, nil
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(plainText []byte, MySecret string) ([]byte, error) {

	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		fmt.Println("Error generating IV:", err)
		return nil, err
	}

	write_to_file("iv.txt", iv)

	fmt.Println("cipher length : ", len([]byte(MySecret)))
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

func Decrypt(cipherText []byte, MySecret string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return nil, err
	}
	iv, _ := read_from_file("iv.txt")
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return plainText, nil
}

func write_to_file(name string, data []byte) {
	file, err := os.Create(name)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	defer file.Close()
}

func read_from_file(name string) ([]byte, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	// Convert the byte slice to a string and print it
	return data, nil
}
