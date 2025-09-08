package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"regexp"
	"strconv"
	"strings"
	"time"
	v1 "valueguard/api/helloworld/v1"
	"valueguard/internal/api"
	"valueguard/internal/biz"
	"valueguard/internal/bzt"
	"valueguard/internal/conf"
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
	addr := strings.ToLower(in.GetAddress()) //转小写
	log.Info("BindWallet address: ", in.GetAddress())
	if !isAddress(addr) {
		return &v1.BindWalletReply{Metadata: "invalid address"}, nil
	}
	exists, err := IsWalletBound(addr)
	if err != nil {
		log.Error("BindWallet IsWalletBound:", err)
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
		log.Error("BindWallet AddUser:", err)
		return nil, err
	}
	return &v1.BindWalletReply{
		Uid:      uid,
		Metadata: message,
		Hash:     hexutil.Encode(api.ComputeMessageHash(message)),
	}, nil
}

func (s *GreeterService) LoginWithWallet(ctx context.Context, in *v1.LoginRequest) (*v1.LoginReply, error) {
	us, err := mongo.GetUser(strings.ToLower(in.GetAddress()))
	if err != nil {
		log.Error("LoginWithWallet GetUser:", err)
		return nil, err
	}
	log.Info("LoginWithWallet :", in.GetAddress())
	addr, err := api.VerifyForAddress(us.OriginalMessage, in.GetSignature())
	if err != nil {
		log.Error("LoginWithWallet VerifyForAddress:", err)
		return nil, errors.New("signature verification failed")
	}
	if strings.ToLower(addr) != strings.ToLower(us.Address) {
		log.Warnf("⚠️ 地址不匹配: 签名地址: %s, 绑定地址: %s", addr, us.Address)
		return nil, errors.New("signature not match with address")
	}

	// 清除已使用的签名
	_ = mongo.UpdateUser(strings.ToLower(us.Address), "")
	log.Info("login success: ", addr)
	// 生成 JWT
	jwtToken, err := api.GetJwtKey(us.Uid, strings.ToLower(us.Address), conf.Secret)
	if err != nil {
		log.Error("LoginWithWallet GetJwtKey:", err)
		return nil, err
	}

	return &v1.LoginReply{
		Token: jwtToken,
	}, nil
}

