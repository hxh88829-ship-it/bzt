package monitorBlock

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-kratos/kratos/v2/log"
	"strings"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/bzt"
	"valueguard/internal/mongo"
	"valueguard/internal/redisQuery"
)

var BztAddr = "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a"
var Platform = "0x331E865F47fd1b197d04Fe60E45DEf0C3A1EBA24"

func ScanBlocks() error {
	chainId := api.ChainId

	// 1. 从Mongo获取上次扫描的区块号
	mongoBln, err := mongo.GetScanBlock(chainId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var bl mongo.ScanBlock
			bl.NetWork = chainId
			bl.Time = time.Now().Unix()
			bl.LatestBlock = 0
			err = mongo.AddScanBlock(bl)
			if err != nil {
				return err
			}
			mongoBln = bl
		} else {
			return err
		}
	}
	//链上新块
	NewBlockNumber, err := api.GetBlockNumber()
	if err != nil {
		log.Error("ScanBlocks: api.GetBlockNumber error:", err)
		return err
	}
	//安全块
	SafeBlock := NewBlockNumber - 10

	if mongoBln.LatestBlock >= SafeBlock {
		log.Infof("无可处理新块，高度为 %d（当前区块 %d）", mongoBln.LatestBlock, NewBlockNumber)
		time.Sleep(time.Second / 100)
		return nil
	}

	const maxRetry = 3

	for blockNum := mongoBln.LatestBlock + 1; blockNum <= SafeBlock; blockNum++ {
		var bl *types.Block
		var err error

		// ----------------------
		// Retry GetBlockByNumber
		// ----------------------
		for i := 1; i <= maxRetry; i++ {
			bl, err = api.GetBlockByNumber(blockNum)
			if err == nil {
				break
			}
			log.Warnf("获取区块 %d 第 %d 次失败: %v", blockNum, i, err)
			time.Sleep(time.Second * time.Duration(i))
		}
		if err != nil {
			log.Errorf("获取区块 %d 最终失败，退出处理: %v", blockNum, err)
			return err
		}

		// ----------------------
		// Retry ProcessTransactions
		// ----------------------
		for i := 1; i <= maxRetry; i++ {
			err = ProcessTransactions(bl)
			if err == nil {
				break
			}
			log.Warnf("处理区块 %d 交易 第 %d 次失败: %v", blockNum, i, err)
			time.Sleep(time.Second * time.Duration(i))
		}
		if err != nil {
			log.Errorf("处理区块 %d 交易失败，退出处理: %v", blockNum, err)
			return err
		}

		// ----------------------
		// Retry AddScanBlockPlace
		// ----------------------
		for i := 1; i <= maxRetry; i++ {
			err = UpdateScanBlockPlace(blockNum)
			if err == nil {
				break
			}
			log.Warnf("更新区块 %d 扫描位置 第 %d 次失败: %v", blockNum, i, err)
			time.Sleep(time.Second * time.Duration(i))
		}
		if err != nil {
			log.Errorf("更新区块 %d 扫描位置失败，退出处理: %v", blockNum, err)
			return err
		}
	}

	return nil
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
			_, err = ParseEvents(tx, receipt, blockTime, from)
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

