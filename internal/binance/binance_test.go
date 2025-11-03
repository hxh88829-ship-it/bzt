package binance

import (
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
	//orderResp, err := client.CreateMarketOrderByAmount("BTCUSDT", "BUY", "10")
	//fmt.Println("下单结果:", string(orderResp), err)
	//buy, _ := ParseBinanceOrder(orderResp)
	//t.Log(buy)
	// 卖单
	//sells, err := client.SellByBuyOrder("BTCUSDT", 9024478)
	//fmt.Println("<UNK>:", string(sells), err)
	//rell, _ := ParseBinanceOrder(sells)
	//t.Log(rell)
	//8301993 3067940 3069654 4245796 4251005 8301993 8306056
	// 查询订单
	//orderInfo, err := client.GetAllOrders("BTCUSDT", 2)
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Println("最近订单:", string(orderInfo))
	//查询单笔交易
	//order, _ := client.GetOrder("BTCUSDT", "8306056")
	//fmt.Println("单笔交易:", string(order))
	// 查询成交历史
	//trades, _ := client.GetMyTrades("BTCUSDT", 5)
	//fmt.Println("成交记录:", string(trades))

	// 查询账户余额
	acc, _ := client.GetAccountBalances("ETH", true)
	t.Log(acc)

}

//BTC 1.00196000
