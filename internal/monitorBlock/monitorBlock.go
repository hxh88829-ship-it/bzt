package monitorBlock

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-kratos/kratos/v2/log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/bzt"
	"valueguard/internal/mongo"
	"valueguard/internal/redisQuery"
)

var BztAddr = "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a"
var Platform = "0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24"

// ScanBlocks 是主扫块逻辑，带 Redis 锁控制并发安全、自动重试与失败块记录
func ScanBlocks(ctx context.Context, maxConcurrency int) error {
	lockKey := fmt.Sprintf("lock:block_scan:%d", api.ChainId)
	lockTTL := 30 * time.Second
	retryDelay := 300 * time.Millisecond

	return redisQuery.WithRedisLock(ctx, lockKey, lockTTL, 3, retryDelay, func() error {
		// 获取已扫记录与当前链上高度
		mongoBln, err := mongo.GetScanBlock(api.ChainId)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				mongoBln = mongo.ScanBlock{
					NetWork:     api.ChainId,
					Time:        time.Now().Unix(),
					LatestBlock: 0,
				}
				if err := mongo.AddScanBlock(mongoBln); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		latestBlock, err := api.GetBlockNumber()
		if err != nil {
			return err
		}
		log.Info(latestBlock)
		safeBlock := latestBlock - 10
		if mongoBln.LatestBlock >= safeBlock {
			log.Infof("✅ 无新块可扫，当前高度 %d", mongoBln.LatestBlock)
			return nil
		}

		blockCh := make(chan uint64, maxConcurrency)
		var wg sync.WaitGroup
		var firstErr atomic.Value         // 存储首个错误
		var maxSuccessBlock atomic.Uint64 // 当前处理成功的最大块

		// 启动多个并发 worker 处理区块
		for i := 0; i < maxConcurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer recoverGoroutine()

				for blockNum := range blockCh {
					// 🧩 调用封装后的块处理函数（包含重试、错误记录、成功记录等）
					handleBlockWithRetry(blockNum, &firstErr, &maxSuccessBlock)
				}
			}()
		}

		// 投递区块任务到 blockCh
		go func() {

			for blockNum := mongoBln.LatestBlock + 1; blockNum <= safeBlock; blockNum++ {
				blockCh <- blockNum
			}
			close(blockCh)
		}()

		wg.Wait()

		successBlock := maxSuccessBlock.Load()
		if successBlock == 0 {
			log.Warn("⚠️ 没有任何区块成功处理")
			return nil
		}

		// 更新 MongoDB 中的已扫区块高度
		if err := UpdateScanBlockPlace(successBlock); err != nil {
			return err
		}

		// 若有错误直接返回（只返回首个错误）
		if errVal := firstErr.Load(); errVal != nil {
			return errVal.(error)
		}

		log.Infof("✅ 成功扫块至 %d", successBlock)
		return nil
	})
}

// handleBlockWithRetry 负责获取区块 + 处理交易 + 错误记录 + 成功块更新
func handleBlockWithRetry(blockNum uint64, firstErr *atomic.Value, maxSuccessBlock *atomic.Uint64) {
	var bl *types.Block
	var err error

	// 尝试获取区块（最多重试 3 次）
	for attempt := 1; attempt <= 3; attempt++ {
		bl, err = api.GetBlockByNumber(blockNum)
		if err == nil {
			break
		}
		log.Warnf("⛏️ 获取区块 %d 第 %d 次失败: %v", blockNum, attempt, err)
		time.Sleep(time.Second * time.Duration(attempt))
	}
	if err != nil {
		recordFirstError(firstErr, fmt.Errorf("获取区块 %d 失败: %w", blockNum, err))
		// 🧩 原先遗漏了记录获取失败的块
		if err := AddLossBlock(blockNum); err != nil {
			log.Warnf("AddLossBlock %d 错误（GetBlock失败）: %v", blockNum, err)
		}
		return
	}

	// 尝试处理区块交易（最多重试 3 次）
	for attempt := 1; attempt <= 3; attempt++ {
		err = ProcessTransactions(bl)
		if err == nil {
			break
		}
		log.Warnf("📦 处理区块 %d 第 %d 次失败: %v", blockNum, attempt, err)
		time.Sleep(time.Second * time.Duration(attempt))
	}
	if err != nil {
		recordFirstError(firstErr, fmt.Errorf("处理区块 %d 交易失败: %w", blockNum, err))

		// 持久化失败块用于后续补偿
		if err := AddLossBlock(blockNum); err != nil {
			log.Warnf("AddLossBlock %d 错误: %v", blockNum, err)
		}
		return
	}

	// ✅ 区块成功处理，尝试记录最大成功高度
	updateSuccessBlock(blockNum, maxSuccessBlock)
}

// recoverGoroutine 防止 goroutine 中 panic 崩溃主流程
func recoverGoroutine() {
	if r := recover(); r != nil {
		log.Errorf("🔥 goroutine panic: %v", r)
	}
}

// recordFirstError 仅记录第一次出现的错误，线程安全
func recordFirstError(firstErr *atomic.Value, err error) {
	firstErr.CompareAndSwap(nil, err)
}

// updateSuccessBlock 安全更新最大已成功处理区块号
func updateSuccessBlock(blockNum uint64, maxBlock *atomic.Uint64) {
	maxBlock.CompareAndSwap(0, blockNum)
	for {
		old := maxBlock.Load()
		if blockNum > old {
			if maxBlock.CompareAndSwap(old, blockNum) {
				break
			}
		} else {
			break
		}
	}
}

