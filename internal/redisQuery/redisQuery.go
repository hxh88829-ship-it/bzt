package redisQuery

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
	"valueguard/internal/mongo"
)

//var (
//	redisKey   = "platform_users_set"
//	lockKey    = "lock:sync_platform_users" // Redis 分布式锁 key
//	lockTTL    = 30 * time.Second           // 锁定时间，防止异常死锁
//	retryDelay = 300 * time.Millisecond
//	RedisCli   *redis.Client
//)
//
//var unlockScript = redis.NewScript(`
//if redis.call("get", KEYS[1]) == ARGV[1] then
//	return redis.call("del", KEYS[1])
//else
//	return 0
//end`)
//
//type LockHandle struct {
//	UnlockFunc context.CancelFunc // 用于停止续期和释放锁
//	LockVal    string             // 锁值，用于安全解锁
//	Locked     bool               // 是否成功拿锁
//}

//func safeUnlock(ctx context.Context, key, val string) error {
//	res, err := unlockScript.Run(ctx, RedisCli, []string{key}, val).Result()
//	if err != nil {
//		return fmt.Errorf("🔓 Redis 解锁脚本执行失败: %w", err)
//	}
//	if res == int64(0) {
//		log.Warnf("⚠️ Redis 解锁失败：锁已被其他客户端持有或过期 key=%s val=%s", key, val)
//	}
//	return nil
//}

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
					log.Warnf("⚠️ Redis 锁续期失败，可能已被删除或锁值不一致: key=%s val=%s err=%v", key, val, err)
					cancel()
					return
				}
			}
		}
	}()

	return cancel
}

// WithRedisLockSeparate 仅尝试加锁，带自动续期和安全解锁，
//func WithRedisLockSeparates(
//	ctx context.Context,
//	key string,
//	ttl time.Duration,
//	maxRetry int,
//	retryDelay time.Duration,
//) (*LockHandle, error) {
//	lockVal := uuid.NewString()
//
//	for i := 0; i <= maxRetry; i++ {
//		locked, err := RedisCli.SetNX(ctx, key, lockVal, ttl).Result()
//		if err != nil {
//			return nil, fmt.Errorf("🔒 Redis 获取锁失败: %w", err)
//		}
//		if locked {
//			// 开启续期
//			cancelKeep := startLockKeepAlive(ctx, RedisCli, key, lockVal, ttl)
//
//			// 返回一个函数用于释放锁（取消续期 + 执行解锁）
//			unlockFunc := func() {
//				cancelKeep()
//				_ = safeUnlock(context.Background(), key, lockVal)
//			}
//
//			return &LockHandle{
//				UnlockFunc: unlockFunc,
//				LockVal:    lockVal,
//				Locked:     true,
//			}, nil
//		}
//
//		if i < maxRetry {
//			log.Infof("⏳ 锁被占用，重试 %d/%d: key=%s", i+1, maxRetry, key)
//			time.Sleep(retryDelay)
//		}
//	}
//
//	return &LockHandle{Locked: false}, fmt.Errorf("🔁 尝试 %d 次仍未获取锁: %s", maxRetry, key)
//}

//func SafeSyncPlatformUserToRedis(ctx context.Context) error {
//	log.Info("🔄 正在从 MongoDB 安全同步平台用户地址到 Redis")
//	addresses, err := GetUserByMongodb(ctx)
//	if err != nil {
//		return err
//	}
//	// 加分布式锁
//	lockHandle, err := WithRedisLockSeparate(ctx, lockKey, lockTTL, 3, retryDelay)
//
//	if err != nil {
//		return err
//	}
//	if !lockHandle.Locked {
//		return errors.New("锁被占用")
//	}
//	defer lockHandle.UnlockFunc() // 业务结束后安全释放锁
//	// 1. 写入临时 key（存放 MongoDB 中所有地址）
//	tmpKey := fmt.Sprintf("%s_tmp_%d", redisKey, time.Now().UnixNano())
//	if err := RedisCli.SAdd(ctx, tmpKey, addresses...).Err(); err != nil {
//		return fmt.Errorf("❌ Redis 写入临时 key 失败: %w", err)
//	}
//
//	// 2. 将 Redis 现有 key 内容合并进临时 key
//	//    即 tmpKey = tmpKey ∪ redisKey
//	//    结果仍然写入 tmpKey
//	if err := RedisCli.SUnionStore(ctx, tmpKey, tmpKey, redisKey).Err(); err != nil {
//		return fmt.Errorf("❌ Redis 合并现有 key 失败: %w", err)
//	}
//
//	// 3. 删除旧 key（以便 RENAME 不冲突）
//	if err := RedisCli.Del(ctx, redisKey).Err(); err != nil {
//		return fmt.Errorf("❌ Redis 删除旧 key 失败: %w", err)
//	}
//
//	// 4. 将合并后的临时 key 改名为正式 key
//	if err := RedisCli.Rename(ctx, tmpKey, redisKey).Err(); err != nil {
//		return fmt.Errorf("❌ Redis 重命名失败: %w", err)
//	}
//
//	log.Infof("✅ 成功同步并合并 %d 个平台用户地址到 Redis", len(addresses))
//	return nil
//}

//func IsPlatformUsers(ctx context.Context, addr string) (bool, error) {
//	addr = strings.ToLower(addr)
//	mongoCli := mongo.MonCli
//	userCol := mongoCli.Client.Database("bzt").Collection("user")
//
//	// 快路径：Redis
//	exists, err := RedisCli.SIsMember(ctx, redisKey, addr).Result()
//	if err == nil && exists {
//		return true, err
//	}
//	if err != nil {
//		log.Warnf("⚠️ Redis 查询失败，回退 MongoDB: %v", err)
//	}
//
//	// 慢路径：MongoDB 兜底
//	count, err := userCol.CountDocuments(ctx, bson.M{
//		"address": addr,
//		"status":  "0", // 或 0，取决于你的字段类型
//	})
//	if err != nil {
//		log.Infof("⚠️ MongoDB 查询失败: %v", err)
//		return false, err
//	}
//
//	if count > 0 {
//		// Redis 缓存漏了？异步修复写回，避免主逻辑阻塞
//		go func(addr string) {
//			if err := RedisCli.SAdd(context.Background(), redisKey, addr).Err(); err != nil {
//				log.Warnf("⚠️ Redis 异步写回失败: addr=%s err=%v", addr, err)
//			}
//		}(addr)
//		return true, nil
//	}
//
//	return false, err
//}

func GetUsersByMongodb(ctx context.Context) ([]interface{}, error) {
	mongoCli := mongo.MonCli
	userCol := mongoCli.Client.Database("bzt").Collection("user")
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
