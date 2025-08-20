package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"regexp"
	"strconv"
	"strings"
	"time"
	v1 "valueguard/api/helloworld/v1"
	"valueguard/internal/api"
	"valueguard/internal/biz"
	"valueguard/internal/bzt"
	"valueguard/internal/dailyAirdrop"
	"valueguard/internal/mongo"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer
	MongoClient *mongo.MongoClient
	NodeClient  *ethclient.Client
	uc          *biz.GreeterUsecase
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

func (s *GreeterService) BindWallet(ctx context.Context, in *v1.BindWalletRequest) (*v1.BindWalletReply, error) {
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	if !isAddress(addr) {
		return &v1.BindWalletReply{Metadata: "invalid address"}, nil
	}
	exists, err := IsWalletBound(addr)
	if err != nil {
		return &v1.BindWalletReply{}, err
	}
	if exists {
		return &v1.BindWalletReply{Metadata: "wallet already bound"}, nil
	}
	uid := api.GenerateUID()
	nonce := api.GenerateUID()
	message := fmt.Sprintf("保值通系统请求绑定你的地址：\n%s\n操作类型: bind_wallet\nNonce: %s\nIssued At: %s", addr, nonce, time.Now().Format(time.RFC3339))
	user := mongo.Users{
		Address:         addr,
		Uid:             uid,
		OriginalMessage: message,
		CreateTimeAt:    time.Now().Unix(),
		Status:          "0",
	}

	err = mongo.AddUser(user)
	if err != nil {
		return nil, err
	}
	return &v1.BindWalletReply{
		Uid:      uid,
		Metadata: message,
		Hash:     hexutil.Encode(api.ComputeMessageHash(message)),
	}, nil
}

func (s *GreeterService) LoginWithWallet(ctx context.Context, in *v1.LoginRequest) (*v1.LoginReply, error) {
	us, err := mongo.GetUser(in.GetAddress())
	if err != nil {
		return nil, err
	}

	addr, err := api.VerifyForAddress(us.OriginalMessage, in.GetSignature())
	if err != nil {
		return nil, errors.New("signature verification failed")
	}
	if addr != us.Address {
		log.Warnf("⚠️ 地址不匹配: 签名地址: %s, 绑定地址: %s", addr, us.Address)
		return nil, errors.New("signature not match with address")
	}

	// 清除已使用的签名
	_ = mongo.UpdateUser(us.Address, "")

	// 生成 JWT

	jwtToken, err := api.GetJwtKey(us.Uid, strings.ToLower(us.Address))
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		Token: jwtToken,
	}, nil
}

func (s *GreeterService) GetLoginMessage(ctx context.Context, in *v1.GetLoginMessageRequest) (*v1.GetLoginMessageReply, error) {
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	if !isAddress(addr) {
		return &v1.GetLoginMessageReply{Metadata: "invalid address"}, nil
	}
	exists, err := IsWalletBound(addr)
	if err != nil {
		return &v1.GetLoginMessageReply{}, err
	}
	if !exists {
		return &v1.GetLoginMessageReply{Metadata: "wallet not exists"}, nil
	}
	nonce := api.GenerateUID()
	message := fmt.Sprintf("保值通系统请求绑定你的地址：\n%s\n操作类型: bind_wallet\nNonce: %s\nIssued At: %s", addr, nonce, time.Now().Format(time.RFC3339))

	//更新用户元数据
	_ = mongo.UpdateUser(strings.ToLower(in.GetAddress()), message)

	return &v1.GetLoginMessageReply{
		Metadata: message,
		Hash:     hexutil.Encode(api.ComputeMessageHash(message)),
	}, nil
}

func (s *GreeterService) WalletBalance(ctx context.Context, in *v1.WalletBalanceRequest) (*v1.WalletBalanceReply, error) {
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	if !isAddress(addr) {
		return &v1.WalletBalanceReply{}, nil
	}

	exists, err := IsWalletBound(addr)
	if err != nil {
		return &v1.WalletBalanceReply{}, err
	}
	if !exists {
		log.Warnf("<addr> no exists: %s", addr)
		return &v1.WalletBalanceReply{}, nil
	}

	// 查询多个代币余额
	symbols := []string{"DTT", "DUSDT"}
	var tokens []*v1.TokenBalance

	for _, symbol := range symbols {
		balance, err := api.GetTokenBalance(ctx, addr, symbol)
		if err != nil {
			log.Warnf("failed to get balance for %s: %v", symbol, err)
			continue
		}
		tokens = append(tokens, &v1.TokenBalance{
			Symbol:  symbol,
			Balance: balance,
		})
	}

	return &v1.WalletBalanceReply{Tokens: tokens}, nil
}

