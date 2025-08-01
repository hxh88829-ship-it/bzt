// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bzt

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

// BztMetaData contains all meta data concerning the Bzt contract.
var BztMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_usdtToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Airdrop\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"int256\",\"name\":\"profitLoss\",\"type\":\"int256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"openPrice\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"closePrice\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"OrderClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"tokenName\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"OrderOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"airdrop\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_orderId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_openPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_closePrice\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_tokenName\",\"type\":\"string\"}],\"name\":\"closeOrder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getContractBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_orderId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_tokenName\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"openOrder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"orders\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tokenName\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"openPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closePrice\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"profitLoss\",\"type\":\"int256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isClosed\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"usdtToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506040516112ab3803806112ab83398101604081905261002f916100d8565b338061005557604051631e4fbdf760e01b81526000600482015260240160405180910390fd5b61005e81610088565b5060018055600280546001600160a01b0319166001600160a01b0392909216919091179055610108565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156100ea57600080fd5b81516001600160a01b038116811461010157600080fd5b9392505050565b611194806101176000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c80638da5cb5b116100665780638da5cb5b146100e3578063a85c38ef14610108578063a98ad46c1461012f578063f2fde38b14610142578063fd53e1f21461015557600080fd5b806310199bdc146100985780636f9fb98a146100ad578063715018a6146100c85780638ba4cc3c146100d0575b600080fd5b6100ab6100a6366004610d0a565b610168565b005b6100b56105b0565b6040519081526020015b60405180910390f35b6100ab610622565b6100ab6100de366004610d80565b610636565b6000546001600160a01b03165b6040516001600160a01b0390911681526020016100bf565b61011b610116366004610daa565b6107f3565b6040516100bf989796959493929190610e09565b6002546100f0906001600160a01b031681565b6100ab610150366004610e61565b6108cd565b6100ab610163366004610e83565b61090b565b610170610b8f565b610178610bbc565b600084815260036020526040812080549091036101d35760405162461bcd60e51b815260206004820152601460248201527313dc99195c88191bd95cc81b9bdd08195e1a5cdd60621b60448201526064015b60405180910390fd5b6006810154600160a01b900460ff16156102265760405162461bcd60e51b815260206004820152601460248201527313dc99195c88185b1c9958591e4818db1bdcd95960621b60448201526064016101ca565b6000841180156102365750600083115b6102825760405162461bcd60e51b815260206004820152601d60248201527f507269636573206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60028101849055600381018390556001810161029e8382610f5c565b5060068101805460ff60a01b1916600160a01b179055600080858511156103aa57600583015486906102d08288611031565b6102da919061104a565b6102e49190611061565b915081905060006102f6600283611061565b9050600081856005015461030a9190611083565b600254600687015460405163a9059cbb60e01b81526001600160a01b03918216600482015260248101849052929350169063a9059cbb906044016020604051808303816000875af1158015610363573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103879190611096565b6103a35760405162461bcd60e51b81526004016101ca906110b8565b505061053c565b8585101561049b57600583015486906103c38783611031565b6103cd919061104a565b6103d79190611061565b90506103e2816110ef565b915060008184600501546103f69190611031565b9050801561049557600254600685015460405163a9059cbb60e01b81526001600160a01b0391821660048201526024810184905291169063a9059cbb906044016020604051808303816000875af1158015610455573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104799190611096565b6104955760405162461bcd60e51b81526004016101ca906110b8565b5061053c565b6002546006840154600585015460405163a9059cbb60e01b81526001600160a01b03928316600482015260248101919091526000945091169063a9059cbb906044016020604051808303816000875af11580156104fc573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105209190611096565b61053c5760405162461bcd60e51b81526004016101ca906110b8565b6004830182905560068301546040805189815260208101859052808201899052606081018890526001600160a01b039092166080830152517f06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d969181900360a00190a15050506105aa60018055565b50505050565b6002546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a0823190602401602060405180830381865afa1580156105f9573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061061d919061110b565b905090565b61062a610b8f565b6106346000610c15565b565b61063e610b8f565b610646610bbc565b6001600160a01b0382166106915760405162461bcd60e51b8152602060048201526012602482015271496e76616c696420746f206164647265737360701b60448201526064016101ca565b600081116106e15760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60025460405163a9059cbb60e01b81526001600160a01b038481166004830152602482018490529091169063a9059cbb906044016020604051808303816000875af1158015610734573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107589190611096565b6107a45760405162461bcd60e51b815260206004820152601c60248201527f555344542061697264726f70207472616e73666572206661696c65640000000060448201526064016101ca565b604080516001600160a01b0384168152602081018390527f8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a910160405180910390a16107ef60018055565b5050565b6003602052600090815260409020805460018201805491929161081590610ed4565b80601f016020809104026020016040519081016040528092919081815260200182805461084190610ed4565b801561088e5780601f106108635761010080835404028352916020019161088e565b820191906000526020600020905b81548152906001019060200180831161087157829003601f168201915b50505060028401546003850154600486015460058701546006909701549596929591945092506001600160a01b0381169060ff600160a01b9091041688565b6108d5610b8f565b6001600160a01b0381166108ff57604051631e4fbdf760e01b8152600060048201526024016101ca565b61090881610c15565b50565b610913610bbc565b600081116109635760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b600083815260036020526040902054156109bf5760405162461bcd60e51b815260206004820152601760248201527f4f7264657220494420616c72656164792065786973747300000000000000000060448201526064016101ca565b6002546040516323b872dd60e01b8152336004820152306024820152604481018390526001600160a01b03909116906323b872dd906064016020604051808303816000875af1158015610a16573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a3a9190611096565b610a7d5760405162461bcd60e51b81526020600482015260146024820152731554d115081d1c985b9cd9995c8819985a5b195960621b60448201526064016101ca565b604080516101008101825284815260208082018581526000838501819052606084018190526080840181905260a084018690523360c085015260e0840181905287815260039092529290208151815591519091906001820190610ae09082610f5c565b506040828101516002830155606083015160038301556080830151600483015560a0830151600583015560c08301516006909201805460e0909401511515600160a01b026001600160a81b03199094166001600160a01b0390931692909217929092179055517fee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b590610b79908590859085903390611124565b60405180910390a1610b8a60018055565b505050565b6000546001600160a01b031633146106345760405163118cdaa760e01b81523360048201526024016101ca565b600260015403610c0e5760405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c0060448201526064016101ca565b6002600155565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b634e487b7160e01b600052604160045260246000fd5b600082601f830112610c8c57600080fd5b813567ffffffffffffffff811115610ca657610ca6610c65565b604051601f8201601f19908116603f0116810167ffffffffffffffff81118282101715610cd557610cd5610c65565b604052818152838201602001851015610ced57600080fd5b816020850160208301376000918101602001919091529392505050565b60008060008060808587031215610d2057600080fd5b843593506020850135925060408501359150606085013567ffffffffffffffff811115610d4c57600080fd5b610d5887828801610c7b565b91505092959194509250565b80356001600160a01b0381168114610d7b57600080fd5b919050565b60008060408385031215610d9357600080fd5b610d9c83610d64565b946020939093013593505050565b600060208284031215610dbc57600080fd5b5035919050565b6000815180845260005b81811015610de957602081850181015186830182015201610dcd565b506000602082860101526020601f19601f83011685010191505092915050565b88815261010060208201526000610e2461010083018a610dc3565b6040830198909852506060810195909552608085019390935260a08401919091526001600160a01b031660c0830152151560e09091015292915050565b600060208284031215610e7357600080fd5b610e7c82610d64565b9392505050565b600080600060608486031215610e9857600080fd5b83359250602084013567ffffffffffffffff811115610eb657600080fd5b610ec286828701610c7b565b93969395505050506040919091013590565b600181811c90821680610ee857607f821691505b602082108103610f0857634e487b7160e01b600052602260045260246000fd5b50919050565b601f821115610b8a57806000526020600020601f840160051c81016020851015610f355750805b601f840160051c820191505b81811015610f555760008155600101610f41565b5050505050565b815167ffffffffffffffff811115610f7657610f76610c65565b610f8a81610f848454610ed4565b84610f0e565b6020601f821160018114610fbe5760008315610fa65750848201515b600019600385901b1c1916600184901b178455610f55565b600084815260208120601f198516915b82811015610fee5787850151825560209485019460019092019101610fce565b508482101561100c5786840151600019600387901b60f8161c191681555b50505050600190811b01905550565b634e487b7160e01b600052601160045260246000fd5b818103818111156110445761104461101b565b92915050565b80820281158282048414176110445761104461101b565b60008261107e57634e487b7160e01b600052601260045260246000fd5b500490565b808201808211156110445761104461101b565b6000602082840312156110a857600080fd5b81518015158114610e7c57600080fd5b6020808252601c908201527f55534454207472616e7366657220746f2075736572206661696c656400000000604082015260600190565b6000600160ff1b82016111045761110461101b565b5060000390565b60006020828403121561111d57600080fd5b5051919050565b84815260806020820152600061113d6080830186610dc3565b6040830194909452506001600160a01b03919091166060909101529291505056fea2646970667358221220622013ae607897e52b6504cb63250462abcd2849be6b3bf3a2616da2a79771a164736f6c634300081d003300000000000000000000000036e6504c968f5c2a310b6af7b97bc22cdd3402cc",
}

