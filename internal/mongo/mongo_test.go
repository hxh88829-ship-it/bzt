package mongo

import (
	"strings"
	"testing"
	"time"
)

const (
	dbUrl = ""
)

// 添加用户·
func TestAddUser(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
}
func TestAddOrder(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var res Order
	res.Symbol = "BTCUSDT"
	err = AddOrder(res)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateOrder(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli

}

func TestGetPriceByTimestamp(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	//res, err := GetPriceByTimestamp(1754027483, "BTCUSDT")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//t.Log(res.Price, res.Timestamp, res.Index, res.Symbol)

}

func TestGetPriceBySymbol(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return

	}
	defer cli.Close()
	MonCli = cli
	res, err := GetPriceBySymbol("BTCUSDT", 1754910845, 1754910856)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestUpdateUserLossAmount(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	addr := strings.ToLower("0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24")
	err = UpdateUserLossAmount("BTCUSDT", addr, "150")
	if err != nil {
		t.Error(err)
		return
	}
	//t.Log(res)
}

func TestUpdateRewardPool(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	err = UpdateRewardPool("BTCUSDT", "7114", "71.14")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetAirdrop(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	res, err := GetAirdrop("0xfffa1424b657a0e809e008d28673f14b2811230fb0eebb5a730ed5aa3629445f")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestGetAirdropForAll(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	res, err := GetAirdropForAll("0x34dc39ff05a10cb21724b477e6f1900fd4d8e72f")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestGetOrderForAll(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	res, err := GetOrderForAll("0x34dc39ff05a10cb21724b477e6f1900fd4d8e72f", 0, 0)

	if err != nil {
		t.Error(err)
		return

	}
	t.Log(res)
}

func TestAddBztDapp(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var a BztDapp
	a.AppId = 1
	a.DappIntroduce = "bzt"
	a.DappIcon = "https://upmpc-test.s3.ap-southeast-1.amazonaws.com/dtc/nft/hx/baozhitong/png/1755742252379_rx38jirb4kj.png"
	a.DappName = "bzt"
	a.DappUrl = "http://13.228.99.71:9015/"
	err = AddBztDapp(a)
	if err != nil {
		return
	}
}

func TestAddScanBlock(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var a ScanBlock
	a.LatestBlock = 11487130
	a.NetWork = 9798
	a.Time = time.Now().Unix()
	err = AddScanBlock(a)
	if err != nil {
		return
	}

}

func TestUpdateBztDapp(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	err = UpdateBztDapp("https://upmpc-test.s3.ap-southeast-1.amazonaws.com/dtc/nft/hx/baozhitong/png/1756194866469_fj47uc5ukam.png", "bzt")
	if err != nil {
		return
	}
}

func TestUpdateScanBlock(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var a ScanBlock
	a.LatestBlock = 9805718
	a.NetWork = 10086
	a.Time = time.Now().Unix()
	err = UpdateScanBlock(a)
	if err != nil {
		return
	}
}

func TestAddOrderSwitch(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var a OrderSwitch
	a.Status = 0
	a.ChainId = 9798
	err = AddOrderSwitch(a)
	if err != nil {
		return
	}
}
func TestGetOrderSwitch(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	var i uint64 = 9798
	res, err := GetOrderSwitch(i)
	if err != nil {
		return
	}
	t.Log(res)
}
