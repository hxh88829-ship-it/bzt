package monitorBlock

import (
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-kratos/kratos/v2/log"
	"strings"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/mongo"
)

var BztAddr = "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a"

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
			err = UpdateScanBlockPlace(bl)
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
	//blockTime := bl.Time
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
		}
		if strings.ToLower(from.String()) == strings.ToLower(BztAddr) {
			//空投（用户触发），关仓（平台）
		}
	}
	return nil
}

func UpdateScanBlockPlace(bl *types.Block) error {
	var ScanBlock mongo.ScanBlock
	ScanBlock.NetWork = api.ChainId
	ScanBlock.Time = time.Now().Unix()
	ScanBlock.LatestBlock = bl.Number().Uint64()
	err := mongo.UpdateScanBlock(ScanBlock)
	if err != nil {
		log.Error("MonitorBlock  AddScanBlockPlace UpdateScanBlock")
		return err
	}
	return nil
}
