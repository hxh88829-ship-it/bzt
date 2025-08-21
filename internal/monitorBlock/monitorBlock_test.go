package monitorBlock

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	"valueguard/internal/api"
	"valueguard/internal/mongo"
)

// 漏扫块手动复扫
func TestScanOneBlock(t *testing.T) {
	var err error
	Client, err := ethclient.Dial("http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	api.Client = Client
	cli, err := mongo.NewMongoClient("mongodb://admin:admin@localhost:27017/?directConnection=true")
	if err != nil {
		t.Error(err)
		return

	}
	defer cli.Close()
	mongo.MonCli = cli

	lossBl, err := mongo.GetLossBlocksByNetwork(10086)
	if err != nil {
		t.Error(err)
		return
	}
	for _, bl := range lossBl {
		err = ScanOneBlock(context.Background(), bl.BlockNr)
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("block number: %d", bl.BlockNr)
		err = mongo.DeleteLossBlock(bl.BlockNr)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
