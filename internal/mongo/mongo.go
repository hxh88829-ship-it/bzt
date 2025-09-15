package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

// MongoClient 封装纯客户端连接
type MongoClient struct {
	Client     *mongo.Client
	cancelFunc context.CancelFunc
}

const MaxPricePerSymbol = 30

var ErrNoDocuments error = errors.New("mongo: no documents in result")
var ErrNoFields error = errors.New("no fields to update")
var ErrAlreadyClaim = errors.New("address already claimed")

const MarketContract = "0x31f3EB0f255178B0fA3FeCbFe7B5314f38949a4B"

// NewMongoClient 创建新的 MongoDB 连接客户端
func NewMongoClient(uri string) (*MongoClient, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// 配置客户端选项
	clientOptions := options.Client().ApplyURI(uri)

	// 建立连接
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cancel() // 立即释放上下文资源
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 验证连接
	if err = client.Ping(ctx, nil); err != nil {
		cancel()
		_ = client.Disconnect(ctx) // 尝试断开无效连接
		return nil, fmt.Errorf("MONGODB connection verification failed: %w", err)
	}

	return &MongoClient{
		Client:     client,
		cancelFunc: cancel,
	}, nil
}

// Close 安全关闭连接
func (mc *MongoClient) Close() error {
	// 创建新的上下文用于关闭操作
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行断开连接
	if err := mc.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %w", err)
	}

	// 调用上下文取消函数
	mc.cancelFunc()

	fmt.Println("MongoDB connection closed gracefully")
	return nil
}

// Ping 检查连接活性
func (mc *MongoClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return mc.Client.Ping(ctx, nil)
}

// 用户内容
func AddUser(a Users) error {
	//判断全局变量MonCli是否为空，
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddUser(a Users) ")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddUser InsertOne err: ", err)
		return errors.New("add user fail")
	}
	return nil
}
func GetUser(addr string) (Users, error) {
	if MonCli == nil {
		return Users{}, errors.New("mongo client is nil " + "GetUser")
	}
	filter := bson.D{{"address", addr}}
	var ma Users
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		log.Error("GetUser FindOne err: ", err)
		return Users{}, ErrNoDocuments
	}
	return ma, nil
}
func UpdateUser(addr, OriginalMessage string) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil")
	}
	filter := bson.D{{"address", addr}}
	update := bson.D{{"$set", bson.D{
		{"original_message", OriginalMessage},
	},
	}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).UpdateMany(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateUser fail ", err)
		return err
	}
	return nil
}