func ProcessTransactions(bl *types.Block) error {
	blockTime := bl.Time()
	ctx := context.Background()
	for _, tx := range bl.Transactions() {
		from, err := api.GetFromByTransaction(tx)
		if err != nil {
			return err
		}
		if tx.To() == nil { //部署合约跳过
			continue
		}
		if strings.ToLower(tx.To().String()) == strings.ToLower(BztAddr) {
			//开仓,同时核查from是否为平台用户，防止数据污染
			ok, err := redisQuery.IsPlatformUser(ctx, strings.ToLower(from.String()))
			if err != nil {
				log.Errorf("redisQuery.IsPlatformUser err: %v", err)
				return err
			}
			if !ok {
				continue
			}
			//拿到开仓数据
			err = OrderOpenedTrade(tx, blockTime)
			if err != nil {
				return err
			}
		} else if strings.ToLower(from.String()) == strings.ToLower(BztAddr) {
			//空投（用户触发），关仓（平台），需要解析事件区分
			receipt, err := api.GetTransactionReceiptByHash(tx.Hash())
			if err != nil {
				log.Errorf("GetTransactionReceiptByHash err: %v", err)
				return err
			}
			_, err = ParseEvents(receipt, blockTime)
			if err != nil {
				log.Errorf("ParseEvents err: %v", err)
				return err
			}
		} else if strings.ToLower(from.String()) == strings.ToLower(Platform) {
			//充值
		} else if strings.ToLower(tx.To().String()) == strings.ToLower(Platform) {
			//提现
		} else {
			continue //什么都不是
		}
	}
	return nil
}

func ParseEvents(receipt *types.Receipt, blTime uint64) (string, error) {
	OrderClosedSigHash := common.HexToHash("0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96")
	AirdropSigHash := common.HexToHash("0x8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a")
	for _, vLog := range receipt.Logs {
		switch vLog.Topics[0] {
		case OrderClosedSigHash:
			// 解析 关仓 事件
			order, err := bzt.GetParseOrderClosed(receipt)
			if err != nil {
				return "", fmt.Errorf("解析 OrderClosed 失败: %w", err)
			}
			if order != nil {
				// 关仓事件数据记录
				err = OrderClosedTrade(order, int64(blTime))
				if err != nil {
					return "", fmt.Errorf("<UNK> OrderClosed <UNK>: %w", err)
				}
				log.Infof("🔒 识别为关仓事件: TxHash=%s", receipt.TxHash.Hex())
				return "order_closed", nil
			}
		case AirdropSigHash:
			// 解析 空投 事件
			airdrop, err := bzt.GetParseAirdrop(receipt)
			if err != nil {
				return "", fmt.Errorf("解析 Airdrop 失败: %w", err)
			}
			if airdrop != nil {
				//空投事件数据记录
				log.Infof("🎁 识别为空投事件: TxHash=%s", receipt.TxHash.Hex())
				return "airdrop", nil
			}
		}
	}
	log.Infof("📭 未识别事件类型: TxHash=%s", receipt.TxHash.Hex())
	return "unknown", nil
}
func OrderOpenedTrade(tx *types.Transaction, blTime uint64) error {
	receipt, err := api.GetTransactionReceiptByHash(tx.Hash())
	if err != nil {
		return err
	}
	event, err := bzt.GetParseOrderOpened(receipt)
	if err != nil {
		return err
	}
	_, err = mongo.GetOrder(event.OrderId.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			orderTime := int64(blTime)
			price, err := mongo.GetPriceByTimestamp(orderTime, event.TokenName)
			if err != nil {
				log.Errorf("GetPriceByTimestamp err: %v", err)
				return err
			}
			var userOrder mongo.Order
			userOrder.OrderId = event.OrderId.String()
			userOrder.Symbol = event.TokenName
			userOrder.OpenPrice = price.Price
			userOrder.OrderStartTime = orderTime
			userOrder.Amount = event.Amount.String()
			userOrder.UsersAddr = strings.ToLower(event.User.String())
			userOrder.IsClosed = false
		} else {
			log.Errorf("GetOrdererr: %v", err)
			return err
		}
	}
	return nil
}
func OrderClosedTrade(event *bzt.BztOrderClosed, blTime int64) error {
	//UserOrder, err := mongo.GetOrder(event.OrderId.String())
	//if err != nil {
	//	log.Errorf("GetOrdererr: %v", err)
	//	return err
	//}
	resOrder, err := bzt.GetOrders(event.OrderId.Int64())
	if err != nil {
		log.Errorf("GetOrderserr: %v", err)
		return err
	}
	err = mongo.UpdateOrder(event.OrderId.String(), resOrder.ClosePrice.String(), resOrder.ProfitLoss.String(), blTime)
	if err != nil {
		log.Errorf("UpdateOrdererr: %v", err)
		return err
	}

	return nil
}

func AirdropTrade() error {

	return nil
}

func AddTransactionTrade() error {
	return nil
}

func UpdateScanBlockPlace(safeBlock uint64) error {
	var ScanBlock mongo.ScanBlock
	ScanBlock.NetWork = api.ChainId
	ScanBlock.Time = time.Now().Unix()
	ScanBlock.LatestBlock = safeBlock
	err := mongo.UpdateScanBlock(ScanBlock)
	if err != nil {
		log.Error("MonitorBlock  AddScanBlockPlace UpdateScanBlock")
		return err
	}
	return nil
}

func AddLossBlock(safeBlock uint64) error {
	_, err := mongo.GetLossBlock(safeBlock)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var LossBlock mongo.LossBlock
			LossBlock.NetWork = api.ChainId
			LossBlock.Time = time.Now().Unix()
			LossBlock.BlockNr = safeBlock
			err = mongo.AddLossBlock(LossBlock)
			if err != nil {
				log.Error("MonitorBlock  AddLossBlock AddLossBlock")
				return err
			}
		} else {
			log.Error("MonitorBlock  AddLossBlock AddLossBlock")
			return err
		}
	}
	return nil
}
