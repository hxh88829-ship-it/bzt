package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"math"
	"math/big"
	"strings"
	"time"
	"valueguard/internal/erc20"
)

var Client *ethclient.Client
var ChainId uint64

const addrContract = "0xaD6780B2A022B79686c5E56017cC4FB8cfCd9726"

// 普通交易
func GetBlockByNumber(num uint64) (*types.Block, error) {
	//TODO 睡眠0.05秒
	time.Sleep(time.Second / 500)
	return Client.BlockByNumber(context.Background(), big.NewInt(int64(num)))
}

func GetTransactionByHash(hash common.Hash) (*types.Transaction, bool, error) {
	return Client.TransactionByHash(context.Background(), hash)
}

func GetTransactionReceiptByHash(hash common.Hash) (*types.Receipt, error) {
	return Client.TransactionReceipt(context.Background(), hash)
}

func GetBalanceByAddress(addr string) (*big.Int, error) {
	return Client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
}

func GetBlockNumber() (uint64, error) {

	return Client.BlockNumber(context.Background())
}

func GetFromByTransaction(tx *types.Transaction) (common.Address, error) {
	signer := types.NewPragueSigner(new(big.Int).SetUint64(ChainId))
	from, err := signer.Sender(tx)
	if err != nil {
		log.Error("FromAdd", "err", err)
		return common.Address{}, err
	}
	return from, nil
}

// 返回合约地址的code
func GetCode(addr string) ([]byte, error) {
	return Client.CodeAt(context.Background(), common.HexToAddress(addr), nil)
}

func NewTransferOpt(key string, Code uint64) (*bind.TransactOpts, error) {
	pri, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}
	opt, err := bind.NewKeyedTransactorWithChainID(pri, new(big.Int).SetUint64(Code))
	if err != nil {
		return nil, err
	}
	return opt, nil
}

func SendTransaction(
	private string, //钱包私钥
	useradd string,
	cli *ethclient.Client, //节点client
	gasLimit uint64, //gas数量
	//data []byte,
	toAdd common.Address, //接受地址
	value *big.Int,
) (string, error) {

	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	//构造签名相关参数
	opts, err := NewTransferOpt(private, chainId.Uint64())
	if err != nil {
		return "", err
	}

	//获取单价
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	//获取nonce
	nonce, err := cli.PendingNonceAt(context.Background(), common.HexToAddress(useradd))
	if err != nil {
		return "", err
	}

	//va, err := valuePow(value)
	//if err != nil {
	//	return "", err
	//}

	//构造交易
	rawTx := types.NewTx(&types.LegacyTx{
		To:       &toAdd,
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Value:    value,
		//Data: data,
	})

	//sign transaction
	signedTx, err := opts.Signer(opts.From, rawTx)
	if err != nil {
		return "", err
	}

	//send transaction
	err = cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().String(), nil
}

// 合约交易
func Erc20Caller_Name() (string, error) {
	con := common.HexToAddress(addrContract)

	//只读合约接口创建
	ca, err := erc20.NewErc20Caller(con, Client)
	if err != nil {
		log.Error("Erc20Caller_Name 接口创建失败")
		return "", err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}

	na, err := ca.Name(&opt)
	if err != nil {
		log.Error("Erc20Caller_Name 名字查询失败")
		return "", err
	}

	return na, nil
}

func Erc20Transactor_BalanceOf(addr string) (string, error) {

	con := common.HexToAddress(addrContract)
	ca, err := erc20.NewErc20Caller(con, Client)
	if err != nil {
		log.Error("Erc20Transactor_BalanceOf  只读接口创建失败")
		return "", err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}

	uAddr := common.HexToAddress(addr)
	ba, err := ca.BalanceOf(&opt, uAddr)
	if err != nil {
		log.Error("Erc20Transactor_BalanceOf  余额查询失败")
		return "", err
	}
	return ba.String(), nil
}

