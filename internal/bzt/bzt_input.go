package bzt

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"valueguard/internal/conf"
	"valueguard/internal/erc20"
)

func GetOpenOrderInput(_orderId *big.Int, _tokenName string, _amount *big.Int) ([]byte, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		log.Warnf("GetOpenOrderInput: abi is nil")
		return nil, errors.New("GetABI returned nil")
	}
	input, err := parsed.Pack("openOrder", _orderId, _tokenName, _amount)
	if err != nil {
		log.Warnf("GetOpenOrderInput: pack openOrder: %v", err)
		return nil, err
	}
	return input, nil
}
func GetCloseOrderInput(orderId, openPrice, closePrice *big.Int, tokenName string) ([]byte, error) {

	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		log.Warnf("GetCloseOrderInput: abi is nil")
		return nil, err
	}
	if parsed == nil {
		log.Warnf("GetCloseOrderInput  parsed : abi is nil")
		return nil, errors.New("GetABI returned nil")
	}

	input, err := parsed.Pack("closeOrder", orderId, openPrice, closePrice, tokenName)
	if err != nil {
		log.Warnf("GetCloseOrderInput: pack closeOrder: %v", err)
		return nil, err
	}
	return input, nil
}
func GetAirdropInput(_spender common.Address, _value *big.Int) ([]byte, error) {
	parsed, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		log.Warnf("GetAirdropInput: abi is nil")
		return nil, err
	}
	if parsed == nil {
		log.Warnf("GetAirdropInput  parsed : abi is nil")
		return nil, errors.New("GetABI returned nil")
	}

	input, err := parsed.Pack("approve", _spender, _value)
	if err != nil {
		log.Warnf("GetAirdropInput: pack airdrop: %v", err)
		return nil, err
	}
	return input, nil
}
func DeployContract(data []byte, cli *ethclient.Client) (string, string, error) {
	apiURL := conf.BaseUrl + "/api/Sign2"
	fromAddress := common.HexToAddress(conf.OwnerAddress)
	nonce, err := cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", "", fmt.Errorf("获取 nonce 失败: %v", err)
	}
	log.Info("nonce:", nonce)

	// 设置 gasPrice 和 gasLimit
	gasPrice := new(big.Int)
	gasPrice.SetString("300000000000", 10) // 200 gwei
	gasPriceStr := gasPrice.String()
	gasLimit := uint64(3000000)

	// 构建交易
	input := "0x" + hex.EncodeToString(data)
	valueZero := big.NewInt(0)
	hexString := "0x" + valueZero.Text(16)

	// 构建请求数据
	requestBody := map[string]interface{}{
		//	"to":       toContract,
		"key":      conf.KeyId,
		"value":    hexString,
		"nonce":    nonce,
		"gaslimit": gasLimit,
		"gasprice": gasPriceStr,
		"input":    input,
		"rpcurl":   conf.RpcUrl,
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("JSON 编码失败: %v", err)
	}

	// HMAC-SHA256 签名
	h := hmac.New(sha256.New, []byte(conf.HmacKey))
	h.Write(jsonBytes)
	signature := hex.EncodeToString(h.Sum(nil))

	// 构造 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", "", fmt.Errorf("请求构建失败: %v", err)
	}

	millis := time.Now().UnixMilli()           // 获取当前时间的毫秒时间戳 (int64)
	strMillis := strconv.FormatInt(millis, 10) // 转换为字符串
	fmt.Println("strMillis:", strMillis)

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", conf.Apikey)
	req.Header.Set("hmac", signature)
	req.Header.Set("timestamp", strMillis)

	// 发送请求
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取原始响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印原始响应内容（调试用）
	fmt.Println("原始响应:", string(bodyBytes))

	// 尝试解析为 JSON 字符串（如 "0x123..."）
	var resultStr string
	if err := json.Unmarshal(bodyBytes, &resultStr); err == nil {
		return resultStr, strconv.FormatUint(nonce, 10), nil
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &resultMap); err == nil {
		resultJSON, _ := json.MarshalIndent(resultMap, "", "  ")
		return string(resultJSON), strconv.FormatUint(nonce, 10), nil
	}

	// 如果两者都不是，返回原始内容
	return "", "", fmt.Errorf("无法解析响应数据: %s", string(bodyBytes))
}
func DeployContractTransfer(data []byte, cli *ethclient.Client) (*types.Transaction, string, error) {
	// 示例调用
	results, nonce, err := DeployContract(data, cli)
	if err != nil {
		log.Warnf("发送交易失败:%v", err)
		return nil, "", err
	}
	jsonData := []byte(results)
	type Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			R                 string `json:"r"`
			S                 string `json:"s"`
			SignedTransaction string `json:"signedtransaction"`
			V                 string `json:"v"`
		} `json:"data"`
	}
	// 创建结构体变量
	var resp Response
	// 解析 JSON
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return nil, "", err
	}

	// 打印解析后的内容
	//fmt.Println("Signed Transaction:", resp.Data.SignedTransaction)
	if resp.Code == 200 {
		signTx := resp.Data.SignedTransaction
		// 广播
		// 解码为字节数组
		rawTxBytes, err := hex.DecodeString(signTx)
		if err != nil {
			log.Warnf("Failed to decode raw tx: %v", err)
		}
		// 解码为 types.Transaction 对象（可选）
		tx := new(types.Transaction)
		err = tx.UnmarshalBinary(rawTxBytes)
		if err != nil {
			log.Warnf("Failed to unmarshal tx: %v", err)
			return nil, "", err
		}
		// 广播交易
		err = cli.SendTransaction(context.Background(), tx)
		if err != nil {
			log.Warnf("Failed to send transaction: %v", err)
			return nil, "", err
		}
		//fmt.Printf("Transaction broadcasted! Hash: %s\n", tx.Hash().Hex())
		return tx, nonce, nil
	}

	fmt.Println("JSON解析出错:", err)
	return nil, "", fmt.Errorf("解析失败: %v", resp)
}
func UrlContractSignOwner(data []byte, cli *ethclient.Client) (string, string, error) {
	apiURL := conf.BaseUrl + "/api/Sign2"
	fromAddress := common.HexToAddress(conf.OwnerAddress)
	nonce, err := cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", "", fmt.Errorf("获取 nonce 失败: %v", err)
	}
	log.Info("nonce:", nonce)
	toContract := common.HexToAddress(conf.ContractBztAddr)

	// 设置 gasPrice 和 gasLimit
	gasPrice := new(big.Int)
	gasPrice.SetString("300000000000", 10) // 200 gwei
	gasPriceStr := gasPrice.String()
	gasLimit := uint64(600000)

	// 构建交易
	input := "0x" + hex.EncodeToString(data)
	valueZero := big.NewInt(0)
	hexString := "0x" + valueZero.Text(16)

	// 构建请求数据
	requestBody := map[string]interface{}{
		"to":       toContract,
		"key":      conf.KeyId,
		"value":    hexString,
		"nonce":    nonce,
		"gaslimit": gasLimit,
		"gasprice": gasPriceStr,
		"input":    input,
		"rpcurl":   conf.RpcUrl,
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("JSON 编码失败: %v", err)
	}

	// HMAC-SHA256 签名
	h := hmac.New(sha256.New, []byte(conf.HmacKey))
	h.Write(jsonBytes)
	signature := hex.EncodeToString(h.Sum(nil))

	// 构造 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", "", fmt.Errorf("请求构建失败: %v", err)
	}

	millis := time.Now().UnixMilli()           // 获取当前时间的毫秒时间戳 (int64)
	strMillis := strconv.FormatInt(millis, 10) // 转换为字符串
	fmt.Println("strMillis:", strMillis)

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", conf.Apikey)
	req.Header.Set("hmac", signature)
	req.Header.Set("timestamp", strMillis)

	// 发送请求
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取原始响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印原始响应内容（调试用）
	fmt.Println("原始响应:", string(bodyBytes))

	// 尝试解析为 JSON 字符串（如 "0x123..."）
	var resultStr string
	if err := json.Unmarshal(bodyBytes, &resultStr); err == nil {
		return resultStr, strconv.FormatUint(nonce, 10), nil
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &resultMap); err == nil {
		resultJSON, _ := json.MarshalIndent(resultMap, "", "  ")
		return string(resultJSON), strconv.FormatUint(nonce, 10), nil
	}

	// 如果两者都不是，返回原始内容
	return "", "", fmt.Errorf("无法解析响应数据: %s", string(bodyBytes))
}
func UrlOwnerContractTransfer(data []byte, cli *ethclient.Client) (string, string, error) {
	time.Sleep(1 * time.Second / 2)
	// 示例调用
	results, nonce, err := UrlContractSignOwner(data, cli)
	if err != nil {
		log.Warnf("发送交易失败:%v", err)
		return "", "", err
	}
	jsonData := []byte(results)
	type Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			R                 string `json:"r"`
			S                 string `json:"s"`
			SignedTransaction string `json:"signedtransaction"`
			V                 string `json:"v"`
		} `json:"data"`
	}
	// 创建结构体变量
	var resp Response
	// 解析 JSON
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return "", "", err
	}

	// 打印解析后的内容
	fmt.Println("Signed Transaction:", resp.Data.SignedTransaction)
	if resp.Code == 200 {
		signTx := resp.Data.SignedTransaction
		// 广播
		// 解码为字节数组
		rawTxBytes, err := hex.DecodeString(signTx)
		if err != nil {
			log.Warnf("Failed to decode raw tx: %v", err)
		}
		// 解码为 types.Transaction 对象（可选）
		tx := new(types.Transaction)
		err = tx.UnmarshalBinary(rawTxBytes)
		if err != nil {
			log.Warnf("Failed to unmarshal tx: %v", err)
			return "", "", err
		}
		// 广播交易
		err = cli.SendTransaction(context.Background(), tx)
		if err != nil {
			log.Warnf("Failed to send transaction: %v", err)
			return "", "", err
		}
		//fmt.Printf("Transaction broadcasted! Hash: %s\n", tx.Hash().Hex())
		return strings.ToLower(tx.Hash().Hex()), nonce, nil
	}

	fmt.Println("JSON解析出错:", err)
	return "", "", fmt.Errorf("解析失败: %v", resp)
}
func InitEthClient(rpcURL string) (*ethclient.Client, error) {
	// 固定 API Key
	x_api_key := "4sip97qapC4vTxS73YdTB6X5hm8Rr8Uk13BdwP2d123"

	// 自定义 HTTP 客户端，注入 x-api-key Header
	httpClient := &http.Client{
		Transport: &transportWithHeader{
			headers: map[string]string{
				"x-api-key": x_api_key,
			},
			base: http.DefaultTransport,
		},
	}

	// 使用自定义 httpClient 初始化 RPC 客户端
	rpcClient, err := rpc.DialHTTPWithClient(rpcURL, httpClient)
	if err != nil {
		log.Error("RPC dial error:", err)
		return nil, err
	}

	// 使用 rpcClient 初始化 ethclient
	return ethclient.NewClient(rpcClient), nil
}

// transportWithHeader 用于注入 header
type transportWithHeader struct {
	headers map[string]string
	base    http.RoundTripper
}

func (t *transportWithHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

func UrlGetKeyAddress() (string, error) {
	apiURL := conf.BaseUrl + "/api/GetAddress?keyid=" + conf.KeyId
	// 构造 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("请求构建失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", conf.Apikey)

	// 发送请求
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取原始响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印原始响应内容（调试用）
	fmt.Println("原始响应:", string(bodyBytes))

	// 尝试解析为 JSON 字符串（如 "0x123..."）
	var resultStr string
	if err := json.Unmarshal(bodyBytes, &resultStr); err == nil {
		return resultStr, nil
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &resultMap); err == nil {
		resultJSON, _ := json.MarshalIndent(resultMap, "", "  ")
		return string(resultJSON), nil
	}

	// 如果两者都不是，返回原始内容
	return "", fmt.Errorf("无法解析响应数据: %s", string(bodyBytes))
}
