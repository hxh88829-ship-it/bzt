package service

import (
	"context"
	"strconv"
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

// Register TODO
/*
  // ...注册逻辑，写入 MongoDB

    // 注册成功后立即写入 Redis
    if err := RedisCli.SAdd(ctx, redisKey, strings.ToLower(user.Address)).Err(); err != nil {
        log.Warnf("⚠️ 注册用户后 Redis SAdd 失败: %v", err)
    }
*/

func (s *GreeterService) MarketCondition(ctx context.Context, in *v1.MarketConditionRequest) (*v1.MarketConditionReply, error) {
	return &v1.MarketConditionReply{}, nil
}
