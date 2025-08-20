package monitorBlock

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-kratos/kratos/v2/log"
	"math"
	"math/big"
	"strings"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/bzt"
	"valueguard/internal/conf"
	"valueguard/internal/mongo"
)

var Plat = strings.ToLower("0xc020e62ce44297e86dA12CF15CfDc20B83eF3b72")

func ScanBlocks(ctx context.Context) error {
	mongoBln, safeBlock, err := GetMongodbBlockAndLinkBlock()
	if err != nil {
		return err
	}
	if mongoBln.LatestBlock >= safeBlock {
		log.Infof("无可处理新块，高度为 %d（当前区块 %d）", mongoBln.LatestBlock, safeBlock+10)
		return nil
	}

	for blockNum := mongoBln.LatestBlock + 1; blockNum <= safeBlock; blockNum++ {
		err := ScanOneBlock(ctx, blockNum)
		if err != nil {
			log.Warnf("扫描区块 %d 失败：%v", blockNum, err)
		}
	}
	return nil
}
func ScanOneBlock(ctx context.Context, blockNum uint64) error {
	const maxRetry = 3

	var bl *types.Block
	err := WithRetry(maxRetry, fmt.Sprintf("获取区块 %d", blockNum), func() error {
		var e error
		bl, e = api.GetBlockByNumber(blockNum)
		return e
	})
	if err != nil {
		if err2 := AddLossBlock(blockNum, "获取区块失败: "+err.Error()); err2 != nil {
			log.Errorf("写入失败区块失败: %v", err2)
		}
		return err
	}

	//start := time.Now()
	err = WithRetry(maxRetry, fmt.Sprintf("处理区块 %d 交易", blockNum), func() error {
		return ProcessTransactions(bl, ctx)
	})
	if err != nil {
		if err2 := AddLossBlock(blockNum, "处理区块失败: "+err.Error()); err2 != nil {
			log.Errorf("写入失败区块失败: %v", err2)
		}
		return err
	}
	//log.Infof("ProcessTransactions耗时: %s", time.Since(start))

	err = WithRetry(maxRetry, fmt.Sprintf("更新区块 %d 扫描位置", blockNum), func() error {
		return UpdateScanBlockPlace(blockNum)
	})
	if err != nil {
		if err2 := AddLossBlock(blockNum, "更新区块失败: "+err.Error()); err2 != nil {
			log.Errorf("写入失败区块失败: %v", err2)
		}
		return err
	}

	return nil
}
func RetryLossBlocks(ctx context.Context) error {
	lossBlocks, err := mongo.GetLossBlocksByNetwork(api.ChainId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil // 无失败块无需处理
		}
		return err
	}
	for _, lb := range lossBlocks {
		log.Infof("尝试重试失败区块 %d", lb.BlockNr)

		err := ScanOneBlock(ctx, lb.BlockNr)
		if err != nil {
			log.Warnf("失败区块 %d 重试失败: %v", lb.BlockNr, err)
			continue
		}

		// 扫描成功，删除失败块记录
		if err := mongo.DeleteLossBlock(lb.BlockNr); err != nil {
			log.Errorf("删除失败区块记录失败 block=%d err=%v", lb.BlockNr, err)
		} else {
			log.Infof("失败区块 %d 重试成功，删除失败记录", lb.BlockNr)
		}
	}
	return nil
}
func UpdateScanBlockPlace(blockNum uint64) error {
	var ScanBlock mongo.ScanBlock
	ScanBlock.NetWork = api.ChainId
	ScanBlock.Time = time.Now().Unix()
	ScanBlock.LatestBlock = blockNum
	err := mongo.UpdateScanBlock(ScanBlock)
	if err != nil {
		log.Error("MonitorBlock  AddScanBlockPlace UpdateScanBlock")
		return err
	}
	return nil
}
func GetMongodbBlockAndLinkBlock() (mongo.ScanBlock, uint64, error) {
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
				return mongo.ScanBlock{}, 0, err
			}
			mongoBln = bl
		} else {
			return mongo.ScanBlock{}, 0, err
		}
	}
	//链上新块
	NewBlockNumber, err := api.GetBlockNumber()
	if err != nil {
		log.Error("ScanBlocks: api.GetBlockNumber error:", err)
		return mongo.ScanBlock{}, 0, err
	}
	//安全块
	var SafeBlock uint64
	if NewBlockNumber > 10 {
		SafeBlock = NewBlockNumber - 10
	} else {
		SafeBlock = 0
	}
	return mongoBln, SafeBlock, nil
}
func AddLossBlock(safeBlock uint64, reason string) error {
	_, err := mongo.GetLossBlock(safeBlock)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var LossBlock mongo.LossBlock
			LossBlock.NetWork = api.ChainId
			LossBlock.Time = time.Now().Unix()
			LossBlock.BlockNr = safeBlock
			LossBlock.Reason = reason
			err = mongo.AddLossBlock(LossBlock)
			if err != nil {
				log.Error("MonitorBlock  AddLossBlock AddLossBlock")
				return err
			}
		} else {
			log.Infof("区块 %d 已在失败块列表中,错误：%v", safeBlock, err)
			return err
		}
	}
	return nil
}
func WithRetry(maxRetry int, label string, fn func() error) error {
	var err error
	for i := 1; i <= maxRetry; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		sleepDuration := time.Second * time.Duration(math.Pow(2, float64(i-1)))
		log.Warnf("%s 第 %d 次失败: %v，等待 %v 后重试", label, i, err, sleepDuration)
		time.Sleep(sleepDuration)
	}
	log.Errorf("%s 最终失败: %v", label, err)
	return err
}

