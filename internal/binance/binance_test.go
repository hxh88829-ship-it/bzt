package binance

import (
	"fmt"
	"os"
	"testing"
)

func TestBinanceClient_PlaceOrder(t *testing.T) {
	// ✅ 从环境变量读取 API Key
	os.Setenv("BINANCE_API_KEY", "DWYXI7f0iInbW4GbG3L3rvNmu5bSh9y4yyP8UIo5xpz7ZeBvS2a2A11sYK3nzTfg")
	os.Setenv("BINANCE_SECRET_KEY", "SnCU4gpwuVlMDZsTYldBKJJGOIXUZkYBsI8NPR5ctDppqBlK9yTmo27gnUE2HAR1")
	BinanceApikey := os.Getenv("BINANCE_API_KEY")
	BinanceSecretKey := os.Getenv("BINANCE_SECRET_KEY")

	client := NewClient(BinanceApikey, BinanceSecretKey, true) // true=测试网

	// 下单（最低5usdt）
	//orderResp, err := client.CreateOrder("BTCUSDT", "SELL", "MARKET", "0.001")
	//fmt.Println("下单结果:", string(orderResp), err)

	// 查询订单
	orderInfo, err := client.GetAllOrders("BTCUSDT", 5)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("最近订单:", string(orderInfo))
	//查询单笔交易
	//order, _ := client.GetOrder("BTCUSDT", "2568149")
	//fmt.Println("单笔交易:", string(order))
	// 查询成交历史
	//trades, _ := client.GetMyTrades("BTCUSDT", 5)
	//fmt.Println("成交记录:", string(trades))

	// 查询账户余额
	//acc, _ := client.GetAccountInfo("BTC")
	//fmt.Println("账户余额:", string(acc))
	//start := time.Now()
	//acc, err := client.GetAccountInfo("BTC")
	//fmt.Println("耗时:", time.Since(start))
	//if err != nil {
	//	fmt.Println("请求出错:", err)
	//} else {
	//	fmt.Println("返回内容:", string(acc))
	//}
}
