package mongo

import (
	"errors"
	"strings"
	"testing"
)

const (
	dbUrl = "mongodb://admin:admin@localhost:27017/?directConnection=true"
)

const (
	rpcUrl = "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979"
	//bvContract = "0xa0fA4D216AAc046b6B3f8fae4869FFC7Da5B2BBa" //BVToken
	userAddr  = "0xc020e62ce44297e86dA12CF15CfDc20B83eF3b72"
	userAddr2 = "0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24"
	key       = "272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4"
	txHash    = "0xd95fc46d72c20555089d617070c12f87205e6b289c7518ec25c9704c2180dbda" //contract
	//txHash = "0xbbed3135c50002823fdf425c23dc05b4e95ae046a80d74c0b43b2faef8f2cde6" //public
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

func TestUpdateLogin(t *testing.T) {
	cli, err := NewMongoClient(dbUrl)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	MonCli = cli
	err = UpdateLogin(strings.ToLower(userAddr2), "coco", "", "", "", "", "")
	if err != nil {
		if errors.Is(err, ErrNoFields) {
			t.Log("<UNK>123")
		} else {
			t.Error(err)
		}
		return
	}
}
