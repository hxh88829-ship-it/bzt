package bzt

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
	"os"
	"testing"
	"valueguard/internal/api"
	"valueguard/internal/conf"
)

const (
	rpcUrl = ""
)

func TestBztCaller_Owner(t *testing.T) {
	os.Setenv("RpcUrl", "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	os.Setenv("ContractAddr", "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a")

	ContractBztAddr := os.Getenv("ContractBztAddr")
	if ContractBztAddr == "" {
		return
	}
	conf.ContractBztAddr = ContractBztAddr
	//初始化节点
	conf.RpcUrl = os.Getenv("RpcUrl")
	if conf.RpcUrl == "" {
		return
	}
	log.Info("rpc url is ", conf.RpcUrl)
	var err error
	api.Client, err = InitEthClient(conf.RpcUrl)
	if err != nil {
		return
	}
	defer api.Client.Close()
	res, err := GetOwner()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(res)
}

func TestBztCaller_GetContractBalance(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	cli := api.Client
	defer cli.Close()
	ba, err := GetTokenBalance()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(ba.String())
}

func TestBztCaller_UsdtToken(t *testing.T) {
	os.Setenv("RpcUrl", "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	os.Setenv("ContractBztAddr", "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a")

	ContractBztAddr := os.Getenv("ContractBztAddr")
	if ContractBztAddr == "" {
		log.Fatal("ContractBztAddr is empty")
		return
	}
	conf.ContractBztAddr = ContractBztAddr
	//初始化节点
	conf.RpcUrl = os.Getenv("RpcUrl")
	if conf.RpcUrl == "" {
		return
	}
	log.Info("rpc url is ", conf.RpcUrl)
	var err error
	api.Client, err = InitEthClient(conf.RpcUrl)
	if err != nil {
		return
	}
	defer api.Client.Close()
	res, err := GetUsdToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	rel, err := api.GetTransactionReceiptByHash(common.HexToHash("0x2539f11029a1815e1f4b7753a166ef4ab235608ba861ae918a4afe85492603dd"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rel.Status)
	t.Log(res)
}

func TestBztCaller_Orders(t *testing.T) {
	//配置变量
	os.Setenv("RpcUrl", "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	os.Setenv("ContractAddr", "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a")

	ContractBztAddr := os.Getenv("ContractBztAddr")
	if ContractBztAddr == "" {
		return
	}
	conf.ContractBztAddr = ContractBztAddr
	//初始化节点
	conf.RpcUrl = os.Getenv("RpcUrl")
	if conf.RpcUrl == "" {
		return
	}
	log.Info("rpc url is ", conf.RpcUrl)
	var err error
	api.Client, err = InitEthClient(conf.RpcUrl)
	if err != nil {
		return
	}
	res, err := GetOrders(6)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(res)
}

func TestBztTransactor_OpenOrder(t *testing.T) {
	Client, err := ethclient.Dial("http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer Client.Close()
	con := common.HexToAddress("0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a")
	ca, err := NewBztTransactor(con, Client)
	if err != nil {
		t.Fatal(err)
		return
	}
	pri, err := crypto.HexToECDSA("f56336cb10bf15d0a7a4466c62b8f84c2b4d8a75c5580db0332d69f0d3efa0c3")
	if err != nil {
		return
	}
	opt, err := bind.NewKeyedTransactorWithChainID(pri, big.NewInt(10086))
	if err != nil {
		return
	}

	tx, err := ca.OpenOrder(opt, big.NewInt(1958181354344546304), "BTCUSDT", big.NewInt(1e+6))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tx.Hash())
	//0x872fb9fb2ff99cbd9a7a4bcb05a3e553d93087a2d64c7e96c2a2461b1ecf9a39
	//0x9c0d3ee1d8c5f29980638ea8eed41391fa99f71c6f1fc495ef9e4dc877b46bfe fail

}

func TestBztTransactor_CloseOrder(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	cli := api.Client
	defer cli.Close()
	tx, err := GetCloseOrder(1955455781419614208, 119234400000, 119216750000, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(tx)
	//0.00000838684138
	//0xc72a69de09276336cb9b17f1a0ced56354a06ff3528a0e810a20069c4728c48a
}

func TestBztTransactor_Airdrop(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	cli := api.Client
	defer cli.Close()
	i := new(big.Int)
	if _, ok := i.SetString("3", 10); !ok {
		return
	}
	t.Log(i, "\n")
	tx, err := GetAirdrop("", i.Int64())
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(tx)
	//0x057eb086df137f5e846451aba3cca59c0ed7c7681526412fbee79afa05c984de
}

func TestBztFilterer_ParseOrderClosed(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer api.Client.Close()
	cli := api.Client
	receipt, err := cli.TransactionReceipt(context.Background(), common.HexToHash("0x414a9c3476a1c8063179850e7123e28262e65570f0aca88c9ea88d9d8512c076"))
	if err != nil {
		return
	}
	res, err := GetParseOrderClosed(receipt)
	if err != nil {
		t.Log(1, err.Error())
	}
	if res == nil {
		t.Log(1, "parse order closed")
	}
	t.Log(res.Raw.BlockNumber)

}

func TestBztFilterer_ParseOrderOpened(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer api.Client.Close()
	cli := api.Client
	receipt, err := cli.TransactionReceipt(context.Background(), common.HexToHash("0x13e45d61e4302e896d4e030b6d3f6adf26aa11b16302b44fa6c0657d8afcd5c1"))
	if err != nil {
		return
	}

	res, err := GetParseOrderOpened(receipt)
	if err != nil {
		return
	}
	t.Log(res)

}

func TestBztFilterer_ParseAirdrop(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer api.Client.Close()
	cli := api.Client
	receipt, err := cli.TransactionReceipt(context.Background(), common.HexToHash("0x057eb086df137f5e846451aba3cca59c0ed7c7681526412fbee79afa05c984de"))
	if err != nil {
		return
	}
	res, err := GetParseAirdrop(receipt)
	if err != nil {
		return
	}
	t.Log(res)
}

func TestName(t *testing.T) {
	var err error
	api.Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	cli := api.Client
	defer cli.Close()
	pri, err := crypto.HexToECDSA("272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4")
	if err != nil {
		return
	}
	opt, err := bind.NewKeyedTransactorWithChainID(pri, big.NewInt(10086))
	if err != nil {
		return
	}
	t.Log(opt)
}