func (s *GreeterService) MarketCondition(ctx context.Context, in *v1.MarketConditionRequest) (*v1.MarketConditionReply, error) {
	symbol := in.GetSymbol()
	startTime := in.GetStartTime()
	endTime := in.GetEndTime()

	if startTime >= endTime {
		return nil, errors.New("invalid time range")
	}

	// 假设你有这样的方法
	prices, err := mongo.GetPriceBySymbol(symbol, startTime, endTime)
	if err != nil {
		log.Warnf("symbol %v --- err: %v", symbol, err)
		return nil, errors.New("failed to get price history")
	}

	var marketPrices []*v1.MarketPrice
	for _, p := range prices {
		marketPrices = append(marketPrices, &v1.MarketPrice{
			Price: p.Price,
			Time:  p.Timestamp,
		})
	}

	return &v1.MarketConditionReply{
		Prices: marketPrices,
	}, nil
}

func (s *GreeterService) OpenOrder(ctx context.Context, in *v1.OpenOrderRequest) (*v1.OpenOrderReply, error) {
	// 1. 参数校验
	if in.GetAddress() == "" || in.GetTimestamp() == 0 || in.GetSymbol() == "" {
		return nil, errors.New("missing required parameters")
	}
	ok, err := IsWalletBound(in.GetAddress())
	if err != nil {
		return nil, err
	}
	if !ok {
		return &v1.OpenOrderReply{
			OrderId: "非平台用户地址",
		}, errors.New("wallet not exists")
	}
	// 不同实例不同节点（0～1023）
	orderId := api.GetSnowflakeID(0)
	res, err := mongo.GetPriceByTimestamp(in.GetTimestamp(), in.GetSymbol())
	if err != nil {
		return nil, err
	}
	count, err := mongo.CountOpenOrdersByAddress(in.GetAddress())
	if err != nil {
		return nil, errors.New("failed to check open order count")
	}

	const maxOpenOrders = 10
	if count >= maxOpenOrders {
		return nil, fmt.Errorf("too many open orders (max %d)", maxOpenOrders)
	}

	order := mongo.Order{
		OrderId:        orderId,
		Symbol:         in.GetSymbol(),
		OpenPrice:      res.Price,
		ClosePrice:     "",
		ProfitLoss:     "",
		Amount:         "",
		UsersAddr:      in.GetAddress(),
		IsClosed:       uint64(0), // 0=未开仓确认（待链上确认）
		OrderStartTime: in.GetTimestamp(),
		OrderEndTime:   0,
		OpenTxHash:     "",
		CloseTxHash:    "",
	}

	err = mongo.AddOrder(order)
	if err != nil {
		log.Errorf("CreateOrder failed: %v", err)
		return nil, errors.New("failed to create order")
	}

	return &v1.OpenOrderReply{
		OrderId: orderId,
	}, nil
}

func (s *GreeterService) CloseOrder(ctx context.Context, in *v1.CloseOrderRequest) (*v1.CloseOrderReply, error) {
	// 1. 参数校验
	if in.GetAddress() == "" || in.GetTimestamp() == 0 || in.GetSymbol() == "" || in.GetOrderId() == "" {
		return nil, errors.New("missing required parameters")
	}
	/* 核对地址 */
	ok, err := IsWalletBound(in.GetAddress())
	if err != nil {
		return nil, err
	}
	if !ok {
		return &v1.CloseOrderReply{}, nil
	}
	res, err := mongo.GetPriceByTimestamp(in.GetTimestamp(), in.GetSymbol())
	if err != nil {
		return nil, err
	}
	UserOrderId, err := mongo.GetOrder(in.GetOrderId())
	if err != nil {
		return nil, err
	}
	if UserOrderId.IsClosed == 0 {
		return &v1.CloseOrderReply{}, errors.New("订单等待链上确认")
	}
	orderId, err := api.StringToBigInt(UserOrderId.OrderId)
	if err != nil {
		return nil, err
	}
	OpenPrice, err := api.StringToBigInt(UserOrderId.OpenPrice)
	if err != nil {
		return nil, err
	}
	ClosePrice, err := api.StringToBigInt(res.Price)
	if err != nil {
		return nil, err
	}
	tx, err := bzt.GetCloseOrder(orderId.Int64(), OpenPrice.Int64(), ClosePrice.Int64(), res.Symbol)
	if err != nil {
		return nil, err
	}
	err = mongo.UpdateOrderClose(orderId.String(), ClosePrice.String(), in.GetTimestamp())
	if err != nil {
		return nil, err
	}
	return &v1.CloseOrderReply{
		Tx: tx.Hash().String(),
	}, nil
}