func (s *GreeterService) GetLoginMessage(ctx context.Context, in *v1.GetLoginMessageRequest) (*v1.GetLoginMessageReply, error) {
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	log.Info("GetLoginMessage: ", in.GetAddress())
	if !isAddress(addr) {
		return &v1.GetLoginMessageReply{Metadata: "invalid address"}, nil
	}
	exists, err := IsWalletBound(addr)
	if err != nil {
		log.Error("GetLoginMessage IsWalletBound:", err)
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

	// 提取 addr ， uid
	address, _, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("GetAddrAndUidByToken err: ", err)
		return nil, err
	}
	if strings.ToLower(address) != addr {
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
	//log.Infof("MarketCondition called – symbol: %s, startTime: %d, endTime: %d", symbol, startTime, endTime)
	if startTime >= endTime {
		return nil, errors.New("invalid time range")
	}

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

func (s *GreeterService) GetHistoricalPrice(ctx context.Context, in *v1.GetHistoricalPriceRequest) (*v1.GetHistoricalPriceReply, error) {
	symbol := in.GetSymbol()
	startTime := in.GetStartTime()
	endTime := in.GetEndTime()
	page := in.GetPage()
	size := in.GetSize()
	if startTime >= endTime {
		return nil, errors.New("invalid time range")
	}

	prices, err := mongo.GetPriceBySymbolPaged(symbol, startTime, endTime, page, size)
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

	return &v1.GetHistoricalPriceReply{
		Prices: marketPrices,
	}, nil
}

func (s *GreeterService) OpenOrder(ctx context.Context, in *v1.OpenOrderRequest) (*v1.OpenOrderReply, error) {
	// 1. 参数校验
	if in.GetAddress() == "" || in.GetTimestamp() == 0 || in.GetSymbol() == "" {
		return nil, errors.New("missing required parameters")
	}
	log.Infof(
		"OpenOrder  parameters – address: %q, timestamp: %d, symbol: %q",
		in.GetAddress(),
		in.GetTimestamp(),
		in.GetSymbol(),
	)
	// 提取 addr ， uid
	addr, uid, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("OpenOrder GetAddrAndUidByToken err: ", err)
		return nil, err
	}
	if strings.ToLower(addr) != strings.ToLower(in.GetAddress()) {
		log.Warnf("[OpenOrder][%q] 地址校验失败: token_addr=%q, req_addr=%q",
			in.GetSymbol(), strings.ToLower(addr), strings.ToLower(in.GetAddress()))
		return &v1.OpenOrderReply{}, err
	}
	// 不同实例不同节点（0～1023）
	orderId := api.GetSnowflakeID(0)
	res, err := mongo.GetPriceByTimestamp(in.GetTimestamp(), in.GetSymbol())
	if err != nil {
		log.Error("OpenOrder GetPriceByTimestamp err: ", err)
		return nil, err
	}
	count, err := mongo.CountOpenOrdersByAddress(strings.ToLower(addr))
	if err != nil {
		log.Error("OpenOrder CountOpenOrdersByAddress err: ", err)
		return nil, errors.New("failed to check open order count")
	}

	const maxOpenOrders = 50
	if count >= maxOpenOrders {
		return nil, fmt.Errorf("too many open orders (max %d)", maxOpenOrders)
	}

	sta, err := mongo.GetOrderSwitch(api.ChainId, "OpenOrder")
	if err != nil {
		log.Error("OpenOrder GetOrderSwitch err: ", err)
		return nil, err
	}
	order := mongo.Order{
		Uid:            uid,
		OrderId:        orderId,
		Symbol:         in.GetSymbol(),
		OpenPrice:      res.Price,
		ClosePrice:     "",
		ProfitLoss:     "",
		Amount:         "",
		UsersAddr:      strings.ToLower(addr),
		IsClosed:       uint64(0), // 0=未开仓确认（待链上确认）
		OrderStartTime: in.GetTimestamp(),
		OrderEndTime:   0,
		OpenTxHash:     "",
		CloseTxHash:    "",
		Status:         sta.Status,
	}
	err = mongo.AddOrder(order)
	if err != nil {
		log.Errorf("CreateOrder failed: %q", err)
		return nil, errors.New("failed to create order")
	}
	log.Infof("OpenOrder  parameters: orderId:%q, sta.Status:%d", orderId, sta.Status)
	return &v1.OpenOrderReply{
		OrderId: orderId,
		Status:  sta.Status,
	}, nil
}

func (s *GreeterService) CloseOrder(ctx context.Context, in *v1.CloseOrderRequest) (*v1.CloseOrderReply, error) {
	// 1. 参数校验
	if in.GetTimestamp() == 0 || in.GetOrderId() == "" {
		return nil, errors.New("missing required parameters")
	}
	log.Infof(
		"CloseOrder  parameters – address: %q, timestamp: %d, OrderId: %q",
		in.GetAddress(),
		in.GetTimestamp(),
		in.GetOrderId(),
	)
	var err error
	// 提取 addr ， uid
	addr, _, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("CloseOrder GetAddrAndUidByToken err: ", err)
		return nil, err
	}
	if strings.ToLower(addr) != strings.ToLower(in.GetAddress()) {
		log.Warnf("[CloseOrder] 地址校验失败: token_addr=%q, req_addr=%q",
			strings.ToLower(addr), strings.ToLower(in.GetAddress()))
		return &v1.CloseOrderReply{}, err
	}
	sta, err := mongo.GetOrderSwitch(api.ChainId, "CloseOrder")
	if err != nil {
		log.Warnf("[CloseOrder] GetOrderSwitch : %v", err)
		return nil, err
	}
	if sta.Status != uint64(0) {
		log.Error("CloseOrder: Abnormal user status ")
		return &v1.CloseOrderReply{}, errors.New("Abnormal user status ")
	}

	//通过关闭订单交易hash是否存在来判断订单是否关闭
	UserOrderId, err := mongo.GetOrder(in.GetOrderId())
	if err != nil {
		return nil, err
	}
	if UserOrderId.CloseTxHash != "" {
		log.Warnf("[CloseOrder][%s] <Repeatedly close>", UserOrderId.Symbol)
		return &v1.CloseOrderReply{}, errors.New("<Repeatedly close>")
	}
	res, err := mongo.GetPriceByTimestamp(in.GetTimestamp(), UserOrderId.Symbol)
	if err != nil {
		log.Error("CloseOrder GetPriceByTimestamp err: ", err)
		return nil, err
	}
	//防止用户重复关仓
	if UserOrderId.IsClosed == 0 {
		return &v1.CloseOrderReply{}, errors.New("订单等待链上确认")
	}
	orderId, err := api.StringToBigInt(UserOrderId.OrderId)
	if err != nil {
		log.Error("CloseOrder StringToBigInt err: ", err)
		return nil, err
	}
	OpenPrice, err := api.StringToBigInt(UserOrderId.OpenPrice)
	if err != nil {
		log.Error("CloseOrder StringToBigInt err: ", err)
		return nil, err
	}
	ClosePrice, err := api.StringToBigInt(res.Price)
	if err != nil {
		log.Error("CloseOrder StringToBigInt err: ", err)
		return nil, err
	}
	log.Infof("orderId:%v,   OpenPrice:%v,   ClosePrice:%v,   UserOrderId.Symbol:%q",
		orderId, OpenPrice, ClosePrice, UserOrderId.Symbol)
	input, err := bzt.GetCloseOrderInput(orderId, OpenPrice, ClosePrice, UserOrderId.Symbol)
	if err != nil {
		log.Error("CloseOrder GetCloseOrderInput err: ", err)
		return nil, err
	}
	tx, nonce, err := bzt.UrlOwnerContractTransfer(input, api.Client)
	if err != nil {
		log.Error("CloseOrder UrlOwnerContractTransfer err: ", err)
		return nil, err
	}
	log.Info(nonce, "\n", tx)

	if tx == "" {
		return &v1.CloseOrderReply{}, errors.New("交易未生成")
	}

	//TODO 是否需要判断返回hash回执是0或1
	err = mongo.UpdateOrderClose(orderId.String(), ClosePrice.String(), strings.ToLower(tx), in.GetTimestamp())
	if err != nil {
		log.Error("CloseOrder UpdateOrderClose err: ", err)
		return nil, err
	}
	return &v1.CloseOrderReply{
		Tx: tx,
	}, nil
}

