package monitorBlock

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
	"strings"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/bzt"
	"valueguard/internal/conf"
	"valueguard/internal/mongo"
)

func ScanBlocks(ctx context.Context) (uint64, error) {
	mongoBln, newBlockNumber, err := GetMongodbBlockAndLinkBlock()
	if err != nil {
		return 0, fmt.Errorf("GetMongodbBlockAndLinkBlock failed: %w", err)
	}

	if mongoBln > newBlockNumber {
		log.Warnf("数据库块 %d 高于链上最新块 %d，可能回滚", mongoBln, newBlockNumber)
		return 0, fmt.Errorf("database block ahead of chain")
	}

	if mongoBln == newBlockNumber {
		//	log.Info("数据库块高", mongoBln.LatestBlock, "\n", "link:", newBlockNumber)
		return 0, errors.New("block is latest")
	}

	//只处理一个块
	blockNum := mongoBln + 1
	err = ScanOneBlock(ctx, blockNum)
	if err != nil {
		// ❗ 出错，立即返回，下一轮从这个块重新开始
		log.Errorf("ScanOneBlock failed at block %d: %v", blockNum, err)
		return 0, err
	}
	return blockNum, nil
}
func ScanOneBlock(ctx context.Context, blockNum uint64) error {
	bl, err := api.GetBlockByNumber(blockNum)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber failed: %w", err)
	}

	err = ProcessTransactions(bl, ctx)
	if err != nil {
		return fmt.Errorf("ProcessTransactions failed: %w", err)
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
func GetMongodbBlockAndLinkBlock() (uint64, uint64, error) {
	chainId := api.ChainId

	//链上新块
	NewBlockNumber, err := api.GetBlockNumber()
	if err != nil {
		log.Error("ScanBlocks: api.GetBlockNumber error:", err)
		return 0, 0, err
	}

	// 1. 从Mongo获取上次扫描的区块号
	mongoBln, err := mongo.GetScanBlock(chainId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var bl mongo.ScanBlock
			bl.NetWork = chainId
			bl.Time = time.Now().Unix()
			bl.LatestBlock = NewBlockNumber
			err = mongo.AddScanBlock(bl)
			if err != nil {
				return 0, 0, err
			}
			mongoBln = bl.LatestBlock
		} else {
			return 0, 0, err
		}
	}

	//安全块
	//var SafeBlock uint64
	//if NewBlockNumber > 10 {
	//	SafeBlock = NewBlockNumber - 10
	//} else {
	//	SafeBlock = 0
	//}

	//if mongoBln < 11696216 {
	//	mongoBln = 11696216
	//}
	return mongoBln, NewBlockNumber, nil
}

