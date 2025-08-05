package monitorBlock

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	"valueguard/internal/api"
)

func TestParseEvents(t *testing.T) {
	var err error
	Client, err := ethclient.Dial("http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	if err != nil {
		t.Error("BLockChain fail")
		return
	}
	defer Client.Close()
	api.Client = Client
	receipt, err := api.GetTransactionReceiptByHash(common.HexToHash("0xa63b3c63131e669dfa803cd882f8852705386a9afaa4675f68305b24ddf6d9ac"))
	if err != nil {
		t.Error(err)
		return
	}
	res, err := ParseEvents(receipt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}
