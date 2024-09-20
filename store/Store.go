// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package store

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// StoreMetaData contains all meta data concerning the Store contract.
var StoreMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"hexBitcoinPodAddress\",\"type\":\"string\"}],\"name\":\"AddressCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"bitcoinPublicKey\",\"type\":\"bytes\"}],\"name\":\"AddressRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"balance\",\"type\":\"uint256\"}],\"name\":\"DepositConfirmed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"psbt\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"ethAddress\",\"type\":\"address\"}],\"name\":\"SignedPSBT\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"withdrawAmount\",\"type\":\"uint256\"}],\"name\":\"WithdrawalConfirmed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"hexBTCAddress\",\"type\":\"string\"}],\"name\":\"WithdrawalRequest\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"accounts\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"hexPodAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"deposit\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"hexBitcoinPodAddress\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"ethAddress\",\"type\":\"address\"}],\"name\":\"confirm_address\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"ethAddress\",\"type\":\"address\"}],\"name\":\"confirm_deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"ethAddress\",\"type\":\"address\"}],\"name\":\"confirm_withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"psbt\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"ethAddress\",\"type\":\"address\"}],\"name\":\"emit_signed_btc_psbt\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"bitcoinPublicKey\",\"type\":\"bytes\"}],\"name\":\"request_bitcoin_address\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"hexAddress\",\"type\":\"string\"}],\"name\":\"withdraw_request\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f80fd5b506111d78061001c5f395ff3fe608060405234801561000f575f80fd5b506004361061007b575f3560e01c806364f51a861161005957806364f51a86146100e8578063722a87b0146101045780639aa6f6b014610120578063eda020931461013c5761007b565b80631bf4800b1461007f578063468856bd1461009b5780635e5c06e2146100b7575b5f80fd5b610099600480360381019061009491906108a9565b610158565b005b6100b560048036038101906100b09190610960565b610291565b005b6100d160048036038101906100cc91906109bd565b6102d1565b6040516100df929190610a60565b60405180910390f35b61010260048036038101906100fd9190610ab8565b610376565b005b61011e60048036038101906101199190610af6565b610481565b005b61013a60048036038101906101359190610ab8565b6105c8565b005b61015660048036038101906101519190610bdb565b6106c8565b005b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f0180546101a290610c4f565b9050146101e4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101db90610cc9565b60405180910390fd5b60405180604001604052808381526020015f8152505f808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f820151815f0190816102489190610e84565b50602082015181600101559050507f4ccb52dccd89dab6fe1b0067fda2557678a6608b46cb41bcd9e61ddf5248d3f9826040516102859190610f53565b60405180910390a15050565b7f6bac91f92e9b5ba1785bd0976843e48e3f7e82242b2efe0976b9e185410775ce8383836040516102c493929190610fae565b60405180910390a1505050565b5f602052805f5260405f205f91509050805f0180546102ef90610c4f565b80601f016020809104026020016040519081016040528092919081815260200182805461031b90610c4f565b80156103665780601f1061033d57610100808354040283529160200191610366565b820191905f5260205f20905b81548152906001019060200180831161034957829003601f168201915b5050505050908060010154905082565b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f0180546103c090610c4f565b905003610402576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103f990611028565b60405180910390fd5b815f808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20600101819055507ff830ab6ff7ae5f21a222c7780fb5d8aee200c86fe582dbb0986fed90aa3c83dc826040516104759190611046565b60405180910390a15050565b5f805f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f0180546104cb90610c4f565b90500361050d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161050490611028565b60405180910390fd5b5f805f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20600101541161058e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610585906110a9565b60405180910390fd5b7fd5c210ff7021a30d5949fcb7278ffba026edd1738385a02b70008c2076b5995a816040516105bd9190610f53565b60405180910390a150565b815f808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f206001015414610649576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161064090611111565b60405180910390fd5b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20600101819055507ff4e72f675e28ea6492e6c3f45e87ae0e4a0ce105ccd24ce5ae44e09351101a03826040516106bc9190611046565b60405180910390a15050565b7f357dccb790461427b31be4c04f456f4d349040cc74df6a94a45844c5987fd697816040516106f79190611181565b60405180910390a150565b5f604051905090565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6107618261071b565b810181811067ffffffffffffffff821117156107805761077f61072b565b5b80604052505050565b5f610792610702565b905061079e8282610758565b919050565b5f67ffffffffffffffff8211156107bd576107bc61072b565b5b6107c68261071b565b9050602081019050919050565b828183375f83830152505050565b5f6107f36107ee846107a3565b610789565b90508281526020810184848401111561080f5761080e610717565b5b61081a8482856107d3565b509392505050565b5f82601f83011261083657610835610713565b5b81356108468482602086016107e1565b91505092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6108788261084f565b9050919050565b6108888161086e565b8114610892575f80fd5b50565b5f813590506108a38161087f565b92915050565b5f80604083850312156108bf576108be61070b565b5b5f83013567ffffffffffffffff8111156108dc576108db61070f565b5b6108e885828601610822565b92505060206108f985828601610895565b9150509250929050565b5f80fd5b5f80fd5b5f8083601f8401126109205761091f610713565b5b8235905067ffffffffffffffff81111561093d5761093c610903565b5b60208301915083600182028301111561095957610958610907565b5b9250929050565b5f805f604084860312156109775761097661070b565b5b5f84013567ffffffffffffffff8111156109945761099361070f565b5b6109a08682870161090b565b935093505060206109b386828701610895565b9150509250925092565b5f602082840312156109d2576109d161070b565b5b5f6109df84828501610895565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f610a1a826109e8565b610a2481856109f2565b9350610a34818560208601610a02565b610a3d8161071b565b840191505092915050565b5f819050919050565b610a5a81610a48565b82525050565b5f6040820190508181035f830152610a788185610a10565b9050610a876020830184610a51565b9392505050565b610a9781610a48565b8114610aa1575f80fd5b50565b5f81359050610ab281610a8e565b92915050565b5f8060408385031215610ace57610acd61070b565b5b5f610adb85828601610aa4565b9250506020610aec85828601610895565b9150509250929050565b5f60208284031215610b0b57610b0a61070b565b5b5f82013567ffffffffffffffff811115610b2857610b2761070f565b5b610b3484828501610822565b91505092915050565b5f67ffffffffffffffff821115610b5757610b5661072b565b5b610b608261071b565b9050602081019050919050565b5f610b7f610b7a84610b3d565b610789565b905082815260208101848484011115610b9b57610b9a610717565b5b610ba68482856107d3565b509392505050565b5f82601f830112610bc257610bc1610713565b5b8135610bd2848260208601610b6d565b91505092915050565b5f60208284031215610bf057610bef61070b565b5b5f82013567ffffffffffffffff811115610c0d57610c0c61070f565b5b610c1984828501610bae565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610c6657607f821691505b602082108103610c7957610c78610c22565b5b50919050565b7f4164647265737320616c726561647920637265617465640000000000000000005f82015250565b5f610cb36017836109f2565b9150610cbe82610c7f565b602082019050919050565b5f6020820190508181035f830152610ce081610ca7565b9050919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302610d437fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610d08565b610d4d8683610d08565b95508019841693508086168417925050509392505050565b5f819050919050565b5f610d88610d83610d7e84610a48565b610d65565b610a48565b9050919050565b5f819050919050565b610da183610d6e565b610db5610dad82610d8f565b848454610d14565b825550505050565b5f90565b610dc9610dbd565b610dd4818484610d98565b505050565b5b81811015610df757610dec5f82610dc1565b600181019050610dda565b5050565b601f821115610e3c57610e0d81610ce7565b610e1684610cf9565b81016020851015610e25578190505b610e39610e3185610cf9565b830182610dd9565b50505b505050565b5f82821c905092915050565b5f610e5c5f1984600802610e41565b1980831691505092915050565b5f610e748383610e4d565b9150826002028217905092915050565b610e8d826109e8565b67ffffffffffffffff811115610ea657610ea561072b565b5b610eb08254610c4f565b610ebb828285610dfb565b5f60209050601f831160018114610eec575f8415610eda578287015190505b610ee48582610e69565b865550610f4b565b601f198416610efa86610ce7565b5f5b82811015610f2157848901518255600182019150602085019450602081019050610efc565b86831015610f3e5784890151610f3a601f891682610e4d565b8355505b6001600288020188555050505b505050505050565b5f6020820190508181035f830152610f6b8184610a10565b905092915050565b5f610f7e83856109f2565b9350610f8b8385846107d3565b610f948361071b565b840190509392505050565b610fa88161086e565b82525050565b5f6040820190508181035f830152610fc7818587610f73565b9050610fd66020830184610f9f565b949350505050565b7f41646472657373206e6f742063726561746564207965740000000000000000005f82015250565b5f6110126017836109f2565b915061101d82610fde565b602082019050919050565b5f6020820190508181035f83015261103f81611006565b9050919050565b5f6020820190506110595f830184610a51565b92915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f6110936014836109f2565b915061109e8261105f565b602082019050919050565b5f6020820190508181035f8301526110c081611087565b9050919050565b7f5769746864726177616c20416d6f756e7420646f6573206e6f74206d617463685f82015250565b5f6110fb6020836109f2565b9150611106826110c7565b602082019050919050565b5f6020820190508181035f830152611128816110ef565b9050919050565b5f81519050919050565b5f82825260208201905092915050565b5f6111538261112f565b61115d8185611139565b935061116d818560208601610a02565b6111768161071b565b840191505092915050565b5f6020820190508181035f8301526111998184611149565b90509291505056fea264697066735822122093c422ad8c1077650c5305e0da095800cbe124db23200346611a35730924e4b364736f6c634300081a0033",
}

