package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"os"
	"sync"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/marketCondition"
	"valueguard/internal/mongo"
	"valueguard/internal/monitorBlock"
	"valueguard/internal/redisQuery"

	"valueguard/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "/Users/huangxin/work/smh/valueguard/configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	if err := LoadConfigInit(); err != nil {
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// 加载配置 load
func LoadConfigInit() error {
	//初始化mongo数据库
	cli, err := mongo.NewMongoClient("mongodb://admin:admin@localhost:27017/?directConnection=true")
	if err != nil {
		return err
	}
	mongo.MonCli = cli

	//初始化节点
	api.Client, err = ethclient.Dial("http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	if err != nil {
		return errors.New("BLockChain fail")
	}

	id, err := api.Client.ChainID(context.Background())
	if err != nil {
		log.Error("Client.ChainID", "err", err)
		return err
	}
	api.ChainId = id.Uint64()

	//初始化redis
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Redis 服务器地址
		Password: "",               // 密码，如果没有则为空字符串
		DB:       0,                // 使用默认数据库 (0)
	})
	redisQuery.RedisCli = redisCli

	symbols := []string{"BTCUSDT", "ETHUSDT"}
	go RunService(context.Background(), symbols)

	return nil
}

func RunService(ctx context.Context, symbols []string) {
	var symbolIndexes = make(map[string]uint64)
	var index uint64
	for _, symbol := range symbols {
		symbolIndexes[symbol] = 0
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in Block checker: %v", r)
			}
		}()
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("Block scanner stopping...")
				return
			case <-ticker.C:
				blockCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
				err := monitorBlock.ScanBlocks(blockCtx)
				cancel()

				if err != nil {
					log.Errorf("ScanBlocks error: %v", err)
				}
			}
		}
	}()

	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	ticker := time.NewTicker(3 * time.Minute)
	//	defer ticker.Stop()
	//
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			log.Info("LossBlock retry task stopping...")
	//			return
	//		case <-ticker.C:
	//			err := monitorBlock.RetryLossBlocks(ctx)
	//			if err != nil {
	//				log.Errorf("RetryLossBlocks error: %v", err)
	//			}
	//		}
	//	}
	//}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in market price checker: %v", r)
			}
		}()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("Market price checker stopping...")
				return
			case <-ticker.C:
				for _, symbol := range symbols {
					index = symbolIndexes[symbol]
					log.Infof("Market price checker: %v", index)
					err := marketCondition.GetMarketCondition(symbol, index)
					if err != nil {
						log.Errorf("Failed to fetch %s: %v", symbol, err)
					}
					symbolIndexes[symbol] = (index + 1) % 30
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in redis checker: %v", r)
			}
		}()
		// ✅ 第一次立即执行
		if err := redisQuery.SafeSyncPlatformUsersToRedis(context.Background()); err != nil {
			log.Errorf("首次同步 Redis 失败: %v", err)
		}
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("redis Add stopping...")
				return
			case <-ticker.C:
				if err := redisQuery.SafeSyncPlatformUsersToRedis(context.Background()); err != nil {
					log.Errorf("mpngodb Scan failed: %v", err)
				}
			}
		}
	}()

	<-ctx.Done()
	wg.Wait()
	log.Info("Service fully stopped")
}
