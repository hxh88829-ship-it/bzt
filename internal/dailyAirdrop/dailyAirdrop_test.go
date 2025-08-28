package dailyAirdrop

import (
	"testing"
	"valueguard/internal/mongo"
)

func TestGetAirdropByDay(t *testing.T) {
	cli, err := mongo.NewMongoClient("")
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	mongo.MonCli = cli
	//var res []string
	//value := append(res, "BTCUSDT")
	//err = GetAirdropByDay(value,)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
}

func TestAddRewardsToPool(t *testing.T) {
	cli, err := mongo.NewMongoClient("")
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	mongo.MonCli = cli
	//var res []string
	//value := append(res, "BTCUSDT")
	//err = AddRewardsToPool(value)
	//
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

}