func ParseEvents(tx *types.Transaction, receipt *types.Receipt, blTime uint64, from common.Address) (string, error) {
	OrderClosedTopic := common.HexToHash("0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96")
	AirdropTopic := common.HexToHash("0x8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a")
	for _, vLog := range receipt.Logs {
		switch vLog.Topics[0] {
		case OrderClosedTopic:
			// 解析 关仓 事件
			order, err := bzt.GetParseOrderClosed(receipt)
			if err != nil {
				return "", fmt.Errorf("解析 OrderClosed 失败: %w", err)
			}
			if order != nil {
				// 关仓事件数据记录
				err = OrderClosedTrade(order, blTime)
				if err != nil {
					return "", fmt.Errorf("<UNK> OrderClosed <UNK>: %w", err)
				}
				log.Infof("🔒 识别为关仓事件: TxHash=%s", receipt.TxHash.Hex())
				err = AddTransactionTrade(tx, receipt, from, blTime, "OrderClosed")
				if err != nil {
					return "", fmt.Errorf("<UNK> OrderClosed <UNK>: %w", err)
				}
				return "order_closed", nil
			}
		case AirdropTopic:
			// 解析 空投 事件
			airdrop, err := bzt.GetParseAirdrop(receipt)
			if err != nil {
				return "", fmt.Errorf("解析 Airdrop 失败: %w", err)
			}
			if airdrop != nil {
				//空投事件数据记录
				err = AirdropTrade(airdrop, receipt, blTime)
				if err != nil {
					log.Errorf("Airdrop : %v", err)
					return "", fmt.Errorf("<UNK> Airdrop <UNK>: %w", err)
				}
				log.Infof("🎁 识别为空投事件: TxHash=%s", receipt.TxHash.Hex())
				err = AddTransactionTrade(tx, receipt, from, blTime, "Airdrop")
				if err != nil {
					return "", fmt.Errorf("<UNK> Airdrop <UNK>: %w", err)
				}
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
	if receipt.Status == 0 {
		log.Warnf(" TxHash=%s", receipt.TxHash.String())
		return nil
	}
	event, err := bzt.GetParseOrderOpened(receipt)
	if err != nil {
		return err
	}
	_, err = mongo.GetOrder(event.OrderId.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			price, err := mongo.GetPriceByTimestamp(blTime, event.TokenName)
			if err != nil {
				log.Errorf("GetPriceByTimestamp err: %v", err)
				return err
			}
			var userOrder mongo.Order
			userOrder.OrderId = event.OrderId.String()
			userOrder.Symbol = event.TokenName
			userOrder.OpenPrice = price.Price
			userOrder.OrderStartTime = blTime
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
func OrderClosedTrade(event *bzt.BztOrderClosed, blTime uint64) error {
	_, err := mongo.GetOrder(event.OrderId.String())
	if err != nil {
		log.Errorf("GetOrdererr: %v\n %v", event.OrderId.String(), err)
		return err
	}
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

func AirdropTrade(event *bzt.BztAirdrop, receipt *types.Receipt, blTime uint64) error {
	_, err := mongo.GetAirdrop(event.Raw.TxHash.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var air mongo.Airdrop
			air.TxHash = strings.ToLower(event.Raw.TxHash.String())
			air.Symbol = "USDT"
			air.ToAddr = strings.ToLower(event.Recipient.String())
			air.Amount = event.Amount.String()
			air.AirdropTime = blTime
			air.Status = receipt.Status
			err = mongo.AddAirdrop(air)
			if err != nil {
				log.Errorf("AddAirdrop err: %v", err)
				return err
			}
		} else {
			log.Errorf("GetAirdroperr: %v", err)
			return err
		}
	}
	return nil
}

func AddTransactionTrade(txh *types.Transaction, receipt *types.Receipt,
	from common.Address, blTime uint64, name string) error {
	_, err := mongo.GetTransaction(receipt.TxHash.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var tx mongo.Transaction
			tx.TxHash = strings.ToLower(receipt.TxHash.String())
			tx.From = strings.ToLower(from.String())
			tx.To = strings.ToLower(txh.To().String())
			tx.Value = txh.Value().String()
			tx.Data = hexutil.Encode(txh.Data())
			tx.Nonce = txh.Nonce()
			tx.Gas = txh.Gas()
			tx.GasPrice = txh.GasPrice().String()
			tx.Number = receipt.BlockNumber.Uint64()
			tx.Status = receipt.Status
			tx.Time = blTime
			tx.TransactionType = name
		} else {
			log.Errorf("GetTransactionerr: %v", err)
			return err
		}
	}
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
