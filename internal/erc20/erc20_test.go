package erc20

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	"testing"
)

const (
	rpcUrl       = ""
	addrContract = "" //dUSDToken
)

func TestErc20Caller_Name(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}

	defer cli.Close()
	bal, err := cli.BalanceAt(context.Background(), common.HexToAddress(""), nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(bal)
	con := common.HexToAddress(addrContract)

	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}

	na, err := ca.Name(&opt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(na)
}

func TestErc20Caller_BalanceOf(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	uAddr := common.HexToAddress("")
	ba, err := ca.BalanceOf(&opt, uAddr)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ba)
	//6750000   499992000000
}

func TestErc20Transactor_Decimals(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	de, err := ca.Decimals(&opt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(de)
}

func TestErc20Transactor_Symbol(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	sa, err := ca.Symbol(&opt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(sa)
}

func TestErc20Transactor_TotalSupply(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	ts, err := ca.TotalSupply(&opt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ts)
}

func TestErc20Transactor_Transfer(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()

	//获取chain ID
	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	//构造签名结构体
	pri, err := crypto.HexToECDSA("")
	if err != nil {
		return
	}
	opts, err := bind.NewKeyedTransactorWithChainID(pri, new(big.Int).SetUint64(chainId.Uint64()))
	if err != nil {
		return
	}

	//构造合约client
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Transactor(con, cli)
	if err != nil {
		t.Error(err)
		return
	}

	//接受地址
	to := common.HexToAddress("")

	value := new(big.Int).Mul(new(big.Int).SetUint64(1), new(big.Int).SetUint64(1e+6))

	tx, err := ca.Transfer(opts, to, value)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tx.Hash().Hex())
}

func TestErc20Caller_Allowance(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Caller(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	ba, err := ca.Allowance(&opt, common.HexToAddress(""), common.HexToAddress(""))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ba)

}

func TestNewErc20Transactor_Approve(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Transactor(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	//获取chain ID
	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	//构造签名结构体
	pri, err := crypto.HexToECDSA("")
	if err != nil {
		return
	}
	opts, err := bind.NewKeyedTransactorWithChainID(pri, new(big.Int).SetUint64(chainId.Uint64()))
	if err != nil {
		return
	}
	va := big.NewInt(3000000)
	tx, err := ca.Approve(opts, common.HexToAddress("0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a"), va)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tx.Hash().Hex())
	//0x30fe44baf4e2e5c1478c67dc0c3d510c7b09ee9dff37c601677577ec336708b5
}

func TestNewErc20Transactor_TransferFrom(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	con := common.HexToAddress(addrContract)
	ca, err := NewErc20Transactor(con, cli)
	if err != nil {
		t.Error(err)
		return
	}
	//获取chain ID
	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	//构造签名结构体
	pri, err := crypto.HexToECDSA("")
	if err != nil {
		return
	}
	opts, err := bind.NewKeyedTransactorWithChainID(pri, new(big.Int).SetUint64(chainId.Uint64()))
	if err != nil {
		return
	}
	tx, err := ca.TransferFrom(opts, common.HexToAddress(""), common.HexToAddress(""), big.NewInt(1000000))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tx.Hash().Hex())
	//0x9e2a6c3a7c1a4c597602f941b78cd2f29a4ca95bc70d3becbaaa0cf10c9bc32d
}
