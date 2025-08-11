package api

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"testing"
	"time"
)

const (
	rpcUrl = "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979"
	//addrContract = "0xa0fA4D216AAc046b6B3f8fae4869FFC7Da5B2BBa" //BVToken
	userAddr  = "0xc020e62ce44297e86dA12CF15CfDc20B83eF3b72"
	userAddr2 = "0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24"
	//key       = "272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4"
	key    = "f56336cb10bf15d0a7a4466c62b8f84c2b4d8a75c5580db0332d69f0d3efa0c3"
	txHash = "0x668da65eff65b2dd4b801e55390dab2aba1e84e66f499de903ab82e49ae1b572" //contract
	//txHash = "0x1a532096b867f38d165551aaa0a099644b2e48f9eb25a42f5e51098acabf8788" //public
	txhashnft = "0xe264acae247fb5977a58884aa5b3c85879e7862d7a69142b949d08eead710b26"
)

func TestGetBlockByNumber(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()

	//bl, err := GetBlockByNumber(7339780)
	bl, err := GetBlockByNumber(7453707)
	if err != nil {
		t.Error("GetBlockByNumbr fail")
		return
	}
	t.Log(bl.Nonce())
	t.Log(len(bl.Transactions()))
	t.Log(bl.Transactions())
	t.Log(bl.Transaction(common.HexToHash(txHash)))
	t.Log(bl.Time())

}

func TestGetTransactionByHash(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	tx, ok, err := GetTransactionByHash(common.HexToHash(txhashnft))
	if err != nil {
		t.Error("GetTransactionByHash fail")
		return
	}

	if ok {
		t.Log("pending")
		return
	}

	//t.Log(tx.Cost())
	t.Log(tx.Value())
	//t.Log(tx.Gas())
	//t.Log(tx.GasPrice())
	//t.Log(tx.Time())
	t.Log(hexutil.Encode(tx.Data()))
	if len(tx.Data()) == 0 {
		t.Error("GetTransactionByHash fail")
	}
	t.Log(tx.To().String())

}

func TestGetTransactionReceiptByHash(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	tx, err := GetTransactionReceiptByHash(common.HexToHash("0xe1bbd1fa7f9e644ba25df7211836961e9f4df85809ab80eac5d176838aa7e9e8"))
	if err != nil {
		t.Error("GetTransactionReceiptByHash fail")
		return
	}
	t.Log(tx.TxHash)
	t.Log(tx.BlockNumber)
	t.Log(tx.CumulativeGasUsed)
	t.Log(tx.GasUsed)
	t.Log(tx.ContractAddress)
	if tx.Status == 0 {
		t.Error("GetTransactionReceiptByHash fail")
		return
	}
	t.Log(tx.EffectiveGasPrice)
	t.Log(tx.Logs)

}

func TestGetBalanceByAddress(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	//bl, err := GetBalanceByAddress(userAddr2)
	//if err != nil {
	//	t.Error("GetBalanceByAddress fail")
	//	return
	//}
	//t.Log(bl)
	res, err := GetTokenBalance(context.Background(), userAddr2, "DUSDT")
	if err != nil {
		t.Error("GetTokenBalance fail")
		return
	}
	t.Log(res)

	//1:9690217625996000000  499992000000
	//2:29166994011000000  8000000
}

func TestGetBlockNumber(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	for i := 1; i > 0; i++ {
		num, err := GetBlockNumber()
		if err != nil {
			t.Error("GetBlockNumber fail")
			return
		}
		t.Log(num)
		t.Log(time.Now().Unix())
		time.Sleep(time.Second * 3)
	}
	//ba, err := Client.BalanceAt(context.Background(), common.HexToAddress(userAddr2), new(big.Int).SetUint64(6345129))
	//if err != nil {
	//	t.Error("GetBalanceByAddress fail")
	//	return
	//}
	//t.Log(ba.String())
}

func TestGetFromByTransaction(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	id, err := Client.ChainID(context.Background())
	if err != nil {
		t.Error("Client.ChainID", "err", err)
		return
	}
	ChainId = id.Uint64()
	//ChainId = 10086
	defer Client.Close()

	tx, _, err := Client.TransactionByHash(context.Background(), common.HexToHash(txhashnft))
	if err != nil {
		t.Error("TransactionByHash fail")
		return
	}
	from, err := GetFromByTransaction(tx)
	if err != nil {
		t.Error("GetFromByTransaction fail")
		return
	}
	t.Log(from)
	t.Log(hexutil.Encode(tx.Data()))
	t.Log(tx.To().String())
}

func TestSendTransaction(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	va, err := ValuePow(0.001)
	if err != nil {
		t.Error("GetBalanceByAddress fail")
		return
	}
	tx, err := SendTransaction(key,
		userAddr,
		Client,
		21000,
		common.HexToAddress(userAddr2),
		va)
	if err != nil {
		t.Error("SendTransaction fail")
		return
	}
	t.Log(tx)
}

// 合约交易
func TestErc20Caller_Name(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	na, err := Erc20Caller_Name()
	if err != nil {
		t.Error("Erc20Caller_Name fail")
		return
	}
	t.Log(na)
}

func TestErc20Transactor_BalanceOf(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	bl, err := Erc20Transactor_BalanceOf(userAddr2)
	if err != nil {
		t.Error("Erc20Transactor_BalanceOf fail")
		return
	}
	t.Log(bl)
}

func TestErc20Transactor_Decimals(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	de, err := Erc20Transactor_Decimals()
	if err != nil {
		t.Error("Erc20Transactor_Decimals fail")
		return
	}
	t.Log(de)
}

func TestErc20Transactor_Symbol(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	s, err := Erc20Transactor_Symbol()
	if err != nil {
		t.Error("Erc20Transactor_Symbol fail")
		return
	}
	t.Log(s)
}

func TestErc20Transactor_TotalSupply(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	Ts, err := Erc20Transactor_TotalSupply()
	if err != nil {
		t.Error("Erc20Transactor_TotalSupply fail")
		return
	}
	t.Log(Ts)
}

func TestErc20Transactor_Transfer(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	i, err := ValuePow(0.001)
	if err != nil {
		t.Error("GetBalanceByAddress fail")

		return
	}

	tx, err := Erc20Transactor_Transfer(key, userAddr2, i)
	if err != nil {
		t.Error("Erc20Transactor_Transfer fail")
		return
	}
	t.Log(tx)
}

func TestValuePow(t *testing.T) {
	i, err := ValuePow(0.001)
	if err != nil {
		t.Error("valuePow fail")
		return
	}
	t.Log(i)

	//new(big.Int).SetString(in.Value, 0)
}

func TestName(t *testing.T) {

	a := new(big.Int).SetUint64(9)
	t.Log(a.String())
}

func TestDirectLogValue(t *testing.T) {
	var err error
	Client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()

	tr, err := DirectLogValue("0x057eb086df137f5e846451aba3cca59c0ed7c7681526412fbee79afa05c984de")
	if err != nil {
		t.Error("DirectLogValue fail:", err)
		return
	}
	t.Log(tr)
}

func TestStringToBigIntSum(t *testing.T) {
	//res, err := StringToBigIntDiv("12", "6")
	//if err != nil {
	//	t.Error("StringToBigIntSum fail")
	//	return
	//}
	res, err := StringToBigIntSum("12", "6")
	if err != nil {
		t.Error("StringToBigIntSum fail")
		return
	}
	t.Log(res)
}
