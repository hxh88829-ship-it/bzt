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
	rpcUrl       = "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979"
	addrContract = "0xaD6780B2A022B79686c5E56017cC4FB8cfCd9726" //dUSDToken
	//addrContract = "0x31f3EB0f255178B0fA3FeCbFe7B5314f38949a4B" //nft交易合约地址
	//addrContract = "0xa0fA4D216AAc046b6B3f8fae4869FFC7Da5B2BBa" //BVToken
	userAddr  = "0xc020e62ce44297e86dA12CF15CfDc20B83eF3b72" //499994000000   9877683599988000000
	userAddr2 = "0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24" //6000000        85388176012000000
	//f56336cb10bf15d0a7a4466c62b8f84c2b4d8a75c5580db0332d69f0d3efa0c3
	//272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4
)

func TestErc20Caller_Name(t *testing.T) {
	cli, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error(err)
		return
	}

	defer cli.Close()
	bal, err := cli.BalanceAt(context.Background(), common.HexToAddress(userAddr2), nil)
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
	uAddr := common.HexToAddress(userAddr2)
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
	pri, err := crypto.HexToECDSA("272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4")
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
	to := common.HexToAddress(userAddr2)

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
	ba, err := ca.Allowance(&opt, common.HexToAddress(userAddr), common.HexToAddress(userAddr2))
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
	pri, err := crypto.HexToECDSA("f56336cb10bf15d0a7a4466c62b8f84c2b4d8a75c5580db0332d69f0d3efa0c3")
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
	pri, err := crypto.HexToECDSA("f56336cb10bf15d0a7a4466c62b8f84c2b4d8a75c5580db0332d69f0d3efa0c3")
	if err != nil {
		return
	}
	opts, err := bind.NewKeyedTransactorWithChainID(pri, new(big.Int).SetUint64(chainId.Uint64()))
	if err != nil {
		return
	}
	tx, err := ca.TransferFrom(opts, common.HexToAddress(userAddr), common.HexToAddress(userAddr2), big.NewInt(1000000))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tx.Hash().Hex())
	//0x9e2a6c3a7c1a4c597602f941b78cd2f29a4ca95bc70d3becbaaa0cf10c9bc32d
}
