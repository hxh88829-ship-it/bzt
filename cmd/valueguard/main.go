package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"sync"
	"time"
	"valueguard/internal/api"
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

	log.Info(bc)
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
	//初始化mongo数据库
	cli, err := mongo.NewMongoClient("mongodb://admin:admin@13.212.58.194:9097")
	if err != nil {
		return err
	}
	mongo.MonCli = cli

	//配置变量
	/*
		os.Setenv("Apikey", "dtcd_xxxxxx")
		os.Setenv("BaseUrl", "http://47.111.28.25:8016")
		os.Setenv("KeyId", "0a1382ae-7e21-49e8-928e-0614103b2045")
		os.Setenv("OwnerAddress", "0x5D001706b0b4bF6a0D5C234E1F966D82D3C84F92")
		os.Setenv("RpcUrl", "https://f82o1hrgdl.execute-api.ap-southeast-1.amazonaws.com/prod")
		//os.Setenv("RpcUrl", "http://ec2-54-251-227-86.ap-southeast-1.compute.amazonaws.com:6979")
		os.Setenv("HmacKey", "hmac")
		os.Setenv("X_Api_Key", "4sip97qapC4vTxS73YdTB6X5hm8Rr8Uk13BdwP2d")
		os.Setenv("ContractDusdtAddress", "0xaD6780B2A022B79686c5E56017cC4FB8cfCd9726") //测试环境DUSDT
		os.Setenv("ContractBztAddr", "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a")
	*/

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

	conf.ContractBztAddr = "0x747294d3e04c1ad8b4897bc6fdab72bfa9b5c3f4"

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

	//初始化节点
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

	//部署合约，拿到Owner
	owner, err := bzt.UrlGetKeyAddress()
	if err != nil {
		log.Error("UrlGetKeyAddress", "err", err)
		return err
	}
	log.Info("UrlGetKeyAddress:", owner)

	if id.Uint64() == 9798 {
		symbols := []string{"BTCUSDT", "ETHUSDT"}
		go RunService(context.Background(), symbols)

		// ✅ 启动空投定时任务
		go dailyAirdrop.StartAirdropCron(symbols)
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
