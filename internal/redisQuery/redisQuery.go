package redisQuery

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
	"valueguard/internal/mongo"
)

var (
	redisKey = "platform_users_set"
	lockKey  = "lock:sync_platform_users" // Redis 分布式锁 key
	lockTTL  = 30 * time.Second           // 锁定时间，防止异常死锁
	RedisCli *redis.Client
)

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

func safeUnlock(ctx context.Context, key, val string) error {
	res, err := unlockScript.Run(ctx, RedisCli, []string{key}, val).Result()
	if err != nil {
		return fmt.Errorf("🔓 Redis 解锁脚本执行失败: %w", err)
	}
	if res == int64(0) {
		log.Warnf("⚠️ Redis 解锁失败：锁已被其他客户端持有或过期 key=%s", key)
	}
	return nil
}

// 自动续期函数，传入 redis 客户端、锁 key、锁值、ttl，返回取消函数
func startLockKeepAlive(ctx context.Context, client *redis.Client, key, val string, ttl time.Duration) context.CancelFunc {
	ctxKeep, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ctxKeep.Done():
				return
			case <-ticker.C:
				lua := `
					if redis.call("get", KEYS[1]) == ARGV[1] then
						return redis.call("pexpire", KEYS[1], ARGV[2])
					else
						return 0
					end
				`
				ok, err := client.Eval(context.Background(), lua, []string{key}, val, int(ttl.Milliseconds())).Int64()
				if err != nil || ok == 0 {
					// 续期失败时停止续期
					cancel()
					return
				}
			}
		}
	}()

	return cancel
}

func WithRedisLock(ctx context.Context, key string, ttl time.Duration, maxRetry int, retryDelay time.Duration, fn func() error) error {
	lockVal := uuid.NewString()

	for i := 0; i <= maxRetry; i++ {
		locked, err := RedisCli.SetNX(ctx, key, lockVal, ttl).Result()
		if err != nil {
			return fmt.Errorf("🔒 Redis 获取锁失败: %w", err)
		}
		if locked {
			// 开启续期
			cancelKeep := startLockKeepAlive(ctx, RedisCli, key, lockVal, ttl)
			defer cancelKeep()

			defer func() {
				if r := recover(); r != nil {
					_ = safeUnlock(context.Background(), key, lockVal)
					panic(r)
				} else {
					_ = safeUnlock(context.Background(), key, lockVal)
				}
			}()
			return fn() // 成功获得锁，执行回调函数
		}

		if i < maxRetry {
			log.Infof("⏳ 锁被占用，重试 %d/%d: key=%s", i+1, maxRetry, key)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("🔁 尝试 %d 次仍未获取锁: %s", maxRetry, key)
}

func SafeSyncPlatformUsersToRedis(ctx context.Context) error {
	log.Info("🔄 正在从 MongoDB 安全同步平台用户地址到 Redis")
	addresses, err := GetUserByMongodb(ctx)
	if err != nil {
		return err
	}
	// 获取分布式锁
	err = WithRedisLock(ctx, lockKey, lockTTL, 3, 300*time.Millisecond, func() error {
		// 临时 Redis key（避免冲突）
		tmpKey := fmt.Sprintf("%s_tmp_%d", redisKey, time.Now().UnixNano())

		// 写入 Redis 临时集合
		if err := RedisCli.SAdd(ctx, tmpKey, addresses...).Err(); err != nil {
			return fmt.Errorf("redis 写入临时 key 失败: %w", err)
		}

		// 原子切换到正式 key
		ok, err := RedisCli.RenameNX(ctx, tmpKey, redisKey).Result()
		if err != nil {
			return fmt.Errorf("redis RENAME 失败: %w", err)
		}
		if !ok {
			return fmt.Errorf("redis RENAME 失败: 目标 key 已存在，可能有并发冲突")
		}

		log.Infof("✅ 成功同步 %d 个平台用户地址到 Redis（安全切换）", len(addresses))
		return nil
	})
	if err != nil {
		log.Errorf("❌ 同步用户地址到 Redis 失败: %v", err)
		return err
	}
	return nil
}

func IsPlatformUser(ctx context.Context, addr string) (bool, error) {
	addr = strings.ToLower(addr)
	mongoCli := mongo.MonCli
	userCol := mongoCli.Client.Database("NftTransaction").Collection("user")

	// 快路径：Redis
	exists, err := RedisCli.SIsMember(ctx, redisKey, addr).Result()
	if err == nil && exists {
		return true, err
	}

	// 慢路径：MongoDB 兜底
	count, err := userCol.CountDocuments(ctx, bson.M{
		"address": addr,
		"status":  "0", // 或 0，取决于你的字段类型
	})
	if err != nil {
		log.Infof("⚠️ MongoDB 查询失败: %v", err)
		return false, err
	}

	if count > 0 {
		// Redis 缓存漏了？可异步修复写回
		_ = RedisCli.SAdd(ctx, redisKey, addr) // 不阻塞逻辑
		return true, err
	}

	return false, err
}

func GetUserByMongodb(ctx context.Context) ([]interface{}, error) {
	mongoCli := mongo.MonCli
	userCol := mongoCli.Client.Database("NftTransaction").Collection("user")

	filter := bson.M{"status": "0"}
	projection := bson.M{"address": 1}
	cursor, err := userCol.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		log.Errorf("MongoDB 查询失败: %w", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	totalCount, _ := userCol.CountDocuments(ctx, bson.M{})
	validCount, _ := userCol.CountDocuments(ctx, filter)
	log.Infof("👥 用户总数: %d，状态正常: %d", totalCount, validCount)

	var addresses []interface{}
	for cursor.Next(ctx) {
		var user struct {
			Address string `bson:"address"`
		}
		if err := cursor.Decode(&user); err == nil && user.Address != "" {
			addresses = append(addresses, strings.ToLower(user.Address))
		} else if err != nil {
			log.Warnf("❗ 解码用户失败: %v", err)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Warnf("MongoDB 游标错误: %v", err)
		return nil, err
	}
	if len(addresses) == 0 {
		log.Infof("MongoDB 没有读取到任何地址，取消同步")
		return nil, errors.New("未读取到任何地址")
	}
	return addresses, nil
}