func (s *GreeterService) GetAirdrop(ctx context.Context, in *v1.GetAirdropRequest) (*v1.GetAirdropReply, error) {
	//TODO
	/*
			1. 加密校验
		    2. 流水号来做后面的维护
		    3. 判断开关
		    4. 判断 是否能够领取，
			5 如果不能领取：领过了，或者没资格领取 （返回）
		    6. 能够领取
	*/
	if in.GetTimestamp() == "" {
		return nil, errors.New("missing required parameters")
	}
	log.Info("GetAirdrop:", in.GetTimestamp())
	// 提取 addr ， uid
	addr, uid, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("GetAirdrop GetAddrAndUidByToken err: ", err)
		return nil, err
	}

	//获取开关
	sta, err := mongo.GetOrderSwitch(api.ChainId, "GetAirdrop")
	if err != nil {
		log.Error("GetAirdrop GetOrderSwitch: ", err)
		return nil, err
	}
	if sta.Status != 0 {
		log.Error("GetOrderSwitch:", sta)
		return &v1.GetAirdropReply{}, errors.New("Abnormal user status ")
	}

	// 判断领取资质
	today := in.GetTimestamp()
	claims, claimed, err := dailyAirdrop.UpdateLossAmount(strings.ToLower(addr), "DUSDT") // 用户今日可领，领后总额
	if err != nil {
		log.Error("GetAirdrop UpdateLossAmount:", err)
		return nil, err
	}
	log.Info("claims:", claimed)

	//通过 UID或者address + 时间 拿到一个唯一的红包订单
	// 生成订单号
	err = mongo.QueryAirdrop(strings.ToLower(addr), today)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			OrderId := api.GetSnowflakeID(1)
			daily, err := mongo.GetDailyAirdrop(today, "DUSDT") // 今日空投总额
			if err != nil {
				log.Error("GetAirdrop GetDailyAirdrop:", err)
				return nil, err
			}
			remain, err := api.StringToBigIntSub(daily.Reward, claims.String()) //空投剩余
			if err != nil {
				log.Error("GetAirdrop StringToBigIntSub:", err)
				return nil, err
			}
			if remain.Sign() < 0 {
				return &v1.GetAirdropReply{Status: "invalid result"}, nil
			}

			input, err := bzt.GetAirdropInput(common.HexToAddress(addr), claims)
			if err != nil {
				log.Error("GetAirdrop GetAirdropInput:", err)
				return nil, err
			}
			tx, nonce, err := bzt.UrlOwnerContractTransfer(input, api.Client)
			if err != nil {
				log.Error("GetAirdrop UrlOwnerContractTransfer err: ", err)
				return nil, err
			}
			log.Info(nonce)

			if tx == "" {
				return &v1.GetAirdropReply{Status: "交易发送失败"}, nil
			}

			var air mongo.Airdrop
			air.OrderId = OrderId
			air.Symbol = "DUSDT"
			air.ToAddr = strings.ToLower(addr)
			air.AirdropTime = today
			air.Status = uint64(0)
			air.Amount = claims.String()
			air.TxHash = strings.ToLower(tx)
			air.Uid = uid

			err = mongo.AddAirdrop(air)
			if err != nil {
				log.Error("GetAirdrop AddAirdrop:", err)
				return nil, err
			}

			err = mongo.UpdateUserClaims("DUSDT", strings.ToLower(addr), claimed)
			if err != nil {
				log.Error("GetAirdrop UpdateUserClaims:", err)
				return nil, err
			}
			err = mongo.UpdateDailyAirdrop(today, "DUSDT", remain.String())
			if err != nil {
				log.Error("GetAirdrop UpdateDailyAirdrop:", err)
				return nil, err
			}

			return &v1.GetAirdropReply{
				Status:  "success",
				Value:   claims.String(),
				TxHash:  tx,
				OrderId: OrderId,
			}, nil
		} else {
			log.Error("QueryAirdropByTimes:", err)
			return nil, err
		}
	}

	return &v1.GetAirdropReply{Status: "address already claim"}, nil
}

