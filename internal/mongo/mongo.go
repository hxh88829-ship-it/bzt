package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		log.Error("InsertOne err: ", err)
		return errors.New("add user fail")
	}
	return nil
}
func GetUser(addr string) (Users, error) {
	if MonCli == nil {
		return Users{}, errors.New("mongo client is nil " + "GetUser")
	}
	filter := bson.M{
		"$or": []bson.M{
			{"address": addr},
			{"uid": addr},
		},
	}
	var ma Users
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		return Users{}, ErrNoDocuments
	}
	return ma, nil
}
func JudgeUser(from, to string) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "JudgeUser")
	}
	filter := bson.M{
		"$or": []bson.M{
			{"address": from},
			{"address": to},
		},
	}
	var ma Users
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		log.Error("FindOne err: ", err)
		return errors.New("find user fail")
	}
	return nil
}
func UpdateUser(uid, OriginalMessage string) error {
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil")
	}
	filter := bson.D{{"uid", uid}}
	update := bson.D{{"$set", bson.D{
		{"original_message", OriginalMessage},
	},
	}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).UpdateMany(context.Background(), filter, update)
	if err != nil {
		log.Error(err, "UpdateUser fail")
		return err
	}
	return nil
}
func UpdateLogin(addr, name, email, twitter, avatar, website, Description string) error {
	if MonCli == nil {
		return errors.New("error: mongo.Client is nil (UpdateLogin)")
	}
	if addr == "" {
		return errors.New("error: address is empty")
	}
	if len(addr) != 42 {
		return errors.New("error: address is invalid")
	}
	filter := bson.D{{"address", addr}}

	// 动态构造更新字段
	setFields := bson.D{}
	if name != "" {
		setFields = append(setFields, bson.E{"name", name})
	}
	if email != "" {
		setFields = append(setFields, bson.E{"email", email})
	}
	if twitter != "" {
		setFields = append(setFields, bson.E{"twitter", twitter})
	}
	if avatar != "" {
		setFields = append(setFields, bson.E{"avatar", avatar})
	}
	if website != "" {
		setFields = append(setFields, bson.E{"website", website})
	}
	if Description != "" {
		setFields = append(setFields, bson.E{"description", Description})
	}
	// 如果没有要更新的字段，直接返回
	if len(setFields) == 0 {
		return ErrNoFields
	}

	update := bson.D{{"$set", setFields}}

	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(user).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err, "UpdateLogin failed")
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
		log.Error("InsertOne err: ", err)
		return errors.New("add price fail")
	}
	return nil
}
func GetPriceForIndex(symbol string, ind int) (CoinPrice, error) {
	if MonCli == nil {
		return CoinPrice{}, errors.New("error:mongo.Client is nil" + "GetPrice")
	}
	filter := bson.D{{"symbol", symbol}, {"index", ind}}
	var ma CoinPrice
	err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).FindOne(context.Background(), filter).Decode(&ma)
	if err != nil {
		log.Error("FindOne err: ", err)
		return CoinPrice{}, ErrNoDocuments
	}
	return ma, nil
}

func SavePrice(symbol, price string, ind int) error {
	if MonCli == nil {
		return errors.New("mongo client is nil")
	}
	filter := bson.D{{"symbol", symbol}, {"index", ind}}
	update := bson.D{{"$set", bson.D{
		{"price", price},
	}}}
	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(newPrice).UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error("UpdateLogin err: ", err)
		return errors.New("save price fail")
	}
	return nil
}

// 扫块内容
func AddScanBlock(a ScanBlock) error {
	//判断全局变量MonCli是否为空，
	if MonCli == nil {
		return errors.New("error:mongo.Client is nil" + "AddScanBlock ")
	}

	_, err := MonCli.Client.Database(DatabaseNameForChain).Collection(scanBlock).InsertOne(context.Background(), a)
	if err != nil {
		log.Error("InsertOne err: ", err)
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
		log.Error(err, "UpdateUser fail")
		return err
	}

	return nil
}
func GetScanBlock(i uint64) (ScanBlock, error) {
	if MonCli == nil {
		return ScanBlock{}, errors.New("mongo client is nil " + "GetBlock")
	}
	filter := bson.D{{"netWork", i}}
	var bl ScanBlock
	err := MonCli.Client.Database(DatabaseNameForChain).
		Collection(scanBlock).FindOne(context.Background(), filter).Decode(&bl)
	if err != nil {
		log.Error("FindOne err: ", err)
		return ScanBlock{}, ErrNoDocuments
	}
	return bl, nil
}