// 最新行情价格
func AddPrice(v CoinPrice) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddPrice")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).InsertOne(context.Background(), v)
	if err != nil {
		log.Error("AddPrice InsertOne err: ", err)
		return errors.New("add price fail")
	}
	return nil
}
func GetPriceForIndex(symbol string, ind uint64) (CoinPrice, error) {
	if MonCli == nil {
		return CoinPrice{}, errors.New("error:mongo.Client is nil" + "GetPrice")
	}
	filter := bson.D{{"symbol", symbol}, {"index", ind}}
	var ma CoinPrice
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		log.Error("GetPriceForIndex FindOne err: ", err)
		return CoinPrice{}, ErrNoDocuments
	}
	return ma, nil
}
func SavePrice(symbol, price string, ind, times uint64) error {
	if MonCli == nil {
		return errors.New("mongo client is nil")
	}
	filter := bson.D{{"symbol", symbol}, {"index", ind}}
	update := bson.D{{"$set", bson.D{
		{"price", price},
		{"timestamp", times},
	}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("SavePrice UpdateLogin err: ", err)
		return errors.New("save price fail")
	}
	return nil
}
func GetPriceByTimestamp(blockTime uint64, symbol string) (CoinPrice, error) {
	if MonCli == nil {
		return CoinPrice{}, errors.New("mongo client is nil" + "GetPriceByTimestamp")
	}
	filter := bson.M{
		"symbol": symbol, // 资产代号，如 "BTCUSDT"
		// TODO 是否可以对时间判断选择小于等于的时间戳
		"timestamp": bson.M{"$lte": blockTime}, // 小于等于块时间戳
	}
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	var priceRecord CoinPrice
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).FindOne(context.Background(), filter, opts).Decode(&priceRecord)
	if err != nil {
		log.Error("GetPriceByTimestamp  FindOne err: ", err)
		return CoinPrice{}, ErrNoDocuments
	}
	return priceRecord, nil
}
func GetPriceBySymbol(symbol string, start, end uint64) ([]CoinPrice, error) {
	if MonCli == nil {
		return nil, errors.New("mongo client is nil: GetPriceBySymbol")
	}

	collection := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice)

	filter := bson.M{
		"symbol": symbol,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // 时间升序排列

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Error("GetPriceBySymbol FindOne err: ", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []CoinPrice
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
func GetPriceBySymbolPaged(symbol string, start, end uint64, page, Size int64) ([]CoinPrice, error) {
	if MonCli == nil {
		return nil, errors.New("mongo client is nil: GetPriceBySymbol")
	}
	filter := bson.M{
		"symbol": symbol,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	if page <= 0 {
		page = 1
	}
	if Size <= 0 || Size > 100 {
		Size = 10 // 默认每页 10 条，最大不超过 100
	}

	skip := (page - 1) * Size

	// 可选项：分页/排序
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetSkip(skip).SetLimit(Size)

	cursor, err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).Find(context.Background(), filter, opts)
	if err != nil {
		log.Error("GetPriceBySymbolPaged FindOne err: ", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var res []CoinPrice
	if err := cursor.All(context.Background(), &res); err != nil {
		log.Error("GetPriceBySymbolPaged cursor FindOne err: ", err)
		return nil, err
	}
	return res, nil
}

// 订单
func AddOrder(a Order) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "OrderOpen")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddOrder InsertOne err: ", err)
		return errors.New("add order fail")
	}
	return nil
}
func GetOrder(OrderId string) (Order, error) {
	if MonCli == nil {
		return Order{}, errors.New("mongo client is nil" + "GetOrder")
	}
	filter := bson.D{{"order_id", OrderId}}
	var ma Order
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		log.Error("GetOrder FindOne err: ", err)
		return Order{}, ErrNoDocuments
	}
	return ma, nil
}
func GetOrderForAll(addr string, page, size int64) ([]Order, error) {
	if MonCli == nil {
		return nil, errors.New("error: mongo.Client is nil -> GetOrderForAll")
	}

	// 构建查询条件
	filter := bson.M{}
	if addr != "" {
		filter["users_addr"] = strings.ToLower(addr)
	}

	// 分页参数处理
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	opts := options.Find().
		SetSort(bson.D{{"create_time", -1}}). // 按创建时间倒序
		SetSkip((page - 1) * size).           // 跳过前面 (page-1)*size 条
		SetLimit(size)                        // 限制返回 size 条

	// 查询
	cursor, err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(order).Find(context.Background(), filter, opts)
	if err != nil {
		log.Error("GetOrderForAll Find err: ", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var res []Order
	if err := cursor.All(context.Background(), &res); err != nil {
		log.Error("GetOrderForAll Cursor decode err: ", err)
		return nil, err
	}
	return res, nil
}

func UpdateOrderClose(OrderId, ClosePri, txHash string, timestamp uint64) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateOrder")
	}
	filter := bson.D{{"order_id", OrderId}}
	update := bson.D{{"$set", bson.D{
		{"close_price", ClosePri},
		{"order_end_time", timestamp},
		{"close_tx_hash", txHash},
	}}}

	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateOrderClose  err: ", err)
		return errors.New("UpdateOrderClose fail")
	}
	return nil
}
func UpdateOrderOpenStatus(OrderId, OpenTx, Amount string, IsClosed uint64) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateOrderStatus")
	}
	filter := bson.D{{"order_id", OrderId}}
	update := bson.D{{"$set", bson.D{
		{"is_closed", IsClosed},
		{"open_tx_hash", OpenTx},
		{"amount", Amount},
	}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateOrderOpenStatus err: ", err)
		return errors.New("UpdateOrderOpenStatus fail")
	}
	return nil
}
func UpdateOrderClosedStatus(OrderId, Profit string, IsClosed uint64) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateOrderStatus")
	}
	filter := bson.D{{"order_id", OrderId}}
	update := bson.D{{"$set", bson.D{
		{"is_closed", IsClosed},
		{"profit_loss", Profit},
	}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateOrderClosedStatus err: ", err)
		return errors.New("UpdateOrderClosedStatus fail")
	}
	return nil
}
func CountOpenOrdersByAddress(address string) (int64, error) {
	if MonCli == nil {
		return 0, errors.New("mongo client is nil" + "CountOpenOrdersByAddress")
	}
	filter := bson.M{
		"users_addr": strings.ToLower(address),
		"is_closed": bson.M{
			"$in": []uint64{0, 1}, // 未结算状态
		},
	}
	count, err := MonCli.Client.Database(DatabaseNameForChain).Collection(order).CountDocuments(context.Background(), filter)
	if err != nil {
		log.Error("CountOpenOrdersByAddress err: ", err)
		return 0, err
	}
	return count, nil
}