func Erc20Transactor_Decimals() (uint8, error) {

	con := common.HexToAddress(addrContract)
	ca, err := erc20.NewErc20Caller(con, Client)
	if err != nil {
		log.Error("Erc20Transactor_Decimals 接口创建失败")
		return 0, err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	de, err := ca.Decimals(&opt)
	if err != nil {
		log.Error("Erc20Transactor_Decimals 查询失败")
		return 0, err
	}
	return de, nil
}

func Erc20Transactor_Symbol() (string, error) {

	con := common.HexToAddress(addrContract)
	ca, err := erc20.NewErc20Caller(con, Client)
	if err != nil {
		log.Error("Erc20Transactor_Symbol 接口创建失败")
		return "", err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	sa, err := ca.Symbol(&opt)
	if err != nil {
		log.Error("Erc20Transactor_Symbol 查询失败")
		return "", err
	}
	return sa, nil
}

func Erc20Transactor_TotalSupply() (uint64, error) {

	con := common.HexToAddress(addrContract)
	ca, err := erc20.NewErc20Caller(con, Client)
	if err != nil {
		log.Error("Erc20Transactor_TotalSupply 接口创建失败")
		return 0, err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	ts, err := ca.TotalSupply(&opt)
	if err != nil {
		log.Error("Erc20Transactor_TotalSupply 查询失败")
		return 0, err
	}
	return ts.Uint64(), nil
}

func Erc20Transactor_Transfer(key, add string, value *big.Int) (string, error) {
	//获取chain ID
	chainId, err := Client.ChainID(context.Background())
	if err != nil {
		log.Error("Erc20Transactor_Transfer ChainID  fail")
		return "", err
	}

	//构造签名结构体
	opts, err := NewTransferOpt(key, chainId.Uint64())
	if err != nil {
		log.Error("Erc20Transactor_Transfer 构造签名错误")
		return "", err
	}

	//构造合约client
	con := common.HexToAddress(addrContract)
	ca, err := erc20.NewErc20Transactor(con, Client)
	if err != nil {
		log.Error("Erc20Transactor_Transfer 合约构造失败")
		return "", err
	}

	//接受地址
	to := common.HexToAddress(add)

	tx, err := ca.Transfer(opts, to, value)
	if err != nil {
		log.Error("Erc20Transactor_Transfer 交易发送失败")
		return "", err
	}
	return tx.Hash().String(), nil
}

func ValuePow(i float64) (*big.Int, error) {
	if i <= 0 {
		value := new(big.Int).Mul(new(big.Int).SetUint64(0), new(big.Int).SetUint64(0))
		return value, nil
	}
	var Gwei float64
	Gwei = math.Pow(10, 9)
	value := i * Gwei
	va := new(big.Int).Mul(new(big.Int).SetUint64(1), new(big.Int).SetUint64(uint64(value)))
	return va, nil
}

func GetLogsEvent(tx string) ([]TransferEvent, error) {
	re, err := GetTransactionReceiptByHash(common.HexToHash(tx))
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(strings.ToLower(addrContract))
	eventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") //固定transfer解析

	eventAbi := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},
                {"indexed":true,"name":"to","type":"address"},
                {"indexed":false,"name":"value","type":"uint256"}],
                "name":"Transfer","type":"event"}]`

	parsedAbi, err := abi.JSON(strings.NewReader(eventAbi))
	if err != nil {
		return nil, fmt.Errorf("ABI parse failed: %w", err)
	}

	var transfers []TransferEvent

	for _, vLog := range re.Logs {
		// 过滤条件：合约地址 + 事件签名 + 参数数量
		if vLog.Address != contractAddress ||
			len(vLog.Topics) == 0 ||
			vLog.Topics[0] != eventSig ||
			len(vLog.Topics) < 3 {
			continue
		}

		var transfer TransferEvent
		transfer.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		transfer.To = common.BytesToAddress(vLog.Topics[2].Bytes())

		if err = parsedAbi.UnpackIntoInterface(&transfer, "Transfer", vLog.Data); err != nil {
			log.Error("ABI unpack failed: ", err)
			continue
		}

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
func GetLogsEventValue(log []TransferEvent) (TransferEvent, error) {
	var transfer TransferEvent
	for _, vLog := range log {
		transfer.Value = vLog.Value
		transfer.From = vLog.From
		transfer.To = vLog.To
	}
	return transfer, nil
}
func DirectLogValue(tx string) (TransferEvent, error) {
	transfer, err := GetLogsEvent(tx)
	if err != nil {
		log.Error("api   DirectLogValue GetLogsEvent fail")
		return TransferEvent{}, err
	}
	tr, err := GetLogsEventValue(transfer)
	if err != nil {
		log.Error("api   DirectLogValue  GetLogsEventValue fail")
		return TransferEvent{}, err
	}
	return tr, nil
}

// 获取十六进制字符串的最后 n 个字节
func GetLastNBytes(hexData string, n int) (string, error) {
	// 移除可能的 "0x" 前缀
	cleaned := strings.TrimPrefix(hexData, "0x")

	// 检查是否是有效的十六进制字符串
	if len(cleaned)%2 != 0 {
		return "", fmt.Errorf("invalid hex string length (must be even number of characters)")
	}

	// 计算实际可获取的字节数
	totalBytes := len(cleaned) / 2
	if totalBytes == 0 {
		return "0x", nil
	}

	// 确定实际要获取的字节数
	bytesToGet := n
	if bytesToGet > totalBytes {
		bytesToGet = totalBytes
	}

	// 计算起始位置（字符索引）
	startPos := len(cleaned) - bytesToGet*2
	lastBytes := cleaned[startPos:]

	return "0x" + lastBytes, nil
}

func StringToBigInt(s string) (*big.Int, error) {
	i, ok := new(big.Int).SetString(s, 0)
	if !ok {
		log.Error("StringToBigInt fail")
		return nil, errors.New("StringToBigInt fail")
	}
	return i, nil
}

func StringToBigIntSum(a, b string) (*big.Int, error) {
	i := new(big.Int)
	_, ok := i.SetString(a, 0) // 支持0x/0前缀的十六进制/八进制
	if !ok {
		return nil, errors.New("StringToBigInt fail: invalid format for first argument")
	}

	j := new(big.Int)
	_, ok = j.SetString(b, 0)
	if !ok {
		return nil, errors.New("StringToBigInt fail: invalid format for second argument")
	}

	sum := new(big.Int)
	sum.Add(i, j) // 大整数加法
	return sum, nil
}
func StringToBigIntSub(a, b string) (*big.Int, error) {
	i := new(big.Int)
	_, ok := i.SetString(a, 0) // 支持0x/0前缀的十六进制/八进制
	if !ok {
		return nil, errors.New("StringToBigInt fail: invalid format for first argument")
	}

	j := new(big.Int)
	_, ok = j.SetString(b, 0)
	if !ok {
		return nil, errors.New("StringToBigInt fail: invalid format for second argument")
	}

	sum := new(big.Int)
	sum.Sub(i, j) // 大整数加法
	return sum, nil
}

func GenerateUID() string {
	return uuid.New().String() // 默认使用 UUID v4（随机生成）
}
func GetJwtKey(uid string) (string, error) {
	claims := jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("123456"))
}
