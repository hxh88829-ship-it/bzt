package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
	"valueguard/internal/mongo"
)

type Client struct {
	ApiKey    string
	SecretKey string
	BaseURL   string
	httpCli   *http.Client
}

// 创建带连接池的 Binance Client
func NewClient(apiKey, secretKey string, isTest bool) *Client {
	base := "https://api.binance.com"
	if isTest {
		base = "https://testnet.binance.vision"
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	}

	httpCli := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return &Client{
		ApiKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   base,
		httpCli:   httpCli,
	}
}

// 生成签名
func (c *Client) sign(params url.Values) string {
	mac := hmac.New(sha256.New, []byte(c.SecretKey))
	mac.Write([]byte(params.Encode()))
	return hex.EncodeToString(mac.Sum(nil))
}

// 发起签名请求
func (c *Client) request(method, endpoint string, params url.Values, signed bool) ([]byte, error) {
	if signed {
		params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))
		sign := c.sign(params)
		params.Set("signature", sign)
	}

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, endpoint, params.Encode())
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", c.ApiKey)
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// 上面属于定义client连接池复用，

// 下单（支持 LIMIT / MARKET）
func (c *Client) CreateOrder(symbol, side, orderType, quantity string) ([]byte, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", side)
	params.Set("type", orderType)
	params.Set("quantity", quantity)
	// 限价单需 price + timeInForce
	if orderType == "LIMIT" {
		//params.Set("price", price)
		params.Set("timeInForce", "GTC")
	}
	return c.request("POST", "/api/v3/order", params, true)
}

// 撤单
func (c *Client) CancelOrder(symbol, orderId string) ([]byte, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderId)
	return c.request("DELETE", "/api/v3/order", params, true)
}

// 查询单笔订单
func (c *Client) GetOrder(symbol, orderId string) ([]byte, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderId)
	return c.request("GET", "/api/v3/order", params, true)
}

// 查询所有订单（历史 + 当前）
func (c *Client) GetAllOrders(symbol string, limit int) ([]byte, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return c.request("GET", "/api/v3/allOrders", params, true)
}

// 查询成交历史
func (c *Client) GetMyTrades(symbol string, limit int) ([]byte, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return c.request("GET", "/api/v3/myTrades", params, true)
}

// 查询账户余额信息
func (c *Client) GetAccountBalances(asset string, test bool) ([]BinanceBalance, error) {
	params := url.Values{}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))

	var raw []byte
	var err error
	if test {
		raw, err = c.request("GET", "/api/v3/account", params, true)
	} else {
		if asset != "" {
			params.Set("asset", asset)
		}
		raw, err = c.request("POST", "/sapi/v3/asset/getUserAsset", params, true)
	}
	if err != nil {
		return nil, err
	}

	balances, err := ParseAccountInfo(raw)
	if err != nil {
		// 可以记录 raw 以便排查
		log.Errorf("ParseAccountInfo failed: %v, raw: %s", err, string(raw))
		return nil, err
	}
	return balances, nil
}

// ParseBinanceOrder 将币安返回的 JSON 解析为结构体
func ParseBinanceOrder(data []byte) (*mongo.BinanceOrder, error) {
	var order mongo.BinanceOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("解析币安订单失败: %v", err)
	}
	return &order, nil
}

type BinanceBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// 测试网返回带 balances 字段
type BinanceAccount struct {
	MakerCommission  int              `json:"makerCommission"`
	TakerCommission  int              `json:"takerCommission"`
	BuyerCommission  int              `json:"buyerCommission"`
	SellerCommission int              `json:"sellerCommission"`
	Balances         []BinanceBalance `json:"balances"`
}

func ParseAccountInfo(data []byte) ([]BinanceBalance, error) {
	// 尝试解析为测试网格式（带 balances）
	var acc BinanceAccount
	if err := json.Unmarshal(data, &acc); err == nil && len(acc.Balances) > 0 {
		return acc.Balances, nil
	}

	// 尝试解析为主网格式（直接是数组）
	var balances []BinanceBalance
	if err := json.Unmarshal(data, &balances); err == nil && len(balances) > 0 {
		return balances, nil
	}

	return nil, fmt.Errorf("未能识别账户返回格式: %s", string(data))
}
