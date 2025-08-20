package mongo

import (
	"strings"
	"testing"
)

const (
	dbUrl = "mongodb://admin:admin@localhost:27017/?directConnection=true"
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
