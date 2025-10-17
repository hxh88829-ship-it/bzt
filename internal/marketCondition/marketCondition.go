package marketCondition

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"math/big"
	"net"
	"net/http"
	"time"
	"valueguard/internal/mongo"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second, // 整体超时控制
		Transport: &http.Transport{
			MaxIdleConns:       200,
			MaxConnsPerHost:    100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
)

type KLine struct {
	OpenTime                 int64  `json:"openTime"`
	OpenPrice                string `json:"openPrice"`
	HighPrice                string `json:"highPrice"`
	LowPrice                 string `json:"lowPrice"`
	ClosePrice               string `json:"closePrice"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	NumberOfTrades           int    `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
	Ignore                   string `json:"ignore"`
}

func GetMarketCondition(symbol string, ind uint64) error {
	apiURL := "https://api.binance.com/api/v3/ticker/price?symbol=" + symbol

	// 创建带超时的上下文（总请求控制在 10 秒内）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构造请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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
	//client := http.DefaultClient //默认代理
	//发起请求
	resp, err := httpClient.Do(req)
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
		log.Errorf(" [%s] --[%d] convert error: %v", symbol, ind, err)
		return err
	}
	// 更新数据库/缓存
	times := uint64(time.Now().Unix())
	if err = UpdateNewPrice(symbol, value, ind, times); err != nil {
		log.Errorf("🔧 [%s]--[%d] Update price error: %v", symbol, ind, err)
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
		log.Errorf("[ConvertPriceToBigIntString] convert error: %v", err)
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

func GetKLines(symbol, interval, start, end, limit string) ([]KLine, error) {
	apiURL := "https://api.binance.com/api/v3/klines?symbol=" + symbol +
		"&interval=" + interval +
		"&startTime=" + start +
		"&endTime=" + end +
		"&limit=" + limit

	// 创建带超时的上下文（总请求控制在 10 秒内）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构造请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		log.Errorf("🔧 [%s] NewRequest error: %v", symbol, err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoFetcher/1.0)")

	// 设置本地代理
	//proxyStr := "http://127.0.0.1:7890"
	//proxyURL, err := url.Parse(proxyStr)
	//if err != nil {
	//	log.Errorf("🔧 [%s] Proxy parse error: %v", symbol, err)
	//	return nil, err
	//}
	//
	//client := &http.Client{
	//	Transport: &http.Transport{
	//		Proxy: http.ProxyURL(proxyURL),
	//	},
	//}
	//client := http.DefaultClient //默认代理
	// 发起请求
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("🔧 [%s] HTTP request error: %v", symbol, err)
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Errorf("🔧 [%s] Unexpected status %d: %s", symbol, resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// 解析响应 JSON 为二维数组
	var rawKLines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKLines); err != nil {
		log.Errorf("🔧 [%s] JSON decode error: %v", symbol, err)
		return nil, err
	}

	// 定义结构体来存储每根 K 线的数据

	// 将解析后的数据转换为结构体切片
	var kLines []KLine
	now := time.Now().UnixMilli()
	for _, raw := range rawKLines {
		closeTimeFloat, ok := raw[6].(float64)
		if !ok {
			continue
		}
		closeTime := int64(closeTimeFloat)
		if closeTime > now { // 跳过未闭合 K 线
			continue
		}
		kline := KLine{
			OpenTime:                 int64(raw[0].(float64)),
			OpenPrice:                raw[1].(string),
			HighPrice:                raw[2].(string),
			LowPrice:                 raw[3].(string),
			ClosePrice:               raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                closeTime,
			QuoteAssetVolume:         raw[7].(string),
			NumberOfTrades:           int(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
			Ignore:                   raw[11].(string),
		}
		kLines = append(kLines, kline)
	}

	return kLines, nil
}

// SplitDailyIntervals 取到间隔时间段
func SplitDailyIntervals(start, end, step int64) [][2]int64 {
	var result [][2]int64

	for s := start; s < end; s += step {
		e := s + step - 1
		if e > end {
			e = end
		}
		result = append(result, [2]int64{s, e})
	}
	return result
}

func AddKLineToMongoDB(res []KLine, DataTypes, symbol string) error {
	var rel []mongo.Kline
	for _, kline := range res {
		rel = append(rel, mongo.Kline{
			OpenTime:                 kline.OpenTime,
			OpenPrice:                kline.OpenPrice,
			HighPrice:                kline.HighPrice,
			LowPrice:                 kline.LowPrice,
			ClosePrice:               kline.ClosePrice,
			Volume:                   kline.Volume,
			CloseTime:                kline.CloseTime,
			QuoteAssetVolume:         kline.QuoteAssetVolume,
			NumberOfTrades:           kline.NumberOfTrades,
			TakerBuyBaseAssetVolume:  kline.TakerBuyBaseAssetVolume,
			TakerBuyQuoteAssetVolume: kline.TakerBuyQuoteAssetVolume,
			Ignore:                   kline.Ignore,
			DataType:                 DataTypes,
			Symbol:                   symbol,
		})
	}
	err := mongo.AddKLineData(rel, symbol, DataTypes)
	if err != nil {
		log.Errorf("<AddKLineToMongoDB> [%s]  error: %v", DataTypes, err)
		return err
	}
	return nil
}
