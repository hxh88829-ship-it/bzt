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

func StartAirdropCron() *cron.Cron {
	loc := time.Now().Location()
	log.Infof(loc.String())
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(loc),
	)
	//TODO
	// 每天 00:00 执行空投发放
	c.AddFunc("0 0 0 * * *", func() {
		dateStr := time.Now().Format("2006-01-02")
		log.Infof("Start daily airdrop for date %s", dateStr)
		resStart, err := mongo.GetRewardAmount("DUSDT")
		if err != nil {
			log.Warnf("StartAirdropCron  GetRewardAmount error: %v", err)
			return
		}
		if err := GetAirdropByDay(resStart); err != nil {
			log.Warnf("StartAirdropCron GetAirdropByDay error: %v", err)
			return
		}

	})

	c.Start()
	return c
}

func GetAirdropByDay(res mongo.RewardAmount) error { //取出当天空投
	divisor := big.NewInt(100) // 除数，100 表示百分之一
	// total = 奖励池总额（整数）
	total := new(big.Int)
	_, ok := total.SetString(res.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("GetAirdropByDay error: %v", res.TotalAmount)
	}

	// reward = total / 100
	reward := new(big.Int).Div(total, divisor)

	// totalAfter = total - reward
	totalAfter := new(big.Int).Sub(total, reward)

	timestamp := time.Now().Format("2006-01-02")
	airdrop, err := mongo.GetDailyAirdropBySymbol("DUSDT")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var dailyAir mongo.DailyAirdropTrade
			dailyAir.Symbol = "DUSDT"
			dailyAir.Reward = reward.String()
			dailyAir.Date = timestamp
			dailyAir.PoolTotal = total.String()
			err = mongo.AddDailyAirdrop(dailyAir)
			if err != nil {
				log.Errorf("GetAirdropByDay AddDailyAirdrop error: %v", err)
				return err
			}
			log.Info("Reward: ", reward.String())
			err = mongo.UpdateRewardPool(
				"DUSDT",
				totalAfter.String(),
			)
			if err != nil {
				log.Errorf("GetAirdropByDay UpdateRewardPool error: %v", err)
				return err
			}
			return nil
		} else {
			log.Errorf("GetAirdropByDay error: %v", err)
			return err
		}
	}
	if airdrop.Date != timestamp {
		// 存入数据库（整数转字符串）
		err := mongo.UpdateRewardPool(
			"DUSDT",
			totalAfter.String(),
		)
		if err != nil {
			log.Errorf("GetAirdropByDay UpdateRewardPool error: %v", err)
			return err
		}
		rewarded := new(big.Int) //昨日剩余空投
		_, ok = rewarded.SetString(airdrop.Reward, 10)
		if !ok {
			return fmt.Errorf("GetAirdropByDay error: %v", airdrop.Reward)
		}
		totalReward := new(big.Int).Add(rewarded, reward)
		log.Infof("GetAirdropByDay totalReward: %v  rewarded: %v  Reward: %v", totalReward, rewarded, reward)
		err = mongo.UpdateDailyAirdropRemain(totalReward.String(), "DUSDT", timestamp, total.String())
		if err != nil {
			log.Errorf("GetAirdropByDay UpdateDailyAirdropRemain error: %v", err)
			return err
		}
		return nil
	}
	log.Warnf("重复发放以回滚")

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
func UpdateLossAmount(addr string) (*big.Int, string, string, error) {
	reward, err := mongo.GetDailyAirdropBySymbol("DUSDT")
	if err != nil {
		log.Errorf("UpdateLossAmount GetDailyAirdropBySymbol error: %v", err)
		return nil, "", "", err
	}
	users, err := mongo.GetUserAmount(strings.ToLower(addr)) //用户当前
	if err != nil {
		log.Errorf("UpdateLossAmount GetUserLossAmount error: %v", err)
		return nil, "", "", err
	}
	log.Infof("UpdateLossAmount, users: %s", users.LossAmount)
	claims, err := CalculateAirdrop(users.LossAmount, reward.PoolTotal, reward.Reward) // 今日可领
	if err != nil {
		log.Errorf("UpdateLossAmount CalculateAirdrop error: %v", err)
		return nil, "", "", err
	}
	Claimed, err := api.StringToBigIntSum(users.ClaimAirdrop, claims.String()) // 目前已领加今日可领
	if err != nil {
		log.Errorf("UpdateLossAmount api.StringToBigIntSum error: %v", err)
		return nil, "", "", err
	}
	compareRes, err := CompareBigInt(users.LossAmount, Claimed.String()) //领取是否超出以损
	if err != nil {
		log.Errorf("UpdateLossAmount CompareBigInt error: %v", err)
		return nil, "", "", err
	}
	if compareRes == -1 {
		return nil, "", "", errors.New("no airdrop claimed")
	}

	return claims, Claimed.String(), users.LossAmount, nil
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
