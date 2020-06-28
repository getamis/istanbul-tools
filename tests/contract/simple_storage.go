// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SimplestorageABI is the input ABI used to generate the binding from.
const SimplestorageABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"storedData\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"retVal\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"initVal\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"constructor\"}]"

// SimplestorageBin is the compiled bytecode used for deploying new contracts.
const SimplestorageBin = `0x6060604052341561000c57fe5b6040516020806100fa83398101604052515b60008190555b505b60c6806100346000396000f300606060405263ffffffff60e060020a6000350416632a1afcd98114603457806360fe47b11460535780636d4ce63c146065575bfe5b3415603b57fe5b60416084565b60408051918252519081900360200190f35b3415605a57fe5b6063600435608a565b005b3415606c57fe5b60416093565b60408051918252519081900360200190f35b60005481565b60008190555b50565b6000545b905600a165627a7a72305820fc8a42e6cfcb798ea356ac3f059ddefd56b41db474058cfbd48f75225d280f9a0029`

// DeploySimplestorage deploys a new Ethereum contract, binding an instance of Simplestorage to it.
func DeploySimplestorage(auth *bind.TransactOpts, backend bind.ContractBackend, initVal *big.Int) (common.Address, *types.Transaction, *Simplestorage, error) {
	parsed, err := abi.JSON(strings.NewReader(SimplestorageABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimplestorageBin), backend, initVal)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Simplestorage{SimplestorageCaller: SimplestorageCaller{contract: contract}, SimplestorageTransactor: SimplestorageTransactor{contract: contract}}, nil
}

// Simplestorage is an auto generated Go binding around an Ethereum contract.
type Simplestorage struct {
	SimplestorageCaller     // Read-only binding to the contract
	SimplestorageTransactor // Write-only binding to the contract
}

// SimplestorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimplestorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimplestorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimplestorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimplestorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimplestorageSession struct {
	Contract     *Simplestorage    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimplestorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimplestorageCallerSession struct {
	Contract *SimplestorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// SimplestorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimplestorageTransactorSession struct {
	Contract     *SimplestorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// SimplestorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimplestorageRaw struct {
	Contract *Simplestorage // Generic contract binding to access the raw methods on
}

// SimplestorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimplestorageCallerRaw struct {
	Contract *SimplestorageCaller // Generic read-only contract binding to access the raw methods on
}

// SimplestorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimplestorageTransactorRaw struct {
	Contract *SimplestorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimplestorage creates a new instance of Simplestorage, bound to a specific deployed contract.
func NewSimplestorage(address common.Address, backend bind.ContractBackend) (*Simplestorage, error) {
	contract, err := bindSimplestorage(address, backend, backend, nil)
	if err != nil {
		return nil, err
	}
	return &Simplestorage{SimplestorageCaller: SimplestorageCaller{contract: contract}, SimplestorageTransactor: SimplestorageTransactor{contract: contract}}, nil
}

// NewSimplestorageCaller creates a new read-only instance of Simplestorage, bound to a specific deployed contract.
func NewSimplestorageCaller(address common.Address, caller bind.ContractCaller) (*SimplestorageCaller, error) {
	contract, err := bindSimplestorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimplestorageCaller{contract: contract}, nil
}

// NewSimplestorageTransactor creates a new write-only instance of Simplestorage, bound to a specific deployed contract.
func NewSimplestorageTransactor(address common.Address, transactor bind.ContractTransactor) (*SimplestorageTransactor, error) {
	contract, err := bindSimplestorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimplestorageTransactor{contract: contract}, nil
}

// bindSimplestorage binds a generic wrapper to an already deployed contract.
func bindSimplestorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimplestorageABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simplestorage *SimplestorageRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Simplestorage.Contract.SimplestorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simplestorage *SimplestorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simplestorage.Contract.SimplestorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simplestorage *SimplestorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simplestorage.Contract.SimplestorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simplestorage *SimplestorageCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Simplestorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simplestorage *SimplestorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simplestorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simplestorage *SimplestorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simplestorage.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(retVal uint256)
func (_Simplestorage *SimplestorageCaller) Get(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Simplestorage.contract.Call(opts, out, "get")
	return *ret0, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(retVal uint256)
func (_Simplestorage *SimplestorageSession) Get() (*big.Int, error) {
	return _Simplestorage.Contract.Get(&_Simplestorage.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(retVal uint256)
func (_Simplestorage *SimplestorageCallerSession) Get() (*big.Int, error) {
	return _Simplestorage.Contract.Get(&_Simplestorage.CallOpts)
}

// StoredData is a free data retrieval call binding the contract method 0x2a1afcd9.
//
// Solidity: function storedData() constant returns(uint256)
func (_Simplestorage *SimplestorageCaller) StoredData(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Simplestorage.contract.Call(opts, out, "storedData")
	return *ret0, err
}

// StoredData is a free data retrieval call binding the contract method 0x2a1afcd9.
//
// Solidity: function storedData() constant returns(uint256)
func (_Simplestorage *SimplestorageSession) StoredData() (*big.Int, error) {
	return _Simplestorage.Contract.StoredData(&_Simplestorage.CallOpts)
}

// StoredData is a free data retrieval call binding the contract method 0x2a1afcd9.
//
// Solidity: function storedData() constant returns(uint256)
func (_Simplestorage *SimplestorageCallerSession) StoredData() (*big.Int, error) {
	return _Simplestorage.Contract.StoredData(&_Simplestorage.CallOpts)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(x uint256) returns()
func (_Simplestorage *SimplestorageTransactor) Set(opts *bind.TransactOpts, x *big.Int) (*types.Transaction, error) {
	return _Simplestorage.contract.Transact(opts, "set", x)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(x uint256) returns()
func (_Simplestorage *SimplestorageSession) Set(x *big.Int) (*types.Transaction, error) {
	return _Simplestorage.Contract.Set(&_Simplestorage.TransactOpts, x)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(x uint256) returns()
func (_Simplestorage *SimplestorageTransactorSession) Set(x *big.Int) (*types.Transaction, error) {
	return _Simplestorage.Contract.Set(&_Simplestorage.TransactOpts, x)
}
