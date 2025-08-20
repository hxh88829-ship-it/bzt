package marketCondition

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"math/big"
	"net/http"
	"time"
	"valueguard/internal/mongo"
)

func GetMarketCondition(symbol string, ind uint64) error {
	apiURL := "https://api.binance.com/api/v3/ticker/price?symbol=" + symbol

	// 创建带超时的上下文（总请求控制在 10 秒内）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构造请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		log.Errorf("🔧 [%s] NewRequest error: %v", symbol, err)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoFetcher/1.0)")

	// 设置本地代理
	//proxyStr := "http://127.0.0.1:7890"
	//proxyURL, err := url.Parse(proxyStr)
	//if err != nil {
	//	log.Errorf("🔧 [%s] Proxy parse error: %v", symbol, err)
	//	return err
	//}
	//
	//client := &http.Client{
	//	Transport: &http.Transport{
	//		Proxy: http.ProxyURL(proxyURL),
	//	},
	//}
	client := http.DefaultClient //默认代理
	//发起请求
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("🔧 [%s] HTTP request error: %v", symbol, err)
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Errorf("🔧 [%s] Unexpected status %d: %s", symbol, resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// 解析响应 JSON
	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Errorf("🔧 [%s] JSON decode error: %v", symbol, err)
		return err
	}

	//log.Infof("✅ [%s] Latest price: %s", result.Symbol, result.Price)
	value, err := ConvertPriceToBigIntString(result.Price, 100)
	if err != nil {
		log.Errorf(" [%s] convert error: %v", symbol, err)
		return err
	}
	// 更新数据库/缓存
	times := uint64(time.Now().Unix())
	if err := UpdateNewPrice(symbol, value, ind, times); err != nil {
		log.Errorf("🔧 [%s] Update price error: %v", symbol, err)
		return err
	}

	return nil
}
func UpdateNewPrice(symbol, newPrice string, ind, times uint64) error {
	_, err := mongo.GetPriceForIndex(symbol, ind)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var price mongo.CoinPrice
			price.Symbol = symbol
			price.Price = newPrice
			price.Timestamp = times
			price.Index = ind
			err = mongo.AddPrice(price)
			if err != nil {
				log.Errorf(" [%s] AddPrice error: %v", symbol, err)
				return err
			}
			return nil
		} else {
			log.Errorf("mongodb other errors [%s] Get new price: %v", symbol, err)
			return err
		}
	}
	err = mongo.SavePrice(symbol, newPrice, ind, times)
	if err != nil {
		log.Errorf("mongodb other errors [%s] Update price error: %v", symbol, err)
		return err
	}
	return nil
}

func ConvertPriceToBigIntString(priceStr string, precision int64) (string, error) {
	// 先用 big.Float 解析字符串
	priceFloat, _, err := big.ParseFloat(priceStr, 10, 256, big.ToNearestEven)
	if err != nil {
		return "", err
	}

	// 乘以精度（转换为整数）
	multiplier := big.NewFloat(float64(precision))
	priceFloat.Mul(priceFloat, multiplier)

	// 转为 big.Int（向下取整）
	priceInt := new(big.Int)
	priceFloat.Int(priceInt)

	// 返回整数字符串形式
	return priceInt.String(), nil
}