// BztABI is the input ABI used to generate the binding from.
// Deprecated: Use BztMetaData.ABI instead.
var BztABI = BztMetaData.ABI

// BztBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BztMetaData.Bin instead.
var BztBin = BztMetaData.Bin

// DeployBzt deploys a new Ethereum contract, binding an instance of Bzt to it.
func DeployBzt(auth *bind.TransactOpts, backend bind.ContractBackend, _usdtToken common.Address, _to common.Address) (common.Address, *types.Transaction, *Bzt, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BztBin), backend, _usdtToken, _to)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Bzt{BztCaller: BztCaller{contract: contract}, BztTransactor: BztTransactor{contract: contract}, BztFilterer: BztFilterer{contract: contract}}, nil
}

// Bzt is an auto generated Go binding around an Ethereum contract.
type Bzt struct {
	BztCaller     // Read-only binding to the contract
	BztTransactor // Write-only binding to the contract
	BztFilterer   // Log filterer for contract events
}

// BztCaller is an auto generated read-only Go binding around an Ethereum contract.
type BztCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BztTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BztTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BztFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BztFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BztSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BztSession struct {
	Contract     *Bzt              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BztCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BztCallerSession struct {
	Contract *BztCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BztTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BztTransactorSession struct {
	Contract     *BztTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BztRaw is an auto generated low-level Go binding around an Ethereum contract.
type BztRaw struct {
	Contract *Bzt // Generic contract binding to access the raw methods on
}

// BztCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BztCallerRaw struct {
	Contract *BztCaller // Generic read-only contract binding to access the raw methods on
}

// BztTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BztTransactorRaw struct {
	Contract *BztTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBzt creates a new instance of Bzt, bound to a specific deployed contract.
func NewBzt(address common.Address, backend bind.ContractBackend) (*Bzt, error) {
	contract, err := bindBzt(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bzt{BztCaller: BztCaller{contract: contract}, BztTransactor: BztTransactor{contract: contract}, BztFilterer: BztFilterer{contract: contract}}, nil
}

// NewBztCaller creates a new read-only instance of Bzt, bound to a specific deployed contract.
func NewBztCaller(address common.Address, caller bind.ContractCaller) (*BztCaller, error) {
	contract, err := bindBzt(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BztCaller{contract: contract}, nil
}

// NewBztTransactor creates a new write-only instance of Bzt, bound to a specific deployed contract.
func NewBztTransactor(address common.Address, transactor bind.ContractTransactor) (*BztTransactor, error) {
	contract, err := bindBzt(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BztTransactor{contract: contract}, nil
}

// NewBztFilterer creates a new log filterer instance of Bzt, bound to a specific deployed contract.
func NewBztFilterer(address common.Address, filterer bind.ContractFilterer) (*BztFilterer, error) {
	contract, err := bindBzt(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BztFilterer{contract: contract}, nil
}

// bindBzt binds a generic wrapper to an already deployed contract.
func bindBzt(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bzt *BztRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bzt.Contract.BztCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bzt *BztRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bzt.Contract.BztTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bzt *BztRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bzt.Contract.BztTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bzt *BztCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bzt.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bzt *BztTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bzt.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bzt *BztTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bzt.Contract.contract.Transact(opts, method, params...)
}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_Bzt *BztCaller) GetContractBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bzt.contract.Call(opts, &out, "getContractBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_Bzt *BztSession) GetContractBalance() (*big.Int, error) {
	return _Bzt.Contract.GetContractBalance(&_Bzt.CallOpts)
}

// GetContractBalance is a free data retrieval call binding the contract method 0x6f9fb98a.
//
// Solidity: function getContractBalance() view returns(uint256)
func (_Bzt *BztCallerSession) GetContractBalance() (*big.Int, error) {
	return _Bzt.Contract.GetContractBalance(&_Bzt.CallOpts)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 orderId, string tokenName, uint256 openPrice, uint256 closePrice, int256 profitLoss, uint256 amount, address user, bool isClosed)
func (_Bzt *BztCaller) Orders(opts *bind.CallOpts, arg0 *big.Int) (struct {
	OrderId    *big.Int
	TokenName  string
	OpenPrice  *big.Int
	ClosePrice *big.Int
	ProfitLoss *big.Int
	Amount     *big.Int
	User       common.Address
	IsClosed   bool
}, error) {
	var out []interface{}
	err := _Bzt.contract.Call(opts, &out, "orders", arg0)

	outstruct := new(struct {
		OrderId    *big.Int
		TokenName  string
		OpenPrice  *big.Int
		ClosePrice *big.Int
		ProfitLoss *big.Int
		Amount     *big.Int
		User       common.Address
		IsClosed   bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.OrderId = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TokenName = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.OpenPrice = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.ClosePrice = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.ProfitLoss = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Amount = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.User = *abi.ConvertType(out[6], new(common.Address)).(*common.Address)
	outstruct.IsClosed = *abi.ConvertType(out[7], new(bool)).(*bool)

	return *outstruct, err

}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 orderId, string tokenName, uint256 openPrice, uint256 closePrice, int256 profitLoss, uint256 amount, address user, bool isClosed)
func (_Bzt *BztSession) Orders(arg0 *big.Int) (struct {
	OrderId    *big.Int
	TokenName  string
	OpenPrice  *big.Int
	ClosePrice *big.Int
	ProfitLoss *big.Int
	Amount     *big.Int
	User       common.Address
	IsClosed   bool
}, error) {
	return _Bzt.Contract.Orders(&_Bzt.CallOpts, arg0)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 orderId, string tokenName, uint256 openPrice, uint256 closePrice, int256 profitLoss, uint256 amount, address user, bool isClosed)
func (_Bzt *BztCallerSession) Orders(arg0 *big.Int) (struct {
	OrderId    *big.Int
	TokenName  string
	OpenPrice  *big.Int
	ClosePrice *big.Int
	ProfitLoss *big.Int
	Amount     *big.Int
	User       common.Address
	IsClosed   bool
}, error) {
	return _Bzt.Contract.Orders(&_Bzt.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bzt *BztCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bzt.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bzt *BztSession) Owner() (common.Address, error) {
	return _Bzt.Contract.Owner(&_Bzt.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bzt *BztCallerSession) Owner() (common.Address, error) {
	return _Bzt.Contract.Owner(&_Bzt.CallOpts)
}

// UsdtToken is a free data retrieval call binding the contract method 0xa98ad46c.
//
// Solidity: function usdtToken() view returns(address)
func (_Bzt *BztCaller) UsdtToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bzt.contract.Call(opts, &out, "usdtToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UsdtToken is a free data retrieval call binding the contract method 0xa98ad46c.
//
// Solidity: function usdtToken() view returns(address)
func (_Bzt *BztSession) UsdtToken() (common.Address, error) {
	return _Bzt.Contract.UsdtToken(&_Bzt.CallOpts)
}

// UsdtToken is a free data retrieval call binding the contract method 0xa98ad46c.
//
// Solidity: function usdtToken() view returns(address)
func (_Bzt *BztCallerSession) UsdtToken() (common.Address, error) {
	return _Bzt.Contract.UsdtToken(&_Bzt.CallOpts)
}

// Airdrop is a paid mutator transaction binding the contract method 0x8ba4cc3c.
//
// Solidity: function airdrop(address _to, uint256 _amount) returns()
func (_Bzt *BztTransactor) Airdrop(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.contract.Transact(opts, "airdrop", _to, _amount)
}

// Airdrop is a paid mutator transaction binding the contract method 0x8ba4cc3c.
//
// Solidity: function airdrop(address _to, uint256 _amount) returns()
func (_Bzt *BztSession) Airdrop(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.Contract.Airdrop(&_Bzt.TransactOpts, _to, _amount)
}

// Airdrop is a paid mutator transaction binding the contract method 0x8ba4cc3c.
//
// Solidity: function airdrop(address _to, uint256 _amount) returns()
func (_Bzt *BztTransactorSession) Airdrop(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.Contract.Airdrop(&_Bzt.TransactOpts, _to, _amount)
}

// CloseOrder is a paid mutator transaction binding the contract method 0x10199bdc.
//
// Solidity: function closeOrder(uint256 _orderId, uint256 _openPrice, uint256 _closePrice, string _tokenName) returns()
func (_Bzt *BztTransactor) CloseOrder(opts *bind.TransactOpts, _orderId *big.Int, _openPrice *big.Int, _closePrice *big.Int, _tokenName string) (*types.Transaction, error) {
	return _Bzt.contract.Transact(opts, "closeOrder", _orderId, _openPrice, _closePrice, _tokenName)
}

// CloseOrder is a paid mutator transaction binding the contract method 0x10199bdc.
//
// Solidity: function closeOrder(uint256 _orderId, uint256 _openPrice, uint256 _closePrice, string _tokenName) returns()
func (_Bzt *BztSession) CloseOrder(_orderId *big.Int, _openPrice *big.Int, _closePrice *big.Int, _tokenName string) (*types.Transaction, error) {
	return _Bzt.Contract.CloseOrder(&_Bzt.TransactOpts, _orderId, _openPrice, _closePrice, _tokenName)
}

// CloseOrder is a paid mutator transaction binding the contract method 0x10199bdc.
//
// Solidity: function closeOrder(uint256 _orderId, uint256 _openPrice, uint256 _closePrice, string _tokenName) returns()
func (_Bzt *BztTransactorSession) CloseOrder(_orderId *big.Int, _openPrice *big.Int, _closePrice *big.Int, _tokenName string) (*types.Transaction, error) {
	return _Bzt.Contract.CloseOrder(&_Bzt.TransactOpts, _orderId, _openPrice, _closePrice, _tokenName)
}

// OpenOrder is a paid mutator transaction binding the contract method 0xfd53e1f2.
//
// Solidity: function openOrder(uint256 _orderId, string _tokenName, uint256 _amount) returns()
func (_Bzt *BztTransactor) OpenOrder(opts *bind.TransactOpts, _orderId *big.Int, _tokenName string, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.contract.Transact(opts, "openOrder", _orderId, _tokenName, _amount)
}

// OpenOrder is a paid mutator transaction binding the contract method 0xfd53e1f2.
//
// Solidity: function openOrder(uint256 _orderId, string _tokenName, uint256 _amount) returns()
func (_Bzt *BztSession) OpenOrder(_orderId *big.Int, _tokenName string, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.Contract.OpenOrder(&_Bzt.TransactOpts, _orderId, _tokenName, _amount)
}

// OpenOrder is a paid mutator transaction binding the contract method 0xfd53e1f2.
//
// Solidity: function openOrder(uint256 _orderId, string _tokenName, uint256 _amount) returns()
func (_Bzt *BztTransactorSession) OpenOrder(_orderId *big.Int, _tokenName string, _amount *big.Int) (*types.Transaction, error) {
	return _Bzt.Contract.OpenOrder(&_Bzt.TransactOpts, _orderId, _tokenName, _amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bzt *BztTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bzt.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bzt *BztSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bzt.Contract.RenounceOwnership(&_Bzt.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bzt *BztTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bzt.Contract.RenounceOwnership(&_Bzt.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bzt *BztTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bzt.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bzt *BztSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bzt.Contract.TransferOwnership(&_Bzt.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bzt *BztTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bzt.Contract.TransferOwnership(&_Bzt.TransactOpts, newOwner)
}

// BztAirdropIterator is returned from FilterAirdrop and is used to iterate over the raw logs and unpacked data for Airdrop events raised by the Bzt contract.
type BztAirdropIterator struct {
	Event *BztAirdrop // Event containing the contract specifics and raw log

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
func (it *BztAirdropIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BztAirdrop)
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
		it.Event = new(BztAirdrop)
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
func (it *BztAirdropIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BztAirdropIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BztAirdrop represents a Airdrop event raised by the Bzt contract.
type BztAirdrop struct {
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAirdrop is a free log retrieval operation binding the contract event 0x8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a.
//
// Solidity: event Airdrop(address recipient, uint256 amount)
func (_Bzt *BztFilterer) FilterAirdrop(opts *bind.FilterOpts) (*BztAirdropIterator, error) {

	logs, sub, err := _Bzt.contract.FilterLogs(opts, "Airdrop")
	if err != nil {
		return nil, err
	}
	return &BztAirdropIterator{contract: _Bzt.contract, event: "Airdrop", logs: logs, sub: sub}, nil
}

// WatchAirdrop is a free log subscription operation binding the contract event 0x8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a.
//
// Solidity: event Airdrop(address recipient, uint256 amount)
func (_Bzt *BztFilterer) WatchAirdrop(opts *bind.WatchOpts, sink chan<- *BztAirdrop) (event.Subscription, error) {

	logs, sub, err := _Bzt.contract.WatchLogs(opts, "Airdrop")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BztAirdrop)
				if err := _Bzt.contract.UnpackLog(event, "Airdrop", log); err != nil {
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

// ParseAirdrop is a log parse operation binding the contract event 0x8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a.
//
// Solidity: event Airdrop(address recipient, uint256 amount)
func (_Bzt *BztFilterer) ParseAirdrop(log types.Log) (*BztAirdrop, error) {
	event := new(BztAirdrop)
	if err := _Bzt.contract.UnpackLog(event, "Airdrop", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BztOrderClosedIterator is returned from FilterOrderClosed and is used to iterate over the raw logs and unpacked data for OrderClosed events raised by the Bzt contract.
type BztOrderClosedIterator struct {
	Event *BztOrderClosed // Event containing the contract specifics and raw log

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
func (it *BztOrderClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BztOrderClosed)
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
		it.Event = new(BztOrderClosed)
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
func (it *BztOrderClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BztOrderClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BztOrderClosed represents a OrderClosed event raised by the Bzt contract.
type BztOrderClosed struct {
	OrderId    *big.Int
	ProfitLoss *big.Int
	OpenPrice  *big.Int
	ClosePrice *big.Int
	User       common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOrderClosed is a free log retrieval operation binding the contract event 0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96.
//
// Solidity: event OrderClosed(uint256 orderId, int256 profitLoss, uint256 openPrice, uint256 closePrice, address user)
func (_Bzt *BztFilterer) FilterOrderClosed(opts *bind.FilterOpts) (*BztOrderClosedIterator, error) {

	logs, sub, err := _Bzt.contract.FilterLogs(opts, "OrderClosed")
	if err != nil {
		return nil, err
	}
	return &BztOrderClosedIterator{contract: _Bzt.contract, event: "OrderClosed", logs: logs, sub: sub}, nil
}

// WatchOrderClosed is a free log subscription operation binding the contract event 0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96.
//
// Solidity: event OrderClosed(uint256 orderId, int256 profitLoss, uint256 openPrice, uint256 closePrice, address user)
func (_Bzt *BztFilterer) WatchOrderClosed(opts *bind.WatchOpts, sink chan<- *BztOrderClosed) (event.Subscription, error) {

	logs, sub, err := _Bzt.contract.WatchLogs(opts, "OrderClosed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BztOrderClosed)
				if err := _Bzt.contract.UnpackLog(event, "OrderClosed", log); err != nil {
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

// ParseOrderClosed is a log parse operation binding the contract event 0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96.
//
// Solidity: event OrderClosed(uint256 orderId, int256 profitLoss, uint256 openPrice, uint256 closePrice, address user)
func (_Bzt *BztFilterer) ParseOrderClosed(log types.Log) (*BztOrderClosed, error) {
	event := new(BztOrderClosed)
	if err := _Bzt.contract.UnpackLog(event, "OrderClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BztOrderOpenedIterator is returned from FilterOrderOpened and is used to iterate over the raw logs and unpacked data for OrderOpened events raised by the Bzt contract.
type BztOrderOpenedIterator struct {
	Event *BztOrderOpened // Event containing the contract specifics and raw log

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
func (it *BztOrderOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BztOrderOpened)
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
		it.Event = new(BztOrderOpened)
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
func (it *BztOrderOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BztOrderOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BztOrderOpened represents a OrderOpened event raised by the Bzt contract.
type BztOrderOpened struct {
	OrderId   *big.Int
	TokenName string
	Amount    *big.Int
	User      common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOrderOpened is a free log retrieval operation binding the contract event 0xee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b5.
//
// Solidity: event OrderOpened(uint256 orderId, string tokenName, uint256 amount, address user)
func (_Bzt *BztFilterer) FilterOrderOpened(opts *bind.FilterOpts) (*BztOrderOpenedIterator, error) {

	logs, sub, err := _Bzt.contract.FilterLogs(opts, "OrderOpened")
	if err != nil {
		return nil, err
	}
	return &BztOrderOpenedIterator{contract: _Bzt.contract, event: "OrderOpened", logs: logs, sub: sub}, nil
}

// WatchOrderOpened is a free log subscription operation binding the contract event 0xee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b5.
//
// Solidity: event OrderOpened(uint256 orderId, string tokenName, uint256 amount, address user)
func (_Bzt *BztFilterer) WatchOrderOpened(opts *bind.WatchOpts, sink chan<- *BztOrderOpened) (event.Subscription, error) {

	logs, sub, err := _Bzt.contract.WatchLogs(opts, "OrderOpened")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BztOrderOpened)
				if err := _Bzt.contract.UnpackLog(event, "OrderOpened", log); err != nil {
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

// ParseOrderOpened is a log parse operation binding the contract event 0xee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b5.
//
// Solidity: event OrderOpened(uint256 orderId, string tokenName, uint256 amount, address user)
func (_Bzt *BztFilterer) ParseOrderOpened(log types.Log) (*BztOrderOpened, error) {
	event := new(BztOrderOpened)
	if err := _Bzt.contract.UnpackLog(event, "OrderOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BztOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bzt contract.
type BztOwnershipTransferredIterator struct {
	Event *BztOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BztOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BztOwnershipTransferred)
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
		it.Event = new(BztOwnershipTransferred)
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
func (it *BztOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BztOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BztOwnershipTransferred represents a OwnershipTransferred event raised by the Bzt contract.
type BztOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bzt *BztFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BztOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bzt.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BztOwnershipTransferredIterator{contract: _Bzt.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bzt *BztFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BztOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bzt.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BztOwnershipTransferred)
				if err := _Bzt.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bzt *BztFilterer) ParseOwnershipTransferred(log types.Log) (*BztOwnershipTransferred, error) {
	event := new(BztOwnershipTransferred)
	if err := _Bzt.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