func ProcessTransactions(bl *types.Block, ctx context.Context) error {
	blockTime := bl.Time()
	for _, tx := range bl.Transactions() {
		from, err := api.GetFromByTransaction(tx)
		if err != nil {
			log.Errorf("MonitorBlock  ProcessTransactions GetFromByTransaction: %v", tx.Hash().String())
			continue
		}
		if tx.To() == nil { //部署合约跳过
			continue
		}
		if strings.ToLower(tx.To().String()) == strings.ToLower(conf.ContractBztAddr) {
			if strings.ToLower(from.String()) != strings.ToLower(conf.OwnerAddress) {
				//拿到开仓数据
				receipt, err := api.GetTransactionReceiptByHash(tx.Hash())
				if err != nil {
					return err
				}
				err = OrderOpenedTrade(tx, receipt, blockTime, from, "OrderOpened")
				if err != nil {
					return err
				}
			} else {
				//空投（用户触发），关仓（平台），需要解析事件区分
				receipt, err := api.GetTransactionReceiptByHash(tx.Hash())
				if err != nil {
					log.Errorf("GetTransactionReceiptByHash err: %v", err)
					return err
				}
				if receipt.Status == 0 {
					_, err = AddTransactionTrade(tx, receipt, from, blockTime, "unknow")
					if err != nil {
						return err
					}
					continue
				}
				_, err = ParseEvents(tx, receipt, blockTime, from)
				if err != nil {
					log.Errorf("ParseEvents err: %v", err)
					return err
				}
			}
		} else {
			continue
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
				log.Infof("🔒 识别为关仓事件: TxHash=%s", receipt.TxHash.Hex())
				isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, "OrderClosed")
				if err != nil {
					return "", fmt.Errorf("OrderClosed : %w", err)
				}
				err = OrderClosedTrade(order, isNewRecord, blTime)
				if err != nil {
					return "", fmt.Errorf(" OrderClosed : %w", err)
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
				log.Infof("🎁 识别为空投事件: TxHash=%s", receipt.TxHash.Hex())
				isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, "Airdrop")
				if err != nil {
					return "", fmt.Errorf("<UNK> Airdrop <UNK>: %w", err)
				}
				err = AirdropTrade(airdrop, blTime, isNewRecord)
				if err != nil {
					log.Errorf("Airdrop : %v", err)
					return "", fmt.Errorf("<UNK> Airdrop <UNK>: %w", err)
				}
				return "airdrop", nil
			}
		}
	}
	log.Infof("📭 未识别事件类型: TxHash=%s", receipt.TxHash.Hex())
	return "unknown", nil
}