func (s *GreeterService) OrderTrade(ctx context.Context, in *v1.OrderTradeRequest) (*v1.OrderTradeReply, error) {
	// 提取 addr ， uid
	addr, _, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("GetAddrAndUidByToken err: ", err)
		return nil, err
	}
	log.Info("OrderTrade GetAddrAndUidByToken:", addr)
	if strings.ToLower(addr) != strings.ToLower(in.GetAddress()) {
		log.Warnf("[OrderTrade]: token_addr=%s, req_addr=%s", addr, in.GetAddress())
		return &v1.OrderTradeReply{}, err
	}
	// 查询订单记录
	res, err := mongo.GetOrderForAll(strings.ToLower(addr), in.GetPage(), in.GetPageSize())
	if err != nil {
		log.Error("OrderTrade GetOrderForAll err: ", err)
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
		rel.Uid = value.Uid
		result.Result = append(result.Result, &rel)
	}
	return &result, nil
}

func (s *GreeterService) AirdropTrade(ctx context.Context, in *v1.AirdropTradeRequest) (*v1.AirdropTradeReply, error) {
	// 提取 addr ， uid
	addr, _, err := GetAddrAndUidByToken(ctx)
	if err != nil {
		log.Error("GetAddrAndUidByToken err: ", err)
		return nil, err
	}
	log.Info("AirdropTrade GetAddrAndUidByToken:", addr)
	if strings.ToLower(addr) != strings.ToLower(in.GetAddr()) {
		log.Warnf("[OrderTrade]: token_addr=%s, req_addr=%s", addr, in.GetAddr())
		return &v1.AirdropTradeReply{}, err
	}
	//查询空投记录
	res, err := mongo.GetAirdropForAll(strings.ToLower(in.GetAddr()))
	if err != nil {
		log.Error("AirdropTrade GetAirdropForAll err: ", err)
		return nil, err
	}
	var result v1.AirdropTradeReply
	for _, v := range res {
		var rel v1.AirdropDetails
		rel.Symbol = v.Symbol
		rel.Amount = v.Amount
		rel.UsersAddr = v.ToAddr
		rel.Times = v.AirdropTime
		rel.Uid = v.Uid
		rel.OrderId = v.OrderId
		rel.Status = v.Status
		rel.TxHash = v.TxHash
		result.Result = append(result.Result, &rel)
	}
	return &result, nil
}

