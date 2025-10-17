package binanceClient

import (
	"sync"
	"valueguard/internal/binance"
)

var (
	BinanceClient *binance.Client
	once          sync.Once
)

func InitBinanceClient(apiKey, secretKey string, isTestNet bool) {
	once.Do(func() {
		BinanceClient = binance.NewClient(apiKey, secretKey, isTestNet)
	})
}
