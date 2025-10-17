package api

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"log"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	addr, err := VerifyForAddress("保值通系统请求绑定你的地址：\\n0x331e865f47fd1b197d04fe60e45def0c3a1eba24\\n操作类型: bind_wallet\\nNonce: 1e501e02-00da-4f27-a03a-923dbc49c017\\nIssued At: 2025-08-06T15:53:49+08:00",
		"0x40f3af53ff57f7f33428619a80fee45eee0a1bafe37ef0bffb185fe5381f9c605c6625815a5ecb733c3d5a949589c4ddf609b8d714a263bd5638d7e45634466600")

	if err != nil {
		t.Error(err)
		return
	}
	t.Log(addr)
	//保值通系统请求绑定你的地址：
	//0x331e865f47fd1b197d04fe60e45def0c3a1eba24
	//操作类型: bind_wallet
	//Nonce: 1e501e02-00da-4f27-a03a-923dbc49c017
	//Issued At: 2025-08-06T15:53:49+08:00
	//保值通系统请求绑定你的地址：\n0x331e865f47fd1b197d04fe60e45def0c3a1eba24\n操作类型: bind_wallet\nNonce: 1e501e02-00da-4f27-a03a-923dbc49c017\nIssued At: 2025-08-06T15:53:49+08:00
}

func TestSignHash(t *testing.T) {
	res := HashData("0xedb6405bcbeabc9f68c41220e5972973acc1c8959a66fb0ab70c34b14b5d9d5c")
	t.Log(hexutil.Encode(res))
	t.Log(res)
}

func TestOriginalMessage(t *testing.T) {
	OriginalMessage := "opensea.io wants you to sign in with your account:\n" + "000000" +
		"\nClick to sign in and accept the OpenSea Terms of Service (https://opensea.io/tos) and Privacy Policy (https://opensea.io/privacy)." +
		"\nURI: https://opensea.io/zh-CN/" + "000000" +
		"\nVersion: 1" +
		"\nChain ID: 1" +
		"\nNonce:" + "123456789" +
		"\nIssued At:" + time.Now().String()

	data := HashData(OriginalMessage)
	log.Println(hexutil.Encode(data))
}

func TestVerifyMessage(t *testing.T) {
	//uri, err := NormalizeTwitter("1")
	//if err != nil {
	//	t.Error(err)
	//}
	//t.Log(uri)
	//err := ValidateWebsite("https://www.123.com")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//err := ValidateAvatarURL("http://upmpc-test.s3.ap-southeast-1.amazonaws.com/1.png")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

}
