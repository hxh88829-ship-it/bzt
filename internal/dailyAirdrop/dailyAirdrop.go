package dailyAirdrop

import (
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
	"math/big"
	"strings"
	"time"
	"valueguard/internal/api"
	"valueguard/internal/mongo"
)

// StartAirdropCron 初始化每日空投定时任务

func StartAirdropCron(symbols []string) *cron.Cron {
	c := cron.New(cron.WithSeconds())

	// 每天 00:00 执行空投发放
	c.AddFunc("0 0 0 * * *", func() {
		dateStr := time.Now().Format("2006-01-02")
		log.Infof("Start daily airdrop for date %s", dateStr)
		for _, symbol := range symbols {
			resStart, err := mongo.GetRewardAmount(symbol)
			if err != nil {
				log.Warnf("[%s] GetRewardAmount error: %v", symbol, err)
				continue
			}
			if err := GetAirdropByDay([]string{symbol}, resStart); err != nil {
				log.Warnf("[%s] GetAirdropByDay error: %v", symbol, err)
			}
		}
	})

	// 每天 23:59:59 执行空投回收
	c.AddFunc("59 59 23 * * *", func() {
		dateStr := time.Now().Format("2006-01-02")
		log.Infof("Start daily airdrop recovery for date %s", dateStr)
		for _, symbol := range symbols {
			resEnd, err := mongo.GetRewardAmount(symbol)
			if err != nil {
				log.Warnf("[%s] GetRewardAmount error: %v", symbol, err)
				continue
			}
			resDaily, err := mongo.GetDailyAirdrop(dateStr, symbol)
			if err != nil {
				log.Warnf("[%s] GetDailyAirdrop error: %v", symbol, err)
				continue
			}
			if err := AddRewardsToPool([]string{symbol}, resEnd, resDaily, dateStr); err != nil {
				log.Errorf("[%s] AddRewardsToPool error: %v", symbol, err)
			}
		}
	})

	c.Start()
	return c
}

func GetAirdropByDay(symbols []string, res mongo.RewardAmount) error { //取出当天空投
	divisor := big.NewInt(100) // 除数，100 表示百分之一

	for _, symbol := range symbols {
		// total = 奖励池总额（整数）
		total := new(big.Int)
		_, ok := total.SetString(res.TotalAmount, 10)
		if !ok {
			return fmt.Errorf("invalid total amount for symbol %s", symbol)
		}

		// reward = total / 100
		reward := new(big.Int).Div(total, divisor)

		// totalAfter = total - reward
		totalAfter := new(big.Int).Sub(total, reward)

		// 存入数据库（整数转字符串）
		err := mongo.UpdateRewardPool(
			symbol,
			totalAfter.String(),
			reward.String(),
		)
		if err != nil {
			return err
		}
		timestamp := time.Now().Format("2006-01-02")
		_, err = mongo.GetDailyAirdrop(timestamp, symbol)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				var dailyAir mongo.DailyAirdropTrade
				dailyAir.Symbol = symbol
				dailyAir.Remain = reward.String()
				dailyAir.Date = timestamp
				err = mongo.AddDailyAirdrop(dailyAir)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func AddRewardsToPool(symbols []string, res mongo.RewardAmount, resDaily mongo.DailyAirdropTrade, timestamp string) error { // 剩余空投放回奖励池
	for _, symbol := range symbols {

		//  解析奖励池总额（整数）
		total := new(big.Int)
		if _, ok := total.SetString(res.TotalAmount, 10); !ok {
			return fmt.Errorf("invalid total amount for symbol %s: %s", symbol, res.TotalAmount)
		}

		// 解析剩余空投奖励
		reward := new(big.Int)
		if _, ok := reward.SetString(resDaily.Remain, 10); !ok {
			return fmt.Errorf("invalid airdrop reward for symbol %s: %s", symbol, res.AirdropReward)
		}

		//  检查空投奖励是否为负数
		if reward.Sign() < 0 {
			return fmt.Errorf("invalid airdrop reward for symbol %s: negative value %s", symbol, reward.String())
		}

		//  计算新的奖励池总额
		newTotal := new(big.Int).Add(total, reward)

		//  检查新奖励池是否为负数（理论上不会，但做防护）
		if newTotal.Sign() < 0 {
			return fmt.Errorf("resulting reward pool for symbol %s is negative: %s", symbol, newTotal.String())
		}

		// 日志记录
		log.Infof("AddRewardsToPool | Symbol: %s | Total: %s | Reward: %s | NewTotal: %s",
			symbol, total.String(), reward.String(), newTotal.String())

		//  更新数据库（奖励池总额 = newTotal, 空投余额归零）
		if err := mongo.UpdateRewardPool(symbol, newTotal.String(), "0"); err != nil {
			return fmt.Errorf("failed to update reward pool for symbol %s: %w", symbol, err)
		}
	}
	return nil
}

// 按损失占比计算可领取奖励
func CalculateAirdrop(userLoss, totalLoss, totalReward string) (*big.Int, error) {
	// 转换成 big.Float 保留高精度
	userLossF, ok := new(big.Float).SetString(userLoss)
	if !ok {
		return nil, fmt.Errorf("invalid userLoss: %s", userLoss)
	}
	totalLossF, ok := new(big.Float).SetString(totalLoss)
	if !ok {
		return nil, fmt.Errorf("invalid totalLoss: %s", totalLoss)
	}
	totalRewardF, ok := new(big.Float).SetString(totalReward)
	if !ok {
		return nil, fmt.Errorf("invalid totalReward: %s", totalReward)
	}

	// 用户损失占比 = userLoss / totalLoss
	ratio := new(big.Float).Quo(userLossF, totalLossF)

	// 可领取额度 = ratio × totalReward
	reward := new(big.Float).Mul(ratio, totalRewardF)

	// 保留 8 位小数（你可以改成精度 6 或 18）
	rewardStr := fmt.Sprintf("%.0f", reward)
	claims := new(big.Int)
	if _, ok := claims.SetString(rewardStr, 10); !ok {
		return nil, fmt.Errorf("invalid claims: %s", rewardStr)
	}

	return claims, nil
}

// 判断是否领取超额并更新数据
func UpdateLossAmount(addr, symbol string) (*big.Int, string, error) {
	totals, err := mongo.GetRewardAmount(symbol)
	if err != nil {
		return nil, "", err
	}
	users, err := mongo.GetUserLossAmount(strings.ToLower(addr), symbol) //用户当前
	if err != nil {
		return nil, "", err
	}
	claims, err := CalculateAirdrop(users.LossAmount, totals.TotalAmount, totals.AirdropReward) // 今日可领
	if err != nil {
		return nil, "", err
	}
	Claimed, err := api.StringToBigIntSum(users.ClaimAirdrop, claims.String()) // 目前已领
	if err != nil {
		return nil, "", err
	}
	compareRes, err := CompareBigInt(users.LossAmount, Claimed.String()) //领取是否超出以损
	if err != nil {
		return nil, "", err
	}
	if compareRes == -1 {
		return nil, "", errors.New("no airdrop claimed")
	}

	return claims, Claimed.String(), nil
}

// 返回值：-1 表示 a < b，0 表示 a == b，1 表示 a > b
func CompareBigInt(a, b string) (int, error) {
	bigA, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return 0, fmt.Errorf("invalid number string: %s", a)
	}

	bigB, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return 0, fmt.Errorf("invalid number string: %s", b)
	}

	return bigA.Cmp(bigB), nil
}