// 奖励池
func AddRewardAmount(a RewardAmount) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "AddRewardAmount")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(rewardPool).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddRewardAmount InsertOne err: ", err)
		return errors.New("add loss amount fail")
	}
	return nil
}
func GetRewardAmount(tokenName string) (RewardAmount, error) {
	if MonCli == nil {
		return RewardAmount{}, errors.New("mongo client is nil" + "GetRewardAmount")
	}
	filter := bson.M{
		"symbol": tokenName,
	}
	var loss RewardAmount
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(rewardPool).FindOne(context.Background(), filter).Decode(&loss)
	if err != nil {
		log.Error("GetRewardAmount FindOne err: ", err)
		return RewardAmount{}, ErrNoDocuments
	}
	return loss, nil
}

func UpdateRewardPool(tokenName, total string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateRewardPool")
	}
	filter := bson.M{
		"symbol": tokenName,
	}
	update := bson.D{
		{"$set", bson.D{
			{"total_amount", total},
			{"update_at", time.Now().Unix()},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(rewardPool).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateRewardPool err: ", err)
		return errors.New("update amount fail")
	}
	return nil
}

// 用户亏损记录
func AddUserAmount(a UserAmount) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "AddUserAmount")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddUserAmount InsertOne err: ", err)
		return errors.New("add user  amount fail")
	}
	return nil
}
func GetUserAmount(addr string) (UserAmount, error) {
	if MonCli == nil {
		return UserAmount{}, errors.New("mongo client is nil" + "GetUserAmount")
	}
	filter := bson.M{
		"user_addr": strings.ToLower(addr),
	}
	var loss UserAmount
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount).FindOne(context.Background(), filter).Decode(&loss)
	if err != nil {
		log.Error("GetUserAmount FindOne err: ", err)
		return UserAmount{}, ErrNoDocuments
	}
	return loss, nil
}
func UpdateUserAmount(addr, Amount string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateUserAmount")
	}
	filter := bson.M{
		"user_addr": strings.ToLower(addr),
	}
	update := bson.D{
		{"$set", bson.D{
			{"loss_amount", Amount},
			{"update_at", time.Now().Unix()},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateUserAmount err: ", err)
		return errors.New("update loss amount fail")
	}
	return nil
}
func UpdateUserProfit(addr, Amount string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateUserProfit")
	}
	filter := bson.M{
		"user_addr": strings.ToLower(addr),
	}
	update := bson.D{
		{"$set", bson.D{
			{"profit", Amount},
			{"update_at", time.Now().Unix()},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateUserAmount err: ", err)
		return errors.New("update profit amount fail")
	}
	return nil
}
func UpdateUserClaims(addr, Amount string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateUserClaims")
	}
	filter := bson.M{
		"user_addr": strings.ToLower(addr),
	}
	update := bson.D{
		{"$set", bson.D{
			{"claim_airdrop", Amount},
			{"update_at", time.Now().Unix()},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateUserClaims err: ", err)
		return errors.New("UpdateUserClaims fail")
	}
	return nil
}

// 空投记录
func AddAirdrop(air Airdrop) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "AddAirdrop")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop).InsertOne(context.Background(), air)
	if err != nil {
		log.Error("AddAirdrop InsertOne err: ", err)
		return errors.New("add airdrop fail")
	}
	return nil
}
func GetAirdrop(tx string) (Airdrop, error) {
	if MonCli == nil {
		return Airdrop{}, errors.New("mongo client is nil" + "GetAirdrop")
	}
	filter := bson.D{{"tx_hash", tx}}
	var a Airdrop
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop).FindOne(context.Background(), filter).Decode(&a)
	if err != nil {
		log.Error("GetAirdrop FindOne err: ", err)
		return Airdrop{}, ErrNoDocuments
	}
	return a, nil
}
func QueryAirdrop(addr, today string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "QueryAirdropByTimes")
	}
	filter := bson.D{{"to_addr", addr}, {"airdrop_time", today}}
	var a Airdrop
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop).FindOne(context.Background(), filter).Decode(&a)
	if err != nil {
		log.Error("QueryAirdrop FindOne err: ", err)
		return ErrNoDocuments
	}
	return nil
}
func GetAirdropForAll(addr string, page, size int64) ([]Airdrop, error) {
	if MonCli == nil {
		return nil, errors.New("error:mongo.Client is nil" + "GetAirdropForAll")
	}
	// 构建查询条件
	filter := bson.M{}
	if addr != "" {
		filter["to_addr"] = strings.ToLower(addr)
	}

	// 分页参数处理
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	opts := options.Find().
		SetSort(bson.D{{"airdrop_time", -1}}). // 按创建时间倒序
		SetSkip((page - 1) * size).            // 跳过前面 (page-1)*size 条
		SetLimit(size)                         // 限制返回 size 条

	// 查询
	cursor, err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(airdrop).Find(context.Background(), filter, opts)
	if err != nil {
		log.Error("GetAirdropForAll Find err: ", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var res []Airdrop
	if err := cursor.All(context.Background(), &res); err != nil {
		log.Error("GetAirdropForAll Cursor  err: ", err)
		return nil, err
	}
	return res, nil
}
func UpdateAirdropStatus(tx string, i uint64) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateAirdrop")
	}
	filter := bson.M{"tx_hash": tx}
	update := bson.D{
		{"$set", bson.D{
			{"status", i},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateAirdropStatus err: ", err)
		return errors.New("UpdateAirdropStatus fail")
	}
	return nil
}
func UpdateAirdropHash(Id, tx string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateAirdrop")
	}
	filter := bson.M{"order_id": Id}
	update := bson.D{
		{"$set", bson.D{
			{"tx_hash", tx},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateAirdrop err: ", err)
		return errors.New("update airdrop fail")
	}
	return nil
}

// 每日发放空投流水
func AddDailyAirdrop(air DailyAirdropTrade) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "AddDailyAirdrop")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops).InsertOne(context.Background(), air)
	if err != nil {
		log.Error("AddDailyAirdrop InsertOne err: ", err)
		return errors.New("add daily airdrop fail")
	}
	return nil
}
func GetDailyAirdrop(timestamp, symbol string) (DailyAirdropTrade, error) {
	if MonCli == nil {
		return DailyAirdropTrade{}, errors.New("mongo client is nil" + "GetDailyAirdrop")
	}
	filter := bson.D{{"date", timestamp}, {"symbol", symbol}}
	var a DailyAirdropTrade
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops).FindOne(context.Background(), filter).Decode(&a)
	if err != nil {
		log.Error("GetDailyAirdrop FindOne err: ", err)
		return DailyAirdropTrade{}, ErrNoDocuments
	}
	return a, nil
}
func GetDailyAirdropBySymbol(symbol string) (DailyAirdropTrade, error) {
	if MonCli == nil {
		return DailyAirdropTrade{}, errors.New("mongo client is nil" + "GetDailyAirdropBySymbol")
	}
	filter := bson.D{{"symbol", symbol}}
	var a DailyAirdropTrade
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops).FindOne(context.Background(), filter).Decode(&a)
	if err != nil {
		log.Error("GetDailyAirdropBySymbol FindOne err: ", err)
		return DailyAirdropTrade{}, ErrNoDocuments
	}
	return a, nil
}
func UpdateDailyAirdrop(date, symbol, value string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateDailyAirdrop")
	}
	filter := bson.D{{"date", date}, {"symbol", symbol}}
	update := bson.D{
		{"$set", bson.D{
			{"reward", value},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateDailyAirdrop err: ", err)
		return errors.New("update airdrop fail")
	}
	return nil
}
func UpdateDailyAirdropRemain(val, symbol, date, total string) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "UpdateDailyAirdropRemain")
	}
	filter := bson.D{{"symbol", symbol}}
	update := bson.D{
		{"$set", bson.D{
			{"reward", val},
			{"date", date},
			{"pool_total", total},
		}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateDailyAirdropStatus err: ", err)
		return errors.New("update airdrop fail")
	}
	return nil
}

