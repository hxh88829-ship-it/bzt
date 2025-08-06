package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-kratos/kratos/v2/log"
	"regexp"
	"strconv"
	"strings"
	"time"
	v1 "valueguard/api/helloworld/v1"
	"valueguard/internal/api"
	"valueguard/internal/biz"
	"valueguard/internal/mongo"
	"valueguard/internal/redisQuery"
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

// BindWallet TODO
/*
  // ...注册逻辑，写入 MongoDB

    // 注册成功后立即写入 Redis
    if err := RedisCli.SAdd(ctx, redisKey, strings.ToLower(user.Address)).Err(); err != nil {
        log.Warnf("⚠️ 注册用户后 Redis SAdd 失败: %v", err)
    }
*/
func (s *GreeterService) BindWallet(ctx context.Context, in *v1.BindWalletRequest) (*v1.BindWalletReply, error) {
	isAddress := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`).MatchString
	addr := strings.ToLower(in.GetAddress())
	if !isAddress(addr) {
		return &v1.BindWalletReply{Metadata: "invalid address"}, nil
	}
	exists, err := IsWalletBound(ctx, addr)
	if err != nil {
		return &v1.BindWalletReply{}, err
	}
	if exists {
		return nil, errors.New("wallet already bound")
	}
	uid := api.GenerateUID()
	nonce := api.GenerateUID()
	message := fmt.Sprintf("保值通系统请求绑定你的地址：\n%s\n操作类型: bind_wallet\nNonce: %s\nIssued At: %s", addr, nonce, time.Now().Format(time.RFC3339))

	user := mongo.Users{
		Address:         addr,
		Uid:             uid,
		Email:           in.GetEmail(),
		Name:            in.GetName(),
		OriginalMessage: message,
		CreateTimeAt:    time.Now().Unix(),
		Status:          "0",
	}

	err = mongo.AddUser(user)
	if err != nil {
		return nil, err
	}
	if err := redisQuery.RedisCli.SAdd(ctx, "platform_users_set", strings.ToLower(user.Address)).Err(); err != nil {
		log.Warnf("⚠️ 注册用户后 Redis SAdd 失败: %v", err)
	}

	return &v1.BindWalletReply{
		Uid:      uid,
		Metadata: message,
		Hash:     hexutil.Encode(api.ComputeMessageHash(message)),
	}, nil
}

func (s *GreeterService) LoginWithWallet(ctx context.Context, in *v1.LoginRequest) (*v1.LoginReply, error) {
	us, err := mongo.GetUser(in.GetUid())
	if err != nil {
		return nil, err
	}

	addr, err := api.VerifyForAddress(us.OriginalMessage, in.GetSignature())
	if err != nil {
		return nil, errors.New("signature verification failed")
	}
	if addr != us.Address {
		log.Warnf(addr, "\n", us.Address)
		return nil, errors.New("signature not match with address")
	}

	// 清除已使用的签名
	_ = mongo.UpdateUser(us.Uid, "")

	// 生成 JWT

	jwtToken, err := api.GetJwtKey(us.Uid, strings.ToLower(us.Address))
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		Token: jwtToken,
	}, nil
}

func (s *GreeterService) MarketCondition(ctx context.Context, in *v1.MarketConditionRequest) (*v1.MarketConditionReply, error) {
	priTime := time.Now().Unix()
	res, err := mongo.GetPriceByTimestamp(uint64(priTime), in.GetSymbol())
	if err != nil {
		log.Warnf("symbol %v---err:%v", in.GetSymbol(), err)
		return nil, errors.New("symbol not exist")
	}
	return &v1.MarketConditionReply{
		Price: res.Price,
		Time:  res.Timestamp,
	}, nil
}

func IsWalletBound(ctx context.Context, addr string) (bool, error) {
	addr = strings.ToLower(addr)
	// Redis 判断
	exists, err := redisQuery.RedisCli.SIsMember(ctx, "platform_users_set", addr).Result()
	if err == nil && exists {
		return true, nil
	}
	// Mongo 判断
	_, err = mongo.GetUser(addr)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, err
}
