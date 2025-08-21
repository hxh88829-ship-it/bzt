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
	rpcUrl = ""
	//addrContract = "" //BVToken
	userAddr  = ""
	userAddr2 = ""
	//key       = ""
	key    = ""
	txHash = "" //contract
	//txHash = "" //public
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
	tx, ok, err := GetTransactionByHash(common.HexToHash("0xd5f9ea3192bd1c6031718cfbd88bbf3a0cbe03bf45a9c83d036674ee5373d3c7"))
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
	Client, err = ethclient.Dial("")
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	tx, err := GetTransactionReceiptByHash(common.HexToHash("0xd96b8cc56ceccec490c79b1bb62287a36025017ac57c3aedb36c817091968de6"))
	if err != nil {
		t.Error("GetTransactionReceiptByHash fail")
		return
	}
	t.Log(tx)
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
	//0x570d15793062f6c6d957ea0ad84ec980304f42bac7346bada543683cd20d9552 部署合约hash

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

	tx, _, err := Client.TransactionByHash(context.Background(), common.HexToHash("0xfc97277fb81f841cbdf5ba6bb94e19d37cba537a1d0395fdcbb3dad78abd217b"))
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

	tr, err := DirectLogValue("0xd5f9ea3192bd1c6031718cfbd88bbf3a0cbe03bf45a9c83d036674ee5373d3c7")
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
	res, err := StringToBigIntSub("0", "6")
	if err != nil {
		t.Error("StringToBigIntSum fail")
		return
	}
	t.Log(res)
}

func TestGetCode(t *testing.T) {
	token, err := GetJwtKey("123", "qweasd123")
	if err != nil {
		t.Error("GetJwtKey fail")
		return
	}
	addr, err := ParseJwtAddr(token)
	if err != nil {
		t.Error("ParseJwtAddr fail")
		return
	}
	t.Log(addr)
}