// 交易详情
func AddTransaction(tx Transaction) error {
	if MonCli == nil {
		return errors.New("mongo client is nil" + "AddTransaction")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(transaction).InsertOne(context.Background(), tx)
	if err != nil {
		log.Error("AddTransaction InsertOne err: ", err)
		return errors.New("add transaction fail")
	}
	return nil
}
func GetTransaction(tx string) (Transaction, error) {
	if MonCli == nil {
		return Transaction{}, errors.New("mongo client is nil" + "GetTransaction")
	}
	filter := bson.D{{"tx_hash", tx}}
	var txh Transaction
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(transaction).FindOne(context.Background(), filter).Decode(&txh)
	if err != nil {
		log.Error("GetTransaction FindOne err: ", err)
		return Transaction{}, ErrNoDocuments
	}
	return txh, nil
}
func QueryTransaction(tx, types string) (Transaction, error) {
	if MonCli == nil {
		return Transaction{}, errors.New("mongo client is nil" + "QueryTransaction")
	}
	filter := bson.D{{"tx_hash", tx}, {"transaction_type", types}}
	var txh Transaction
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(transaction).FindOne(context.Background(), filter).Decode(&txh)
	if err != nil {
		log.Error("FindOne err: ", err)
		return Transaction{}, ErrNoDocuments
	}
	return txh, nil
}

// 扫块内容
func AddScanBlock(a ScanBlock) error {
	//判断全局变量MonCli是否为空，
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddScanBlock ")
	}

	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(scanBlock).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddScanBlock InsertOne err: ", err)
		return err
	}
	return nil
}
func UpdateScanBlock(sbl ScanBlock) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil")
	}
	filter := bson.D{{"netWork", sbl.NetWork}}
	update := bson.D{{"$set", bson.D{
		{"latestBlock", sbl.LatestBlock},
		{"time", sbl.Time},
	},
	}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(scanBlock).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateScanBlock fail ", err)
		return err
	}

	return nil
}
func GetScanBlock(i uint64) (uint64, error) {
	if MonCli == nil {
		return 0, errors.New("mongo client is nil " + "GetBlock")
	}
	filter := bson.D{{"netWork", i}}
	var bl ScanBlock
	err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(scanBlock).FindOne(context.Background(), filter).Decode(&bl)
	if err != nil {
		log.Error("GetScanBlock FindOne err: ", err)
		return 0, ErrNoDocuments
	}
	return bl.LatestBlock, nil
}