func (s *GreeterService) Health(ctx context.Context, _ *v1.HealthCheckRequest) (*v1.HealthCheckReply, error) {

	return &v1.HealthCheckReply{
		Status: "ok",
	}, nil
}

func (s *GreeterService) BztDapp(ctx context.Context, _ *v1.BztDappRequest) (*v1.BztDappReply, error) {
	res, err := mongo.GetBztDapp("bzt")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var a mongo.BztDapp
			a.AppId = 1
			a.DappIntroduce = "bzt"
			a.DappIcon = ""
			a.DappName = "bzt"
			a.DappUrl = ""
			err := mongo.AddBztDapp(a)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var rel v1.DataDetails
	rel.AppId = res.AppId
	rel.DappIntroduce = res.DappIntroduce
	rel.DappIcon = res.DappIcon
	rel.DappName = res.DappName
	rel.DappUrl = res.DappUrl
	var result v1.BztDappReply
	result.Data = append(result.Data, &rel)
	return &v1.BztDappReply{
		Code:    200,
		Message: "success",
		Data:    result.Data,
	}, nil
}

func (s *GreeterService) DeployContract(ctx context.Context, in *v1.DeployContractRequest) (*v1.DeployContractReply, error) {
	const BztProduceBin = "0x608060405234801561001057600080fd5b506040516112ab3803806112ab83398101604081905261002f916100d8565b338061005557604051631e4fbdf760e01b81526000600482015260240160405180910390fd5b61005e81610088565b5060018055600280546001600160a01b0319166001600160a01b0392909216919091179055610108565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156100ea57600080fd5b81516001600160a01b038116811461010157600080fd5b9392505050565b611194806101176000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c80638da5cb5b116100665780638da5cb5b146100e3578063a85c38ef14610108578063a98ad46c1461012f578063f2fde38b14610142578063fd53e1f21461015557600080fd5b806310199bdc146100985780636f9fb98a146100ad578063715018a6146100c85780638ba4cc3c146100d0575b600080fd5b6100ab6100a6366004610d0a565b610168565b005b6100b56105b0565b6040519081526020015b60405180910390f35b6100ab610622565b6100ab6100de366004610d80565b610636565b6000546001600160a01b03165b6040516001600160a01b0390911681526020016100bf565b61011b610116366004610daa565b6107f3565b6040516100bf989796959493929190610e09565b6002546100f0906001600160a01b031681565b6100ab610150366004610e61565b6108cd565b6100ab610163366004610e83565b61090b565b610170610b8f565b610178610bbc565b600084815260036020526040812080549091036101d35760405162461bcd60e51b815260206004820152601460248201527313dc99195c88191bd95cc81b9bdd08195e1a5cdd60621b60448201526064015b60405180910390fd5b6006810154600160a01b900460ff16156102265760405162461bcd60e51b815260206004820152601460248201527313dc99195c88185b1c9958591e4818db1bdcd95960621b60448201526064016101ca565b6000841180156102365750600083115b6102825760405162461bcd60e51b815260206004820152601d60248201527f507269636573206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60028101849055600381018390556001810161029e8382610f5c565b5060068101805460ff60a01b1916600160a01b179055600080858511156103aa57600583015486906102d08288611031565b6102da919061104a565b6102e49190611061565b915081905060006102f6600283611061565b9050600081856005015461030a9190611083565b600254600687015460405163a9059cbb60e01b81526001600160a01b03918216600482015260248101849052929350169063a9059cbb906044016020604051808303816000875af1158015610363573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103879190611096565b6103a35760405162461bcd60e51b81526004016101ca906110b8565b505061053c565b8585101561049b57600583015486906103c38783611031565b6103cd919061104a565b6103d79190611061565b90506103e2816110ef565b915060008184600501546103f69190611031565b9050801561049557600254600685015460405163a9059cbb60e01b81526001600160a01b0391821660048201526024810184905291169063a9059cbb906044016020604051808303816000875af1158015610455573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104799190611096565b6104955760405162461bcd60e51b81526004016101ca906110b8565b5061053c565b6002546006840154600585015460405163a9059cbb60e01b81526001600160a01b03928316600482015260248101919091526000945091169063a9059cbb906044016020604051808303816000875af11580156104fc573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105209190611096565b61053c5760405162461bcd60e51b81526004016101ca906110b8565b6004830182905560068301546040805189815260208101859052808201899052606081018890526001600160a01b039092166080830152517f06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d969181900360a00190a15050506105aa60018055565b50505050565b6002546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a0823190602401602060405180830381865afa1580156105f9573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061061d919061110b565b905090565b61062a610b8f565b6106346000610c15565b565b61063e610b8f565b610646610bbc565b6001600160a01b0382166106915760405162461bcd60e51b8152602060048201526012602482015271496e76616c696420746f206164647265737360701b60448201526064016101ca565b600081116106e15760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60025460405163a9059cbb60e01b81526001600160a01b038481166004830152602482018490529091169063a9059cbb906044016020604051808303816000875af1158015610734573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107589190611096565b6107a45760405162461bcd60e51b815260206004820152601c60248201527f555344542061697264726f70207472616e73666572206661696c65640000000060448201526064016101ca565b604080516001600160a01b0384168152602081018390527f8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a910160405180910390a16107ef60018055565b5050565b6003602052600090815260409020805460018201805491929161081590610ed4565b80601f016020809104026020016040519081016040528092919081815260200182805461084190610ed4565b801561088e5780601f106108635761010080835404028352916020019161088e565b820191906000526020600020905b81548152906001019060200180831161087157829003601f168201915b50505060028401546003850154600486015460058701546006909701549596929591945092506001600160a01b0381169060ff600160a01b9091041688565b6108d5610b8f565b6001600160a01b0381166108ff57604051631e4fbdf760e01b8152600060048201526024016101ca565b61090881610c15565b50565b610913610bbc565b600081116109635760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b600083815260036020526040902054156109bf5760405162461bcd60e51b815260206004820152601760248201527f4f7264657220494420616c72656164792065786973747300000000000000000060448201526064016101ca565b6002546040516323b872dd60e01b8152336004820152306024820152604481018390526001600160a01b03909116906323b872dd906064016020604051808303816000875af1158015610a16573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a3a9190611096565b610a7d5760405162461bcd60e51b81526020600482015260146024820152731554d115081d1c985b9cd9995c8819985a5b195960621b60448201526064016101ca565b604080516101008101825284815260208082018581526000838501819052606084018190526080840181905260a084018690523360c085015260e0840181905287815260039092529290208151815591519091906001820190610ae09082610f5c565b506040828101516002830155606083015160038301556080830151600483015560a0830151600583015560c08301516006909201805460e0909401511515600160a01b026001600160a81b03199094166001600160a01b0390931692909217929092179055517fee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b590610b79908590859085903390611124565b60405180910390a1610b8a60018055565b505050565b6000546001600160a01b031633146106345760405163118cdaa760e01b81523360048201526024016101ca565b600260015403610c0e5760405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c0060448201526064016101ca565b6002600155565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b634e487b7160e01b600052604160045260246000fd5b600082601f830112610c8c57600080fd5b813567ffffffffffffffff811115610ca657610ca6610c65565b604051601f8201601f19908116603f0116810167ffffffffffffffff81118282101715610cd557610cd5610c65565b604052818152838201602001851015610ced57600080fd5b816020850160208301376000918101602001919091529392505050565b60008060008060808587031215610d2057600080fd5b843593506020850135925060408501359150606085013567ffffffffffffffff811115610d4c57600080fd5b610d5887828801610c7b565b91505092959194509250565b80356001600160a01b0381168114610d7b57600080fd5b919050565b60008060408385031215610d9357600080fd5b610d9c83610d64565b946020939093013593505050565b600060208284031215610dbc57600080fd5b5035919050565b6000815180845260005b81811015610de957602081850181015186830182015201610dcd565b506000602082860101526020601f19601f83011685010191505092915050565b88815261010060208201526000610e2461010083018a610dc3565b6040830198909852506060810195909552608085019390935260a08401919091526001600160a01b031660c0830152151560e09091015292915050565b600060208284031215610e7357600080fd5b610e7c82610d64565b9392505050565b600080600060608486031215610e9857600080fd5b83359250602084013567ffffffffffffffff811115610eb657600080fd5b610ec286828701610c7b565b93969395505050506040919091013590565b600181811c90821680610ee857607f821691505b602082108103610f0857634e487b7160e01b600052602260045260246000fd5b50919050565b601f821115610b8a57806000526020600020601f840160051c81016020851015610f355750805b601f840160051c820191505b81811015610f555760008155600101610f41565b5050505050565b815167ffffffffffffffff811115610f7657610f76610c65565b610f8a81610f848454610ed4565b84610f0e565b6020601f821160018114610fbe5760008315610fa65750848201515b600019600385901b1c1916600184901b178455610f55565b600084815260208120601f198516915b82811015610fee5787850151825560209485019460019092019101610fce565b508482101561100c5786840151600019600387901b60f8161c191681555b50505050600190811b01905550565b634e487b7160e01b600052601160045260246000fd5b818103818111156110445761104461101b565b92915050565b80820281158282048414176110445761104461101b565b60008261107e57634e487b7160e01b600052601260045260246000fd5b500490565b808201808211156110445761104461101b565b6000602082840312156110a857600080fd5b81518015158114610e7c57600080fd5b6020808252601c908201527f55534454207472616e7366657220746f2075736572206661696c656400000000604082015260600190565b6000600160ff1b82016111045761110461101b565b5060000390565b60006020828403121561111d57600080fd5b5051919050565b84815260806020820152600061113d6080830186610dc3565b6040830194909452506001600160a01b03919091166060909101529291505056fea2646970667358221220622013ae607897e52b6504cb63250462abcd2849be6b3bf3a2616da2a79771a164736f6c634300081d003300000000000000000000000036e6504c968f5c2a310b6af7b97bc22cdd3402cc"
	input, err := hexutil.Decode(BztProduceBin)
	if err != nil {
		log.Fatalf("hexutil.Decode(input) err: %s", err)
		return nil, errors.New("hexutil.Decode(input) err")
	}
	txh, nonce, err := bzt.DeployContractTransfer(input, api.Client)
	if err != nil {
		log.Fatalf("DeployContractTransfer err: %s", err)
		return nil, errors.New("DeployContractTransfer err")
	}
	log.Info("DeployContract:", nonce)
	log.Info("DeployContract: ", txh)

	var tx mongo.DeployTransaction
	tx.TxHash = strings.ToLower(strings.ToLower(txh.Hash().String()))
	tx.Nonce = nonce
	tx.Gas = txh.Gas()
	tx.GasPrice = txh.GasPrice().String()
	err = mongo.AddDeployTransaction(tx)
	if err != nil {
		log.Fatalf("mongo.AddDeployTransaction err: %s", err)
		return nil, errors.New("mongo.AddDeployTransaction err")
	}
	return &v1.DeployContractReply{TxHash: tx.TxHash}, nil
}

