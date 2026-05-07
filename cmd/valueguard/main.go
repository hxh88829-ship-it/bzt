package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"sync"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/binanceClient"
	"valueguard/internal/bzt"
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
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
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
func LoadConfigInit() error {
	//配置变量
	//初始化mongo数据库
	//初始化节点
	mongoDbUrl := os.Getenv("mongoDbUrl")
	if mongoDbUrl == "" {
		return errors.New("mongoDbUrl is empty")
	}
	cli, err := mongo.NewMongoClient(mongoDbUrl)
	if err != nil {
		return err
	}
	mongo.MonCli = cli

	BinanceApikey := os.Getenv("BINANCE_API_KEY")
	if BinanceApikey == "" {
		return errors.New("BINANCE_API_KEY is empty")
	}
	conf.BinanceApikey = BinanceApikey

	BinanceSecretKey := os.Getenv("BINANCE_SECRET_KEY")
	if BinanceSecretKey == "" {
		return errors.New("BINANCE_SECRET_KEY is empty")
	}
	conf.BinanceSecretKey = BinanceSecretKey

	// TODO true 测试网
	binanceClient.InitBinanceClient(conf.BinanceApikey, conf.BinanceSecretKey, false)

	headerKey := os.Getenv("Apikey")
	if headerKey == "" {
		return errors.New("apikey is required")
	}
	conf.Apikey = headerKey

	BaseUrl := os.Getenv("BaseUrl")
	if BaseUrl == "" {
		return errors.New("BaseUrl is required")
	}
	conf.BaseUrl = BaseUrl

	KeyId := os.Getenv("KeyId")
	if KeyId == "" {
		return errors.New("KeyId is required")
	}
	conf.KeyId = KeyId

	OwnerAddress := os.Getenv("OwnerAddress")
	if OwnerAddress == "" {
		return errors.New("OwnerAddress is required")
	}
	conf.OwnerAddress = OwnerAddress
	log.Info("OwnerAddress:", conf.OwnerAddress)
	HmacKey := os.Getenv("HmacKey")
	if HmacKey == "" {
		return errors.New("HmacKey is required")
	}
	conf.HmacKey = HmacKey

	conf.ContractBztAddr = "0x747294d3e04c1ad8b4897bc6fdab72bfa9b5c3f4" //生产环境
	//TODO 测试需要修改
	//conf.ContractBztAddr = "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a" //测试环境
	ContractDusdtAddress := os.Getenv("ContractDusdtAddress")
	if ContractDusdtAddress == "" {
		return errors.New("ContractDusdtAddress is required")
	}
	conf.ContractDusdtAddress = ContractDusdtAddress

	X_Api_Key := os.Getenv("X_Api_Key")
	if X_Api_Key == "" {
		return errors.New("X_Api_Key is required")
	}
	conf.X_Api_Key = X_Api_Key

	conf.Secret = "d7b2d30b1945d542ceee7a85e31d2edd07b37be1b87d1dc091feb40b432ece2d"
	rpcUrl := os.Getenv("RpcUrl")
	if rpcUrl == "" {
		return errors.New("rpc url is empty")
	}
	conf.RpcUrl = rpcUrl
	log.Info("rpc url is ", rpcUrl)
	api.Client, err = bzt.InitEthClient(rpcUrl)
	if err != nil {
		return errors.New("rpc url is invalid")
	}

	id, err := api.Client.ChainID(context.Background())
	if err != nil {
		log.Error("Client.ChainID", "err", err)
		return err
	}
	api.ChainId = id.Uint64()
	log.Info("chain id is:  ", id)

	api.InitSnowflake(1)
	//TODO 测试需要修改链归属
	if id.Uint64() == 9798 {
		symbols := []string{"BTCUSDT", "ETHUSDT"}
		go RunService(context.Background(), symbols)

		// ✅ 启动空投定时任务
		go dailyAirdrop.StartAirdropCron()
	}
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
				log.Errorf("panic in block scanner: %v", r)
				// panic 后自动退出，下一次程序重启可再次启动
			}
		}()

		for {
			select {
			case <-ctx.Done():
				log.Info("Block scanner stopping...")
				return
			default:
				// 扫块
				number, err := monitorBlock.ScanBlocks(ctx)
				if err != nil {
					if err.Error() == "block is latest" {
						time.Sleep(3 * time.Second) // 防止快速循环刷日志
						continue
					}

					log.Errorf("ScanBlocks returned error: %v", err)
					// 出错立即停止本轮，下次循环会重新从错误块开始
					time.Sleep(3 * time.Second) // 防止快速循环刷日志
					continue
				} else {
					//更新块高 num =>  mongo
					// ✅ 只有处理成功才更新数据库
					err = monitorBlock.UpdateScanBlockPlace(number)
					if err != nil {
						log.Errorf("UpdateScanBlockPlace failed at block %d: %v", number, err)
						time.Sleep(3 * time.Second)
						continue
					}
				}
			}
		}
	}() //扫块

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
					symbolIndexes[symbol] = (index + 1) % 3600
				}
			}
		}
	}() //获取实时行情

	<-ctx.Done()
	wg.Wait()
	log.Info("Service fully stopped")
}