// 错误块
func AddLossBlock(a LossBlock) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddLossBlock")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossBlock).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("InsertOne err: ", err)
		return err
	}
	return nil
}
func GetLossBlock(blockNr, chain uint64) (LossBlock, error) {
	if MonCli == nil {
		return LossBlock{}, errors.New("mongo client is nil " + "GetLossBlock")
	}
	filter := bson.D{{"blockNr", blockNr}, {"netWork", chain}}
	var bl LossBlock
	err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(lossBlock).FindOne(context.Background(), filter).Decode(&bl)
	if err != nil {
		log.Error("FindOne err: ", err)
		return LossBlock{}, ErrNoDocuments
	}
	return bl, nil
}
func GetLossBlocksByNetwork(networkID uint64) ([]LossBlock, error) {
	if MonCli == nil {
		return nil, errors.New("mongo client is nil in GetLossBlocksByNetwork")
	}

	// 创建查询过滤器：查找指定网络的所有区块
	filter := bson.D{{"netWork", networkID}}

	// 执行查询
	cursor, err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(lossBlock).
		Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query loss blocks: %v", err)
	}
	defer cursor.Close(context.Background())

	// 准备结果切片
	var blocks []LossBlock

	// 遍历游标，解码所有匹配的文档
	for cursor.Next(context.Background()) {
		var block LossBlock
		if err := cursor.Decode(&block); err != nil {
			return nil, fmt.Errorf("failed to decode loss block: %v", err)
		}
		blocks = append(blocks, block)
	}

	// 检查游标遍历过程中是否有错误
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return blocks, nil
}
func DeleteLossBlock(i uint64) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "DeleteLossBlock")
	}
	filter := bson.D{{"blockNr", i}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(lossBlock).DeleteOne(context.Background(), filter)
	if err != nil {
		log.Error("FindOne err: ", err)
		return err
	}
	return nil
}