func (s *GreeterService) GetAirdrop(ctx context.Context, in *v1.GetAirdropRequest) (*v1.GetAirdropReply, error) {
	// 1. 参数校验
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	if !isAddress(addr) {
		return &v1.GetAirdropReply{Status: "invalid address"}, nil
	}
	ok, err := IsWalletBound(in.GetAddress())
	if err != nil {
		return nil, err
	}
	if !ok {
		return &v1.GetAirdropReply{
			Status: "address not exists",
		}, nil
	}
	// 判断是否领取
	if in.GetIsClaims() == 0 {
		return &v1.GetAirdropReply{Status: "no airdrop claimed"}, nil
	}
	if in.GetIsClaims() == 1 {
		today := time.Now().Format("2006-01-02")
		claims, claimed, err := dailyAirdrop.UpdateLossAmount(strings.ToLower(in.GetAddress()), in.GetSymbol()) // 用户今日可领，领后总额
		if err != nil {
			return nil, err
		}
		daily, err := mongo.GetDailyAirdrop(today, in.GetSymbol()) // 今日空投总额
		if err != nil {
			return nil, err
		}
		remain, err := api.StringToBigIntSub(daily.Remain, claims.String()) //空投剩余
		if err != nil {
			return nil, err
		}
		if remain.Sign() < 0 {
			return &v1.GetAirdropReply{Status: "invalid result"}, nil
		}
		err = mongo.QueryAirdropByTimes(today, strings.ToLower(in.GetAddress()))
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				tx, err := bzt.GetAirdrop(strings.ToLower(in.GetAddress()), claims.Int64())
				if err != nil {
					return nil, err
				}
				err = mongo.UpdateUserClaims(in.GetSymbol(), strings.ToLower(in.GetAddress()), claimed)
				if err != nil {
					return nil, err
				}
				err = mongo.UpdateDailyAirdrop(today, in.GetSymbol(), remain.String())
				if err != nil {
					return nil, err
				}
				return &v1.GetAirdropReply{
					Status: "success",
					Value:  claims.String(),
					TxHash: tx.Hash().String(),
				}, nil
			} else {
				return nil, err
			}
		}
	}
	return &v1.GetAirdropReply{Status: "address already claim"}, nil
}

func (s *GreeterService) OrderTrade(ctx context.Context, in *v1.OrderTradeRequest) (*v1.OrderTradeReply, error) {
	// 1. 参数校验
	if in.GetAddress() == "" {
		return nil, errors.New("missing required parameters")
	}
	ok, err := IsWalletBound(in.GetAddress())
	if err != nil {
		return nil, err
	}
	if !ok {
		return &v1.OrderTradeReply{}, errors.New("wallet not exists")
	}
	res, err := mongo.GetOrderForAll(in.GetAddress(), in.GetPage(), in.GetPageSize())
	if err != nil {
		return nil, err
	}
	var result v1.OrderTradeReply
	for _, value := range res {
		var rel v1.OrderDetails
		rel.OrderId = value.OrderId
		rel.Symbol = value.Symbol
		rel.OpenedPrice = value.OpenPrice
		rel.ClosePrice = value.ClosePrice
		rel.ProfitLoss = value.ProfitLoss
		rel.Amount = value.Amount
		rel.UsersAddr = value.UsersAddr
		rel.IsClosed = value.IsClosed
		rel.OrderStartTime = value.OrderStartTime
		rel.OrderEndTime = value.OrderEndTime
		rel.OpenTxHash = value.OpenTxHash
		rel.CloseTxHash = value.CloseTxHash
		result.Result = append(result.Result, &rel)
	}
	return &result, nil
}

func (s *GreeterService) AirdropTrade(ctx context.Context, in *v1.AirdropTradeRequest) (*v1.AirdropTradeReply, error) {
	res, err := mongo.GetAirdropForAll(strings.ToLower(in.GetAddr()))
	if err != nil {
		return nil, err
	}
	var result v1.AirdropTradeReply
	for _, v := range res {
		var rel v1.AirdropDetails
		rel.Symbol = v.Symbol
		rel.Amount = v.Amount
		rel.UsersAddr = v.ToAddr
		rel.Times = v.AirdropTime
		result.Result = append(result.Result, &rel)
	}
	return &result, nil
}

// Check 健康检查
func (s *GreeterService) Health(ctx context.Context, _ *v1.HealthCheckRequest) (*v1.HealthCheckReply, error) {

	return &v1.HealthCheckReply{
		Status: "ok",
	}, nil
}

func IsWalletBound(addr string) (bool, error) {
	addr = strings.ToLower(addr)
	_, err := mongo.GetUser(addr)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, err
}
