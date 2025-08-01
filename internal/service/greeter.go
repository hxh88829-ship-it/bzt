package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	v1 "valueguard/api/helloworld/v1"
	"valueguard/internal/biz"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name, Value: in.Value})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello + strconv.FormatUint(g.Value, 10)}, nil
}

func (s *GreeterService) MarketCondition(ctx context.Context, in *v1.MarketConditionRequest) (*v1.MarketConditionReply, error) {
	res, err := GetMarketCondition(in.GetSymbol())
	if err != nil {
		return nil, err
	}
	return &v1.MarketConditionReply{
		Price: res,
	}, nil
}
func GetMarketCondition(symbol string) (string, error) {
	url := "https://api.binance.com/api/v3/ticker/price?symbol=" + symbol
	//设置自己的请求超时（5秒）
	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// ✅ 添加浏览器常用的 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/114.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Price, nil
}
