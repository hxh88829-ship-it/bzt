package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
	"sync"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/conf"
	"valueguard/internal/dailyAirdrop"
	"valueguard/internal/marketCondition"
	"valueguard/internal/mongo"
	"valueguard/internal/monitorBlock"

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

	log.Info(bc)
	if err := LoadConfigInit(&bc); err != nil {
		panic(err)
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
func LoadConfigInit(bc *conf.Bootstrap) error {
	//初始化mongo数据库
	cli, err := mongo.NewMongoClient(bc.Chain.GetMongoUrl())
	if err != nil {
		return err
	}
	mongo.MonCli = cli

	//初始化节点
	os.Setenv("RPC_Url", "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
	// 通过环境变量获取节点rpc url
	rurl := os.Getenv("RPC_Url")
	if rurl == "" {
		return errors.New("rpc url is empty")
	}
	log.Info("rpc url is ", rurl)
	api.Client, err = ethclient.Dial(rurl)
	if err != nil {
		return errors.New("BLockChain fail")
	}

	id, err := api.Client.ChainID(context.Background())
	if err != nil {
		log.Error("Client.ChainID", "err", err)
		return err
	}
	api.ChainId = id.Uint64()

	//初始化签名机器
	//key := os.Getenv("Key_id")
	//if key == "" {
	//	return errors.New("key is empty")
	//}
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	go RunService(context.Background(), symbols)

	// ✅ 启动空投定时任务
	go dailyAirdrop.StartAirdropCron(symbols)
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
	}() //扫块

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
					//	log.Infof("Market price checker: %v", index)
					err := marketCondition.GetMarketCondition(symbol, index)
					if err != nil {
						log.Errorf("Failed to fetch %s: %v", symbol, err)
					}
					symbolIndexes[symbol] = (index + 1) % 30
				}
			}
		}
	}() //获取实时行情

	<-ctx.Done()
	wg.Wait()
	log.Info("Service fully stopped")
}
