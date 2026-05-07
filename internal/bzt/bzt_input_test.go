package bzt

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
	"testing"
	"valueguard/internal/conf"
)

const BztTestBin = ""

func TestGetCloseOrderInput(t *testing.T) {
	data, err := GetCloseOrderInput(
		new(big.Int).SetUint64(1111),
		new(big.Int).SetUint64(122),
		new(big.Int).SetUint64(123),
		"BTCUSDT")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(data)
	t.Log(hexutil.Encode(data))
}

func TestGetOpenOrderInput(t *testing.T) {
	data, err := GetOpenOrderInput(
		new(big.Int).SetUint64(111),
		"BTCUSDT",
		new(big.Int).SetUint64(123))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(hexutil.Encode(data))
}

func TestGetAirdropInput(t *testing.T) {
	data, err := GetAirdropInput(common.HexToAddress(""), new(big.Int).SetUint64(111))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(hexutil.Encode(data))
}

func TestUrlGetKeyAddress(t *testing.T) {

	conf.Apikey = ""
	conf.BaseUrl = ""
	conf.KeyId = ""

	res, err := UrlGetKeyAddress()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestUrlContractSignOwner(t *testing.T) {
	//配置变量

	conf.Apikey = ""
	conf.BaseUrl = ""
	conf.KeyId = ""
	conf.OwnerAddress = ""

	conf.HmacKey = ""

	conf.ContractBztAddr = ""
	//初始化节点
	conf.RpcUrl = ""

	log.Info("rpc url is ", conf.RpcUrl)
	var err error
	Client, err := InitEthClient(conf.RpcUrl)
	if err != nil {
		return
	}
	orderId := new(big.Int).SetUint64(5)
	amount := new(big.Int).SetUint64(1000000)
	input, err := GetOpenOrderInput(orderId, "BTCUSDT", amount)
	if err != nil {
		t.Fatalf("GetOpenOrderInput err: %s", err)
		return
	}
	a, b, err := UrlContractSignOwner(input, Client)
	if err != nil {
		t.Fatalf("UrlContractSignOwner err: %s", err)
		return
	}
	t.Log(a)
	t.Log(b)
}

func TestUrlOwnerContractTransfer(t *testing.T) {
	//配置变量

	conf.Apikey = ""
	conf.BaseUrl = ""
	conf.KeyId = ""
	conf.OwnerAddress = ""
	conf.HmacKey = ""
	//初始化节点
	conf.RpcUrl = ""
	var err error
	Client, err := ethclient.Dial(conf.RpcUrl)
	if err != nil {
		return
	}
	defer Client.Close()

	//amount := new(big.Int).SetUint64(116)
	//input, err := GetAirdropInput(common.HexToAddress(""), amount)
	//if err != nil {
	//	t.Fatalf("GetOpenOrderInput err: %s", err)
	//	return
	//}
	input, err := hexutil.Decode(BztTestBin)
	if err != nil {
		t.Fatalf("hexutil.Decode(input) err: %s", err)
		return
	}
	a, b, err := DeployContractTransfer(input, Client)
	if err != nil {
		t.Fatalf("UrlOwnerContractTransfer err: %s", err)
		return
	}
	t.Log(a.Hash().Hex())
	t.Log(b)
}

func TestInitEthClient(t *testing.T) {
	conf.X_Api_Key = ""
	conf.RpcUrl = ""
	cli, err := InitEthClient(conf.RpcUrl)
	if err != nil {
		t.Fatalf("InitEthClient err: %s", err)
		return
	}
	blockNr, err := cli.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("InitEthClient err: %s", err)
		return
	}
	t.Logf("block nr: %d", blockNr)
}