func AddBztDapp(a BztDapp) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AffBztDapp")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(bztDapp).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("AddBztDapp InsertOne err: ", err)
		return err
	}
	return nil
}
func GetBztDapp(name string) (BztDapp, error) {
	if MonCli == nil {
		return BztDapp{}, nil
	}
	filter := bson.D{{"dapp_name", name}}
	var b BztDapp
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(bztDapp).FindOne(context.Background(), filter).Decode(&b)
	if err != nil {
		log.Error("GetBztDapp FindOne err: ", err)
		return b, ErrNoDocuments
	}
	return b, nil
}
func UpdateBztDapp(url, name string) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "UpdateBztDapp")
	}
	filter := bson.D{{"dapp_name", name}}
	update := bson.D{{"$set", bson.D{{"dapp_icon", url}}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(bztDapp).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateBztDapp err: ", err)
		return err
	}
	return nil
}

func AddDeployTransaction(tx DeployTransaction) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddDeployTransaction")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(deployContract).InsertOne(context.Background(), tx)
	if err != nil {
		log.Error("AddDeployTransaction InsertOne err: ", err)
		return err
	}
	return nil
}
func GetDeployTransaction(tx string) (DeployTransaction, error) {
	if MonCli == nil {
		return DeployTransaction{}, errors.New("error:mongo.Client is nil" + "GetDeployTransaction")
	}
	filter := bson.D{{"tx_hash", tx}}
	var txh DeployTransaction
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(deployContract).FindOne(context.Background(), filter).Decode(&txh)
	if err != nil {
		log.Error("FindOne err: ", err)
		return DeployTransaction{}, ErrNoDocuments
	}
	return txh, nil
}

func AddOrderSwitch(i OrderSwitch) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddOrderSwitch")
	}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(orderSwitch).InsertOne(context.Background(), i)
	if err != nil {
		log.Error("AddOrderSwitch InsertOne err: ", err)
		return err
	}
	return nil
}
func GetOrderSwitch(i uint64, types string) (OrderSwitch, error) {
	if MonCli == nil {
		return OrderSwitch{}, errors.New("error:mongo.Client is nil" + "GetOrderSwitch")
	}
	filter := bson.D{{"chain_id", i}, {"types", types}}
	var o OrderSwitch
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(orderSwitch).FindOne(context.Background(), filter).Decode(&o)
	if err != nil {
		log.Error("GetOrderSwitch FindOne err: ", err)
		return o, ErrNoDocuments
	}
	return o, nil
}

func AddKLineData(res []Kline, symbol, dataType string) error {
	var CollectionName string
	switch dataType {
	case "1h":
		CollectionName = kLineByOneHour //3天一个阶段
	case "4h":

		CollectionName = kLineByFourHour
	case "1d":
		CollectionName = kLineByOneDay
	case "3d":
		CollectionName = kLineByThreeDay
		//3天一个阶段
	default:
		return errors.New("invalid interval")
	}
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(CollectionName)

	for _, kline := range res {
		doc := Kline{
			OpenTime:                 kline.OpenTime,
			OpenPrice:                kline.OpenPrice,
			HighPrice:                kline.HighPrice,
			LowPrice:                 kline.LowPrice,
			ClosePrice:               kline.ClosePrice,
			Volume:                   kline.Volume,
			CloseTime:                kline.CloseTime,
			QuoteAssetVolume:         kline.QuoteAssetVolume,
			NumberOfTrades:           kline.NumberOfTrades,
			TakerBuyBaseAssetVolume:  kline.TakerBuyBaseAssetVolume,
			TakerBuyQuoteAssetVolume: kline.TakerBuyQuoteAssetVolume,
			Ignore:                   kline.Ignore,
			DataType:                 dataType,
			Symbol:                   symbol,
		}

		_, err := coll.InsertOne(context.Background(), doc)
		if err != nil {
			// 检查是否是 duplicate key error
			if isDuplicateKeyError(err) {
				// 跳过这一条
				log.Infof("跳过重复的 KLine: symbol=%s dataType=%s closeTime=%v", symbol, dataType, kline.CloseTime)
				continue
			}
			// 如果是别的错误，就返回
			log.Errorf("插入失败: %v", err)
			return err
		}
		log.Infof("addKLineData success :%s--%d", symbol, kline.CloseTime)
	}

	return nil
}