func ProcessTransactions(bl *types.Block, ctx context.Context) error {
	blockTime := bl.Time()
	for _, tx := range bl.Transactions() {
		if tx.To() == nil { //部署合约跳过
			continue
		}
		if strings.ToLower(tx.To().String()) == strings.ToLower(conf.ContractBztAddr) {
			from, err := api.GetFromByTransaction(tx)
			if err != nil {
				log.Errorf("MonitorBlock  ProcessTransactions GetFromByTransaction: %v", tx.Hash().String())
				continue
			}
			if strings.ToLower(from.String()) != strings.ToLower(conf.OwnerAddress) {
				//拿到开仓数据
				receipt, err := api.GetTransactionReceiptByHash(tx.Hash())
				if err != nil {
					log.Errorf("MonitorBlock  ProcessTransactions GetTransactionReceiptByHash: %v", tx.Hash().String())
					return err
				}
				err = OrderOpenedTrade(tx, receipt, blockTime, from, "OrderOpened")
				if err != nil {
					log.Errorf("MonitorBlock  ProcessTransactions GetTransactionReceiptByHash: %v", tx.Hash().String())
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
						log.Errorf("MonitorBlock  ProcessTransactions AddTransactionTrade: %v", tx.Hash().String())
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
				log.Errorf("GetParseOrderClosed err: %v", err)
				return "", fmt.Errorf("解析 OrderClosed 失败: %w", err)
			}
			if order != nil {
				// 关仓事件数据记录
				log.Infof("🔒 识别为关仓事件: TxHash=%s", receipt.TxHash.Hex())
				isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, "OrderClosed")
				if err != nil {
					log.Errorf("AddTransactionTrade err: %v", err)
					return "", fmt.Errorf("OrderClosed : %w", err)
				}
				err = OrderClosedTrade(order, isNewRecord, blTime)
				if err != nil {
					log.Errorf("OrderClosedTrade err: %v", err)
					return "", fmt.Errorf(" OrderClosed : %w", err)
				}
				return "order_closed", nil
			}
		case AirdropTopic:
			// 解析 空投 事件
			airdrop, err := bzt.GetParseAirdrop(receipt)
			if err != nil {
				log.Errorf("GetParseAirdrop err: %v", err)
				return "", fmt.Errorf("解析 Airdrop 失败: %w", err)
			}
			if airdrop != nil {
				//空投事件数据记录
				log.Infof("🎁 识别为空投事件: TxHash=%s", receipt.TxHash.Hex())
				isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, "Airdrop")
				if err != nil {
					log.Errorf("AddTransactionTrade err: %v", err)
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
			log.Errorf("AddTransactionTrade err: %v", err)
			return err
		}
		return nil
	}

	isNewRecord, err := AddTransactionTrade(tx, receipt, from, blTime, types)
	if err != nil {
		log.Errorf("AddTransactionTrade err: %v", err)
		return err
	}
	if isNewRecord {
		event, err := bzt.GetParseOrderOpened(receipt)
		if err != nil {
			log.Errorf("GetParseOrderOpened err: %v", err)
			return err
		}
		err = mongo.UpdateOrderOpenStatus(event.OrderId.String(), strings.ToLower(tx.Hash().String()), event.Amount.String(), uint64(1))
		if err != nil {
			log.Errorf("UpdateOrderOpenStatus err: %v", err)
			return err
		}
	}
	return nil
}

func OrderClosedTrade(event *bzt.BztOrderClosed, status bool, blTime uint64) error {
	if !status {
		return nil
	}
	err := mongo.UpdateOrderClosedStatus(event.OrderId.String(), event.ProfitLoss.String(), uint64(2))
	if err != nil {
		log.Errorf("UpdateOrderClosedStatus err: %v", err)
		return err
	}
	user := strings.ToLower(event.User.String())
	// 使用 event.ProfitLoss 进行判断和计算
	profitLoss := event.ProfitLoss
	var value *big.Int

	if profitLoss.Sign() >= 0 {
		// 除以2，整数除法
		value = new(big.Int).Div(profitLoss, big.NewInt(2))
	} else {
		// 取绝对值，相当于乘以-1，但使用Abs方法更安全
		value = new(big.Int).Abs(profitLoss)
	}
	if value.Sign() < 0 {
		return errors.New(" value must be non-negative")
	}

	// 现在根据正负执行不同的逻辑
	if profitLoss.Sign() >= 0 {
		err = RewardPool(value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss >= 0 RewardPool : %v", err)
			return err
		}
		err = UserProfitAmount(user, value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss >= 0 UserProfitAmount : %v", err)
			return err
		}
	} else {
		err = RewardPool(value, blTime)
		if err != nil {
			log.Errorf("ProfitLoss < 0 RewardPool : %v", err)
			return err
		}
		err = UserLossAmount(user, value, blTime)
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
	err := mongo.UpdateAirdropStatus(strings.ToLower(event.Raw.TxHash.String()), uint64(1))
	if err != nil {
		log.Errorf("UpdateAirdropStatus err: %v", err)
		return err
	}

	return nil
}

func AddTransactionTrade(txh *types.Transaction, receipt *types.Receipt,
	from common.Address, blTime uint64, types string) (bool, error) {
	_, err := mongo.GetTransaction(strings.ToLower(receipt.TxHash.String()))
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

func RewardPool(value *big.Int, blTime uint64) error {
	Res, err := mongo.GetRewardAmount("DUSDT")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var amount mongo.RewardAmount
			amount.Symbol = "DUSDT"
			amount.UpdateAt = blTime
			amount.TotalAmount = value.String()
			err = mongo.AddRewardAmount(amount)
			if err != nil {
				log.Errorf("AddRewardAmount err: %v", err)
				return err
			}
			return nil
		} else {
			log.Errorf("GetRewardAmounterr: %v", err)
			return err
		}
	}
	newValue, err := api.StringToBigIntSum(Res.TotalAmount, value.String())
	if err != nil {
		log.Errorf("GetRewardAmounterr: %v", err)
		return err
	}
	err = mongo.UpdateRewardPool("DUSDT", newValue.String())
	if err != nil {
		log.Errorf("UpdateRewardAmounterr: %v", err)
		return err
	}
	return nil
}
func UserLossAmount(user string, value *big.Int, blTime uint64) error {
	log.Infof("UserLossAmount start user=%s value=%s blockTime=%d", user, value.String(), blTime)

	res, err := mongo.GetUserAmount(user, "DUSDT")
	if err != nil {
		log.Infof("GetUserAmount err: %v", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			var amount mongo.UserAmount
			amount.Symbol = "DUSDT"
			amount.LossAmount = value.String()
			amount.UpdateAt = int64(blTime)
			amount.UserAddr = user
			amount.ClaimAirdrop = "0"
			amount.Profit = "0"
			err := mongo.AddUserAmount(amount)
			if err != nil {
				log.Errorf("AddUserAmount failed: %v", err)
				return err
			}
			return nil
		}
		return err
	}

	log.Infof("GetUserAmount found existing record: %+v", res)
	newValue, err := api.StringToBigIntSum(res.LossAmount, value.String())
	if err != nil {
		log.Errorf("StringToBigIntSum error: %v", err)
		return err
	}
	log.Infof("lossAmount:= %s addvalue:= %s  newValue:= %s", res.LossAmount, value, newValue.String())
	err = mongo.UpdateUserAmount("DUSDT", user, newValue.String())
	if err != nil {
		log.Errorf("UpdateUserAmount failed: %v", err)
		return err
	}
	log.Infof("UpdateUserAmount success user=%s newLoss=%s", user, newValue.String())
	return nil
}

func UserProfitAmount(user string, value *big.Int, blTime uint64) error {
	log.Infof("UserProfitAmount start user=%s value=%s blockTime=%d", user, value.String(), blTime)

	res, err := mongo.GetUserAmount(user, "DUSDT")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var amount mongo.UserAmount
			amount.Symbol = "DUSDT"
			amount.LossAmount = "0"
			amount.UpdateAt = int64(blTime)
			amount.UserAddr = user
			amount.ClaimAirdrop = "0"
			amount.Profit = value.String()
			err = mongo.AddUserAmount(amount)
			if err != nil {
				log.Errorf("AddUserAmount err: %v", err)
				return err
			}
			return nil
		} else {
			log.Errorf("UserProfitAmount: %v", err)
			return err
		}
	}
	log.Infof("GetUserAmount found existing record: %+v", res)
	newValue, err := api.StringToBigIntSum(res.Profit, value.String())
	if err != nil {
		log.Errorf("GetUserLossAmounterr: %v", err)
		return err
	}
	log.Infof("UserProfitAmount:= %s addvalue:= %s  newValue:=%s", res.Profit, value, newValue.String())
	err = mongo.UpdateUserProfit("DUSDT", user, newValue.String())
	if err != nil {
		log.Errorf("UpdateUserLossAmount err: %v", err)
		return err
	}
	return nil
}