func OrderOpenedTrade(tx *types.Transaction, receipt *types.Receipt, blTime uint64, from common.Address, types string) error {
	if receipt.Status == 0 {
		_, err := AddTransactionTrade(tx, receipt, from, blTime, types)
		if err != nil {
			return err
		}
		return nil
	}

	isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, types)
	if err != nil {
		return err
	}
	if isNewRecord {
		event, err := bzt.GetParseOrderOpened(receipt)
		if err != nil {
			return err
		}
		err = mongo.UpdateOrderOpenStatus(event.OrderId.String(), strings.ToLower(tx.Hash().String()), event.Amount.String(), uint64(1))
		if err != nil {
			return err
		}
	}
	return nil
}

func OrderClosedTrade(event *bzt.BztOrderClosed, status bool, blTime uint64) error {
	if !status {
		return nil
	}
	err := mongo.UpdateOrderClosedStatus(event.OrderId.String(), strings.ToLower(event.Raw.TxHash.String()), event.ProfitLoss.String(), uint64(2))
	if err != nil {
		return err
	}
	Order, err := bzt.GetOrders(event.OrderId.Int64())
	if err != nil {
		return err
	}
	if event.ProfitLoss.Sign() >= 0 {
		value, err := api.StringToBigIntDiv(Order.ProfitLoss.String(), "2")
		if err != nil {
			log.Errorf("ProfitLoss > 0 StringToBigIntDiv err: %v", err)
			return err
		}
		err = RewardPool(Order, value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss > 0 RewardPool : %v", err)
			return err
		}
	} else {
		value, err := api.StringToBigIntDiv(Order.ProfitLoss.String(), "-1")
		if err != nil {
			log.Errorf("ProfitLoss < 0 StringToBigIntDiv err: %v", err)
			return err
		}
		err = RewardPool(Order, value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss < 0 RewardPool : %v", err)
			return err
		}
		err = UserLossAmount(Order, value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss < 0 UserLossAmount : %v", err)
			return err
		}
	}
	return nil
}

func AirdropTrade(event *bzt.BztAirdrop, blTime uint64, status bool) error {
	if !status {
		return nil
	}
	t1 := time.Unix(int64(blTime), 0).In(time.Local)
	ts1 := t1.Format("2006-01-02")
	_, err := mongo.GetAirdrop(strings.ToLower(event.Raw.TxHash.String()))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var air mongo.Airdrop
			air.TxHash = strings.ToLower(event.Raw.TxHash.String())
			air.Symbol = "DUSDT"
			air.ToAddr = strings.ToLower(event.Recipient.String())
			air.Amount = event.Amount.String()
			air.AirdropTime = ts1
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
	from common.Address, blTime uint64, types string) (bool, error) {
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
			tx.TransactionType = types
			err = mongo.AddTransaction(tx)
			if err != nil {
				log.Errorf("AddTransaction err: %v", err)
				return false, err
			}
			return true, nil
		} else {
			log.Errorf("GetTransactionerr: %v", err)
			return false, err
		}
	}
	return false, nil
}

func RewardPool(Order *bzt.OrderInfo, value *big.Int, blTime uint64) error {
	Res, err := mongo.GetRewardAmount(Order.TokenName)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var amount mongo.RewardAmount
			amount.Symbol = Order.TokenName
			amount.UpdateAt = blTime
			amount.TotalAmount = value.String()
			amount.AirdropReward = "0"
			err = mongo.AddRewardAmount(amount)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	newValue, err := api.StringToBigIntSum(Res.TotalAmount, value.String())
	if err != nil {
		return err
	}
	err = mongo.UpdateRewardAmount(Order.TokenName, newValue.String())
	if err != nil {
		return err
	}
	return nil
}
func UserLossAmount(Order *bzt.OrderInfo, value *big.Int, blTime uint64) error {
	res, err := mongo.GetUserLossAmount(strings.ToLower(Order.User.String()), Order.TokenName)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var amount mongo.UserLossAmount
			amount.Symbol = Order.TokenName
			amount.LossAmount = value.String()
			amount.UpdateAt = blTime
			amount.UserAddr = strings.ToLower(Order.User.String())
			amount.ClaimAirdrop = "0"
			err = mongo.AddUserLossAmount(amount)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	newValue, err := api.StringToBigIntSum(res.LossAmount, value.String())
	if err != nil {
		return err
	}
	err = mongo.UpdateUserLossAmount(Order.TokenName, strings.ToLower(Order.User.String()), newValue.String())
	if err != nil {
		return err
	}
	return nil
}