func (s *GreeterService) GetBztDetails(ctx context.Context, in *v1.GetBztDetailsRequest) (*v1.GetBztDetailsReply, error) {
	urlOwnerAddr, err := bzt.UrlGetKeyAddress()
	if err != nil {
		log.Error("GetBztOwnerAddress err: ", err)
		return nil, err
	}
	log.Info("UrlGetKeyAddress", urlOwnerAddr)
	ba, err := bzt.GetTokenBalance()
	if err != nil {
		log.Error("GetTokenBalance err: ", err)
		return nil, err
	}
	dusdt, err := bzt.GetUsdToken()
	if err != nil {
		log.Error("GetUsdToken err: ", err)
		return nil, err
	}
	bztOwner, err := bzt.GetOwner()
	if err != nil {
		log.Error("GetOwner err: ", err)
		return nil, err
	}
	return &v1.GetBztDetailsReply{
		BztAddr:    bztOwner,
		Balance:    ba.String(),
		DusdtToken: dusdt,
		UrlAddr:    urlOwnerAddr,
	}, nil
}

func (s *GreeterService) GetBztVersion(ctx context.Context, _ *v1.GetBztVersionRequest) (*v1.GetBztVersionReply, error) {
	return &v1.GetBztVersionReply{
		Version:   "v0.0.26",
		BuildTime: "2025-09-08T13:50:00Z",
	}, nil
}

func (s *GreeterService) GetConfigs(ctx context.Context, in *v1.GetConfigsRequest) (*v1.GetConfigsReply, error) {

	return &v1.GetConfigsReply{
		ChainId:              api.ChainId,
		BztContractAddress:   conf.ContractBztAddr,
		DusdtContractAddress: conf.ContractDusdtAddress,
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
	log.Errorf("IsWalletBound mongo.GetUser err: %s", err)
	return false, err
}

func GetAddrAndUidByToken(ctx context.Context) (string, string, error) {
	cla, ok := kratosjwt.FromContext(ctx)
	if !ok {
		log.Error("err: jwt.FromContext")
		return "", "", errors.New("jwt.FromContext err")
	}

	claims, ok := cla.(jwtv5.MapClaims)
	if !ok {
		log.Error("err: invalid claims type")
		return "", "", errors.New("invalid claims type")
	}

	address, _ := claims["addr"].(string)
	if address == "" {
		return "", "", errors.New("invalid address")
	}

	uid, _ := claims["sub"].(string)
	if uid == "" {
		return "", "", errors.New("invalid uid")
	}
	return address, uid, nil
}

/*
	余额，等于未关仓加上用户亏损减去盈利部分
	奖励，等于用户盈利一半加上用户亏损
*/