func AddKLineDataMany(docs []interface{}, collectionName string) error {
	if MonCli == nil {
		return errors.New("mongo.Client is nil: AddKLineData")
	}

	if len(docs) == 0 {
		return nil // 没有数据直接返回
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := MonCli.Client.Database(DatabaseNameForChain).Collection(collectionName)

	// 批量插入
	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		log.Errorf("AddKLineData InsertMany error: %v", err)
		return err
	}

	return nil
}
func GetKLineData(types, symbol, CollectionName string, page, size int64) ([]Kline, error) {
	if MonCli == nil {
		return []Kline{}, errors.New("error:mongo.Client is nil ：GetKLineData")
	}
	filter := bson.M{
		"data_type": types,
		"symbol":    symbol,
	}
	// 分页参数处理
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}

	opts := options.Find().
		SetSort(bson.D{{"close_time", -1}}). // 按创建时间倒序
		SetSkip((page - 1) * size).          // 跳过前面 (page-1)*size 条
		SetLimit(size)                       // 限制返回 size 条

	// 查询
	cursor, err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(CollectionName).Find(context.Background(), filter, opts)
	if err != nil {
		log.Error("GetKLineData Find err: ", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var res []Kline
	if err := cursor.All(context.Background(), &res); err != nil {
		log.Error("GetKLineData Cursor  err: ", err)
		return nil, err
	}
	return res, nil
}

// EnsureKlineIndexes TODO 编写接口创建索引，提升效率
func EnsureKlineIndexes() error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil: EnsureKlineIndexes")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := KlineOneDayIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = KlineThreeDayIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = KlineOneHourIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = KlineFourHourIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = TxHashIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = AirdropOneIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = UserAmountIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = OrderIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = OrderAddrIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = PriceIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = PriceAndTimeIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	err = UserIndex(ctx)
	if err != nil {
		log.Error("EnsureKlineIndexes err: ", err)
		return err
	}
	return nil
}
func DropIndexIfExists(collectionName, indexName string) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil: DropIndexIfExists")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取集合引用
	collection := MonCli.Client.Database(DatabaseNameForChain).Collection(collectionName)

	// 获取所有索引
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("获取索引列表失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 检查目标索引是否存在
	var indexFound bool
	for cursor.Next(ctx) {
		var index map[string]interface{}
		if err := cursor.Decode(&index); err != nil {
			return fmt.Errorf("解析索引信息失败: %v", err)
		}

		if name, ok := index["name"].(string); ok && name == indexName {
			indexFound = true
			break
		}
	}

	// 如果索引存在，则删除
	if indexFound {
		result, err := collection.Indexes().DropOne(ctx, indexName)
		if err != nil {
			return fmt.Errorf("删除索引失败: %v", err)
		}
		log.Infof("成功删除索引: %s, 结果: %s", indexName, result)
	} else {
		log.Infof("索引 '%s' 不存在，无需删除", indexName)
	}

	return nil
}

func KlineOneDayIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(kLineByOneDay)

	// 定义需要的索引
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"symbol", 1},
				{"data_type", 1},
				{"close_time", -1},
			},
			Options: options.Index().SetUnique(true).SetName("symbol_dataType_closeTime_idx"),
		},
	}
	// 创建索引（已存在的不会重复建）
	_, err := coll.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Errorf("创建 symbol_dataType_closeTime_idx 索引失败: %s.%s", DatabaseNameForChain, kLineByOneDay)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, kLineByOneDay)
	return nil
}
func KlineThreeDayIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(kLineByThreeDay)

	// 定义需要的索引
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"symbol", 1},
				{"data_type", 1},
				{"close_time", -1},
			},
			Options: options.Index().SetUnique(true).SetName("symbol_dataType_closeTime_idx"),
		},
	}
	// 创建索引（已存在的不会重复建）
	_, err := coll.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Errorf("创建 symbol_dataType_closeTime_idx 索引失败: %s.%s", DatabaseNameForChain, kLineByThreeDay)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, kLineByThreeDay)
	return nil
}
func KlineOneHourIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(kLineByOneHour)

	// 定义需要的索引
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"symbol", 1},
				{"data_type", 1},
				{"close_time", -1},
			},
			Options: options.Index().SetUnique(true).SetName("symbol_dataType_closeTime_idx"),
		},
	}
	// 创建索引（已存在的不会重复建）
	_, err := coll.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Errorf("创建 symbol_dataType_closeTime_idx 索引失败: %s.%s", DatabaseNameForChain, kLineByOneHour)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, kLineByOneHour)
	return nil
}
func KlineFourHourIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(kLineByFourHour)

	// 定义需要的索引
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"symbol", 1},
				{"data_type", 1},
				{"close_time", -1},
			},
			Options: options.Index().SetUnique(true).SetName("symbol_dataType_closeTime_idx"),
		},
	}
	// 创建索引（已存在的不会重复建）
	_, err := coll.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Errorf("创建 symbol_dataType_closeTime_idx 索引失败: %s.%s", DatabaseNameForChain, kLineByFourHour)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, kLineByFourHour)
	return nil
}

func TxHashIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(transaction)

	// 在交易记录表创建索引
	index := mongo.IndexModel{
		Keys:    bson.D{{"tx_hash", 1}}, // 升序索引
		Options: options.Index().SetUnique(true).SetName("txHash_idx"),
	}

	// 创建索引
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 txHash_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, transaction)
	return nil
}
func AirdropOneIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(airdrop)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"to_addr", 1},
		},
		Options: options.Index().SetName("toAddr_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 toAddr_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, airdrop)
	return nil
}
func UserAmountIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(lossAmount)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"user_addr", 1},
		},
		Options: options.Index().SetName("userAddr_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 userAddr_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, lossAmount)
	return nil
}
func OrderIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(order)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"order_id", 1},
		},
		Options: options.Index().SetUnique(true).SetName("orderId_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 orderId_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, order)
	return nil
}
func OrderAddrIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(order)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"users_addr", 1},
		},
		Options: options.Index().SetName("usersAddr_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 usersAddr_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, order)
	return nil
}
func PriceIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"symbol", 1},
			{"index", 1},
		},
		Options: options.Index().SetUnique(true).SetName("symbol_index_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 symbol_index_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, newPrice)
	return nil
}
func PriceAndTimeIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"symbol", 1},
			{"timestamp", -1},
		},
		Options: options.Index().SetName("symbol_timestamp_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 symbol_timestamp_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, newPrice)
	return nil
}
func UserIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(user)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"address", 1},
		},
		Options: options.Index().SetUnique(true).SetName("address_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 address_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 索引已确保存在: %s.%s", DatabaseNameForChain, user)
	return nil
}

func DailyAirdropIndex(ctx context.Context) error {
	coll := MonCli.Client.Database(DatabaseNameForChain).Collection(dailyAirdrops)
	index := mongo.IndexModel{
		Keys: bson.D{
			{"symbol", 1},
		},
		Options: options.Index().SetName("symbol_idx"),
	}
	_, err := coll.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Warnf("创建 symbol_idx 索引失败: %v", err)
		return err
	}
	log.Infof("✅ 单个索引已确保存在: %s.%s", DatabaseNameForChain, dailyAirdrops)
	return nil
}

func isDuplicateKeyError(err error) bool {
	var writeExc mongo.WriteException
	if errors.As(err, &writeExc) {
		for _, we := range writeExc.WriteErrors {
			if we.Code == 11000 { // E11000 是 duplicate key 的 code
				return true
			}
		}
	}
	return false
}
