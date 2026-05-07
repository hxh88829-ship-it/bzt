package marketCondition

import (
	"os"
	"testing"
)

func TestGetMarketCondition(t *testing.T) {
	os.Setenv("BINANCE_API_KEY", "")
	os.Setenv("BINANCE_SECRET_KEY", "")
	BinanceApikey := os.Getenv("BINANCE_API_KEY")
	BinanceSecretKey := os.Getenv("BINANCE_SECRET_KEY")

}