// StoreABI is the input ABI used to generate the binding from.
// Deprecated: Use StoreMetaData.ABI instead.
var StoreABI = StoreMetaData.ABI

// StoreBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use StoreMetaData.Bin instead.
var StoreBin = StoreMetaData.Bin

// DeployStore deploys a new Ethereum contract, binding an instance of Store to it.
func DeployStore(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Store, error) {
	parsed, err := StoreMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(StoreBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

// Store is an auto generated Go binding around an Ethereum contract.
type Store struct {
	StoreCaller     // Read-only binding to the contract
	StoreTransactor // Write-only binding to the contract
	StoreFilterer   // Log filterer for contract events
}

// StoreCaller is an auto generated read-only Go binding around an Ethereum contract.
type StoreCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StoreTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StoreFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StoreSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StoreSession struct {
	Contract     *Store            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StoreCallerSession struct {
	Contract *StoreCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// StoreTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StoreTransactorSession struct {
	Contract     *StoreTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StoreRaw is an auto generated low-level Go binding around an Ethereum contract.
type StoreRaw struct {
	Contract *Store // Generic contract binding to access the raw methods on
}

// StoreCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StoreCallerRaw struct {
	Contract *StoreCaller // Generic read-only contract binding to access the raw methods on
}

// StoreTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StoreTransactorRaw struct {
	Contract *StoreTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStore creates a new instance of Store, bound to a specific deployed contract.
func NewStore(address common.Address, backend bind.ContractBackend) (*Store, error) {
	contract, err := bindStore(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Store{StoreCaller: StoreCaller{contract: contract}, StoreTransactor: StoreTransactor{contract: contract}, StoreFilterer: StoreFilterer{contract: contract}}, nil
}

// NewStoreCaller creates a new read-only instance of Store, bound to a specific deployed contract.
func NewStoreCaller(address common.Address, caller bind.ContractCaller) (*StoreCaller, error) {
	contract, err := bindStore(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StoreCaller{contract: contract}, nil
}

// NewStoreTransactor creates a new write-only instance of Store, bound to a specific deployed contract.
func NewStoreTransactor(address common.Address, transactor bind.ContractTransactor) (*StoreTransactor, error) {
	contract, err := bindStore(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StoreTransactor{contract: contract}, nil
}

// NewStoreFilterer creates a new log filterer instance of Store, bound to a specific deployed contract.
func NewStoreFilterer(address common.Address, filterer bind.ContractFilterer) (*StoreFilterer, error) {
	contract, err := bindStore(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StoreFilterer{contract: contract}, nil
}

// bindStore binds a generic wrapper to an already deployed contract.
func bindStore(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StoreMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Store.Contract.StoreCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Store.Contract.StoreTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Store.Contract.StoreTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Store *StoreCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Store.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Store *StoreTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Store.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Store *StoreTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Store.Contract.contract.Transact(opts, method, params...)
}

// Accounts is a free data retrieval call binding the contract method 0x5e5c06e2.
//
// Solidity: function accounts(address ) view returns(string hexPodAddress, uint256 deposit)
func (_Store *StoreCaller) Accounts(opts *bind.CallOpts, arg0 common.Address) (struct {
	HexPodAddress string
	Deposit       *big.Int
}, error) {
	var out []interface{}
	err := _Store.contract.Call(opts, &out, "accounts", arg0)

	outstruct := new(struct {
		HexPodAddress string
		Deposit       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.HexPodAddress = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Deposit = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Accounts is a free data retrieval call binding the contract method 0x5e5c06e2.
//
// Solidity: function accounts(address ) view returns(string hexPodAddress, uint256 deposit)
func (_Store *StoreSession) Accounts(arg0 common.Address) (struct {
	HexPodAddress string
	Deposit       *big.Int
}, error) {
	return _Store.Contract.Accounts(&_Store.CallOpts, arg0)
}

// Accounts is a free data retrieval call binding the contract method 0x5e5c06e2.
//
// Solidity: function accounts(address ) view returns(string hexPodAddress, uint256 deposit)
func (_Store *StoreCallerSession) Accounts(arg0 common.Address) (struct {
	HexPodAddress string
	Deposit       *big.Int
}, error) {
	return _Store.Contract.Accounts(&_Store.CallOpts, arg0)
}

// ConfirmAddress is a paid mutator transaction binding the contract method 0x1bf4800b.
//
// Solidity: function confirm_address(string hexBitcoinPodAddress, address ethAddress) returns()
func (_Store *StoreTransactor) ConfirmAddress(opts *bind.TransactOpts, hexBitcoinPodAddress string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "confirm_address", hexBitcoinPodAddress, ethAddress)
}

// ConfirmAddress is a paid mutator transaction binding the contract method 0x1bf4800b.
//
// Solidity: function confirm_address(string hexBitcoinPodAddress, address ethAddress) returns()
func (_Store *StoreSession) ConfirmAddress(hexBitcoinPodAddress string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmAddress(&_Store.TransactOpts, hexBitcoinPodAddress, ethAddress)
}

// ConfirmAddress is a paid mutator transaction binding the contract method 0x1bf4800b.
//
// Solidity: function confirm_address(string hexBitcoinPodAddress, address ethAddress) returns()
func (_Store *StoreTransactorSession) ConfirmAddress(hexBitcoinPodAddress string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmAddress(&_Store.TransactOpts, hexBitcoinPodAddress, ethAddress)
}

// ConfirmDeposit is a paid mutator transaction binding the contract method 0x64f51a86.
//
// Solidity: function confirm_deposit(uint256 amount, address ethAddress) returns()
func (_Store *StoreTransactor) ConfirmDeposit(opts *bind.TransactOpts, amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "confirm_deposit", amount, ethAddress)
}

// ConfirmDeposit is a paid mutator transaction binding the contract method 0x64f51a86.
//
// Solidity: function confirm_deposit(uint256 amount, address ethAddress) returns()
func (_Store *StoreSession) ConfirmDeposit(amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmDeposit(&_Store.TransactOpts, amount, ethAddress)
}

// ConfirmDeposit is a paid mutator transaction binding the contract method 0x64f51a86.
//
// Solidity: function confirm_deposit(uint256 amount, address ethAddress) returns()
func (_Store *StoreTransactorSession) ConfirmDeposit(amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmDeposit(&_Store.TransactOpts, amount, ethAddress)
}

// ConfirmWithdraw is a paid mutator transaction binding the contract method 0x9aa6f6b0.
//
// Solidity: function confirm_withdraw(uint256 amount, address ethAddress) returns()
func (_Store *StoreTransactor) ConfirmWithdraw(opts *bind.TransactOpts, amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "confirm_withdraw", amount, ethAddress)
}

// ConfirmWithdraw is a paid mutator transaction binding the contract method 0x9aa6f6b0.
//
// Solidity: function confirm_withdraw(uint256 amount, address ethAddress) returns()
func (_Store *StoreSession) ConfirmWithdraw(amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmWithdraw(&_Store.TransactOpts, amount, ethAddress)
}

// ConfirmWithdraw is a paid mutator transaction binding the contract method 0x9aa6f6b0.
//
// Solidity: function confirm_withdraw(uint256 amount, address ethAddress) returns()
func (_Store *StoreTransactorSession) ConfirmWithdraw(amount *big.Int, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.ConfirmWithdraw(&_Store.TransactOpts, amount, ethAddress)
}

// EmitSignedBtcPsbt is a paid mutator transaction binding the contract method 0x468856bd.
//
// Solidity: function emit_signed_btc_psbt(string psbt, address ethAddress) returns()
func (_Store *StoreTransactor) EmitSignedBtcPsbt(opts *bind.TransactOpts, psbt string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "emit_signed_btc_psbt", psbt, ethAddress)
}

// EmitSignedBtcPsbt is a paid mutator transaction binding the contract method 0x468856bd.
//
// Solidity: function emit_signed_btc_psbt(string psbt, address ethAddress) returns()
func (_Store *StoreSession) EmitSignedBtcPsbt(psbt string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.EmitSignedBtcPsbt(&_Store.TransactOpts, psbt, ethAddress)
}

// EmitSignedBtcPsbt is a paid mutator transaction binding the contract method 0x468856bd.
//
// Solidity: function emit_signed_btc_psbt(string psbt, address ethAddress) returns()
func (_Store *StoreTransactorSession) EmitSignedBtcPsbt(psbt string, ethAddress common.Address) (*types.Transaction, error) {
	return _Store.Contract.EmitSignedBtcPsbt(&_Store.TransactOpts, psbt, ethAddress)
}

// RequestBitcoinAddress is a paid mutator transaction binding the contract method 0xeda02093.
//
// Solidity: function request_bitcoin_address(bytes bitcoinPublicKey) returns()
func (_Store *StoreTransactor) RequestBitcoinAddress(opts *bind.TransactOpts, bitcoinPublicKey []byte) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "request_bitcoin_address", bitcoinPublicKey)
}

// RequestBitcoinAddress is a paid mutator transaction binding the contract method 0xeda02093.
//
// Solidity: function request_bitcoin_address(bytes bitcoinPublicKey) returns()
func (_Store *StoreSession) RequestBitcoinAddress(bitcoinPublicKey []byte) (*types.Transaction, error) {
	return _Store.Contract.RequestBitcoinAddress(&_Store.TransactOpts, bitcoinPublicKey)
}

// RequestBitcoinAddress is a paid mutator transaction binding the contract method 0xeda02093.
//
// Solidity: function request_bitcoin_address(bytes bitcoinPublicKey) returns()
func (_Store *StoreTransactorSession) RequestBitcoinAddress(bitcoinPublicKey []byte) (*types.Transaction, error) {
	return _Store.Contract.RequestBitcoinAddress(&_Store.TransactOpts, bitcoinPublicKey)
}

// WithdrawRequest is a paid mutator transaction binding the contract method 0x722a87b0.
//
// Solidity: function withdraw_request(string hexAddress) returns()
func (_Store *StoreTransactor) WithdrawRequest(opts *bind.TransactOpts, hexAddress string) (*types.Transaction, error) {
	return _Store.contract.Transact(opts, "withdraw_request", hexAddress)
}

// WithdrawRequest is a paid mutator transaction binding the contract method 0x722a87b0.
//
// Solidity: function withdraw_request(string hexAddress) returns()
func (_Store *StoreSession) WithdrawRequest(hexAddress string) (*types.Transaction, error) {
	return _Store.Contract.WithdrawRequest(&_Store.TransactOpts, hexAddress)
}

// WithdrawRequest is a paid mutator transaction binding the contract method 0x722a87b0.
//
// Solidity: function withdraw_request(string hexAddress) returns()
func (_Store *StoreTransactorSession) WithdrawRequest(hexAddress string) (*types.Transaction, error) {
	return _Store.Contract.WithdrawRequest(&_Store.TransactOpts, hexAddress)
}

// StoreAddressCreatedIterator is returned from FilterAddressCreated and is used to iterate over the raw logs and unpacked data for AddressCreated events raised by the Store contract.
type StoreAddressCreatedIterator struct {
	Event *StoreAddressCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreAddressCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreAddressCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreAddressCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreAddressCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreAddressCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreAddressCreated represents a AddressCreated event raised by the Store contract.
type StoreAddressCreated struct {
	HexBitcoinPodAddress string
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterAddressCreated is a free log retrieval operation binding the contract event 0x4ccb52dccd89dab6fe1b0067fda2557678a6608b46cb41bcd9e61ddf5248d3f9.
//
// Solidity: event AddressCreated(string hexBitcoinPodAddress)
func (_Store *StoreFilterer) FilterAddressCreated(opts *bind.FilterOpts) (*StoreAddressCreatedIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "AddressCreated")
	if err != nil {
		return nil, err
	}
	return &StoreAddressCreatedIterator{contract: _Store.contract, event: "AddressCreated", logs: logs, sub: sub}, nil
}

// WatchAddressCreated is a free log subscription operation binding the contract event 0x4ccb52dccd89dab6fe1b0067fda2557678a6608b46cb41bcd9e61ddf5248d3f9.
//
// Solidity: event AddressCreated(string hexBitcoinPodAddress)
func (_Store *StoreFilterer) WatchAddressCreated(opts *bind.WatchOpts, sink chan<- *StoreAddressCreated) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "AddressCreated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreAddressCreated)
				if err := _Store.contract.UnpackLog(event, "AddressCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddressCreated is a log parse operation binding the contract event 0x4ccb52dccd89dab6fe1b0067fda2557678a6608b46cb41bcd9e61ddf5248d3f9.
//
// Solidity: event AddressCreated(string hexBitcoinPodAddress)
func (_Store *StoreFilterer) ParseAddressCreated(log types.Log) (*StoreAddressCreated, error) {
	event := new(StoreAddressCreated)
	if err := _Store.contract.UnpackLog(event, "AddressCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StoreAddressRequestedIterator is returned from FilterAddressRequested and is used to iterate over the raw logs and unpacked data for AddressRequested events raised by the Store contract.
type StoreAddressRequestedIterator struct {
	Event *StoreAddressRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreAddressRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreAddressRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreAddressRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreAddressRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreAddressRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreAddressRequested represents a AddressRequested event raised by the Store contract.
type StoreAddressRequested struct {
	BitcoinPublicKey []byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterAddressRequested is a free log retrieval operation binding the contract event 0x357dccb790461427b31be4c04f456f4d349040cc74df6a94a45844c5987fd697.
//
// Solidity: event AddressRequested(bytes bitcoinPublicKey)
func (_Store *StoreFilterer) FilterAddressRequested(opts *bind.FilterOpts) (*StoreAddressRequestedIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "AddressRequested")
	if err != nil {
		return nil, err
	}
	return &StoreAddressRequestedIterator{contract: _Store.contract, event: "AddressRequested", logs: logs, sub: sub}, nil
}

// WatchAddressRequested is a free log subscription operation binding the contract event 0x357dccb790461427b31be4c04f456f4d349040cc74df6a94a45844c5987fd697.
//
// Solidity: event AddressRequested(bytes bitcoinPublicKey)
func (_Store *StoreFilterer) WatchAddressRequested(opts *bind.WatchOpts, sink chan<- *StoreAddressRequested) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "AddressRequested")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreAddressRequested)
				if err := _Store.contract.UnpackLog(event, "AddressRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddressRequested is a log parse operation binding the contract event 0x357dccb790461427b31be4c04f456f4d349040cc74df6a94a45844c5987fd697.
//
// Solidity: event AddressRequested(bytes bitcoinPublicKey)
func (_Store *StoreFilterer) ParseAddressRequested(log types.Log) (*StoreAddressRequested, error) {
	event := new(StoreAddressRequested)
	if err := _Store.contract.UnpackLog(event, "AddressRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StoreDepositConfirmedIterator is returned from FilterDepositConfirmed and is used to iterate over the raw logs and unpacked data for DepositConfirmed events raised by the Store contract.
type StoreDepositConfirmedIterator struct {
	Event *StoreDepositConfirmed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreDepositConfirmedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreDepositConfirmed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreDepositConfirmed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreDepositConfirmedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreDepositConfirmedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreDepositConfirmed represents a DepositConfirmed event raised by the Store contract.
type StoreDepositConfirmed struct {
	Balance *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDepositConfirmed is a free log retrieval operation binding the contract event 0xf830ab6ff7ae5f21a222c7780fb5d8aee200c86fe582dbb0986fed90aa3c83dc.
//
// Solidity: event DepositConfirmed(uint256 balance)
func (_Store *StoreFilterer) FilterDepositConfirmed(opts *bind.FilterOpts) (*StoreDepositConfirmedIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "DepositConfirmed")
	if err != nil {
		return nil, err
	}
	return &StoreDepositConfirmedIterator{contract: _Store.contract, event: "DepositConfirmed", logs: logs, sub: sub}, nil
}

// WatchDepositConfirmed is a free log subscription operation binding the contract event 0xf830ab6ff7ae5f21a222c7780fb5d8aee200c86fe582dbb0986fed90aa3c83dc.
//
// Solidity: event DepositConfirmed(uint256 balance)
func (_Store *StoreFilterer) WatchDepositConfirmed(opts *bind.WatchOpts, sink chan<- *StoreDepositConfirmed) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "DepositConfirmed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreDepositConfirmed)
				if err := _Store.contract.UnpackLog(event, "DepositConfirmed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositConfirmed is a log parse operation binding the contract event 0xf830ab6ff7ae5f21a222c7780fb5d8aee200c86fe582dbb0986fed90aa3c83dc.
//
// Solidity: event DepositConfirmed(uint256 balance)
func (_Store *StoreFilterer) ParseDepositConfirmed(log types.Log) (*StoreDepositConfirmed, error) {
	event := new(StoreDepositConfirmed)
	if err := _Store.contract.UnpackLog(event, "DepositConfirmed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StoreSignedPSBTIterator is returned from FilterSignedPSBT and is used to iterate over the raw logs and unpacked data for SignedPSBT events raised by the Store contract.
type StoreSignedPSBTIterator struct {
	Event *StoreSignedPSBT // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreSignedPSBTIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreSignedPSBT)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreSignedPSBT)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreSignedPSBTIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreSignedPSBTIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreSignedPSBT represents a SignedPSBT event raised by the Store contract.
type StoreSignedPSBT struct {
	Psbt       string
	EthAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSignedPSBT is a free log retrieval operation binding the contract event 0x6bac91f92e9b5ba1785bd0976843e48e3f7e82242b2efe0976b9e185410775ce.
//
// Solidity: event SignedPSBT(string psbt, address ethAddress)
func (_Store *StoreFilterer) FilterSignedPSBT(opts *bind.FilterOpts) (*StoreSignedPSBTIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "SignedPSBT")
	if err != nil {
		return nil, err
	}
	return &StoreSignedPSBTIterator{contract: _Store.contract, event: "SignedPSBT", logs: logs, sub: sub}, nil
}

// WatchSignedPSBT is a free log subscription operation binding the contract event 0x6bac91f92e9b5ba1785bd0976843e48e3f7e82242b2efe0976b9e185410775ce.
//
// Solidity: event SignedPSBT(string psbt, address ethAddress)
func (_Store *StoreFilterer) WatchSignedPSBT(opts *bind.WatchOpts, sink chan<- *StoreSignedPSBT) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "SignedPSBT")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreSignedPSBT)
				if err := _Store.contract.UnpackLog(event, "SignedPSBT", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignedPSBT is a log parse operation binding the contract event 0x6bac91f92e9b5ba1785bd0976843e48e3f7e82242b2efe0976b9e185410775ce.
//
// Solidity: event SignedPSBT(string psbt, address ethAddress)
func (_Store *StoreFilterer) ParseSignedPSBT(log types.Log) (*StoreSignedPSBT, error) {
	event := new(StoreSignedPSBT)
	if err := _Store.contract.UnpackLog(event, "SignedPSBT", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StoreWithdrawalConfirmedIterator is returned from FilterWithdrawalConfirmed and is used to iterate over the raw logs and unpacked data for WithdrawalConfirmed events raised by the Store contract.
type StoreWithdrawalConfirmedIterator struct {
	Event *StoreWithdrawalConfirmed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreWithdrawalConfirmedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreWithdrawalConfirmed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreWithdrawalConfirmed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreWithdrawalConfirmedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreWithdrawalConfirmedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreWithdrawalConfirmed represents a WithdrawalConfirmed event raised by the Store contract.
type StoreWithdrawalConfirmed struct {
	WithdrawAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalConfirmed is a free log retrieval operation binding the contract event 0xf4e72f675e28ea6492e6c3f45e87ae0e4a0ce105ccd24ce5ae44e09351101a03.
//
// Solidity: event WithdrawalConfirmed(uint256 withdrawAmount)
func (_Store *StoreFilterer) FilterWithdrawalConfirmed(opts *bind.FilterOpts) (*StoreWithdrawalConfirmedIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "WithdrawalConfirmed")
	if err != nil {
		return nil, err
	}
	return &StoreWithdrawalConfirmedIterator{contract: _Store.contract, event: "WithdrawalConfirmed", logs: logs, sub: sub}, nil
}

// WatchWithdrawalConfirmed is a free log subscription operation binding the contract event 0xf4e72f675e28ea6492e6c3f45e87ae0e4a0ce105ccd24ce5ae44e09351101a03.
//
// Solidity: event WithdrawalConfirmed(uint256 withdrawAmount)
func (_Store *StoreFilterer) WatchWithdrawalConfirmed(opts *bind.WatchOpts, sink chan<- *StoreWithdrawalConfirmed) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "WithdrawalConfirmed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreWithdrawalConfirmed)
				if err := _Store.contract.UnpackLog(event, "WithdrawalConfirmed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalConfirmed is a log parse operation binding the contract event 0xf4e72f675e28ea6492e6c3f45e87ae0e4a0ce105ccd24ce5ae44e09351101a03.
//
// Solidity: event WithdrawalConfirmed(uint256 withdrawAmount)
func (_Store *StoreFilterer) ParseWithdrawalConfirmed(log types.Log) (*StoreWithdrawalConfirmed, error) {
	event := new(StoreWithdrawalConfirmed)
	if err := _Store.contract.UnpackLog(event, "WithdrawalConfirmed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StoreWithdrawalRequestIterator is returned from FilterWithdrawalRequest and is used to iterate over the raw logs and unpacked data for WithdrawalRequest events raised by the Store contract.
type StoreWithdrawalRequestIterator struct {
	Event *StoreWithdrawalRequest // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *StoreWithdrawalRequestIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StoreWithdrawalRequest)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(StoreWithdrawalRequest)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *StoreWithdrawalRequestIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StoreWithdrawalRequestIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StoreWithdrawalRequest represents a WithdrawalRequest event raised by the Store contract.
type StoreWithdrawalRequest struct {
	HexBTCAddress string
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalRequest is a free log retrieval operation binding the contract event 0xd5c210ff7021a30d5949fcb7278ffba026edd1738385a02b70008c2076b5995a.
//
// Solidity: event WithdrawalRequest(string hexBTCAddress)
func (_Store *StoreFilterer) FilterWithdrawalRequest(opts *bind.FilterOpts) (*StoreWithdrawalRequestIterator, error) {

	logs, sub, err := _Store.contract.FilterLogs(opts, "WithdrawalRequest")
	if err != nil {
		return nil, err
	}
	return &StoreWithdrawalRequestIterator{contract: _Store.contract, event: "WithdrawalRequest", logs: logs, sub: sub}, nil
}

// WatchWithdrawalRequest is a free log subscription operation binding the contract event 0xd5c210ff7021a30d5949fcb7278ffba026edd1738385a02b70008c2076b5995a.
//
// Solidity: event WithdrawalRequest(string hexBTCAddress)
func (_Store *StoreFilterer) WatchWithdrawalRequest(opts *bind.WatchOpts, sink chan<- *StoreWithdrawalRequest) (event.Subscription, error) {

	logs, sub, err := _Store.contract.WatchLogs(opts, "WithdrawalRequest")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StoreWithdrawalRequest)
				if err := _Store.contract.UnpackLog(event, "WithdrawalRequest", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalRequest is a log parse operation binding the contract event 0xd5c210ff7021a30d5949fcb7278ffba026edd1738385a02b70008c2076b5995a.
//
// Solidity: event WithdrawalRequest(string hexBTCAddress)
func (_Store *StoreFilterer) ParseWithdrawalRequest(log types.Log) (*StoreWithdrawalRequest, error) {
	event := new(StoreWithdrawalRequest)
	if err := _Store.contract.UnpackLog(event, "WithdrawalRequest", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
