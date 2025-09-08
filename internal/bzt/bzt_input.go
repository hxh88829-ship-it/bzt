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
)

const ProduceBztBin = "0x608060405234801561001057600080fd5b506040516112ab3803806112ab83398101604081905261002f916100d8565b338061005557604051631e4fbdf760e01b81526000600482015260240160405180910390fd5b61005e81610088565b5060018055600280546001600160a01b0319166001600160a01b0392909216919091179055610108565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156100ea57600080fd5b81516001600160a01b038116811461010157600080fd5b9392505050565b611194806101176000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c80638da5cb5b116100665780638da5cb5b146100e3578063a85c38ef14610108578063a98ad46c1461012f578063f2fde38b14610142578063fd53e1f21461015557600080fd5b806310199bdc146100985780636f9fb98a146100ad578063715018a6146100c85780638ba4cc3c146100d0575b600080fd5b6100ab6100a6366004610d0a565b610168565b005b6100b56105b0565b6040519081526020015b60405180910390f35b6100ab610622565b6100ab6100de366004610d80565b610636565b6000546001600160a01b03165b6040516001600160a01b0390911681526020016100bf565b61011b610116366004610daa565b6107f3565b6040516100bf989796959493929190610e09565b6002546100f0906001600160a01b031681565b6100ab610150366004610e61565b6108cd565b6100ab610163366004610e83565b61090b565b610170610b8f565b610178610bbc565b600084815260036020526040812080549091036101d35760405162461bcd60e51b815260206004820152601460248201527313dc99195c88191bd95cc81b9bdd08195e1a5cdd60621b60448201526064015b60405180910390fd5b6006810154600160a01b900460ff16156102265760405162461bcd60e51b815260206004820152601460248201527313dc99195c88185b1c9958591e4818db1bdcd95960621b60448201526064016101ca565b6000841180156102365750600083115b6102825760405162461bcd60e51b815260206004820152601d60248201527f507269636573206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60028101849055600381018390556001810161029e8382610f5c565b5060068101805460ff60a01b1916600160a01b179055600080858511156103aa57600583015486906102d08288611031565b6102da919061104a565b6102e49190611061565b915081905060006102f6600283611061565b9050600081856005015461030a9190611083565b600254600687015460405163a9059cbb60e01b81526001600160a01b03918216600482015260248101849052929350169063a9059cbb906044016020604051808303816000875af1158015610363573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103879190611096565b6103a35760405162461bcd60e51b81526004016101ca906110b8565b505061053c565b8585101561049b57600583015486906103c38783611031565b6103cd919061104a565b6103d79190611061565b90506103e2816110ef565b915060008184600501546103f69190611031565b9050801561049557600254600685015460405163a9059cbb60e01b81526001600160a01b0391821660048201526024810184905291169063a9059cbb906044016020604051808303816000875af1158015610455573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104799190611096565b6104955760405162461bcd60e51b81526004016101ca906110b8565b5061053c565b6002546006840154600585015460405163a9059cbb60e01b81526001600160a01b03928316600482015260248101919091526000945091169063a9059cbb906044016020604051808303816000875af11580156104fc573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105209190611096565b61053c5760405162461bcd60e51b81526004016101ca906110b8565b6004830182905560068301546040805189815260208101859052808201899052606081018890526001600160a01b039092166080830152517f06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d969181900360a00190a15050506105aa60018055565b50505050565b6002546040516370a0823160e01b81523060048201526000916001600160a01b0316906370a0823190602401602060405180830381865afa1580156105f9573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061061d919061110b565b905090565b61062a610b8f565b6106346000610c15565b565b61063e610b8f565b610646610bbc565b6001600160a01b0382166106915760405162461bcd60e51b8152602060048201526012602482015271496e76616c696420746f206164647265737360701b60448201526064016101ca565b600081116106e15760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b60025460405163a9059cbb60e01b81526001600160a01b038481166004830152602482018490529091169063a9059cbb906044016020604051808303816000875af1158015610734573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107589190611096565b6107a45760405162461bcd60e51b815260206004820152601c60248201527f555344542061697264726f70207472616e73666572206661696c65640000000060448201526064016101ca565b604080516001600160a01b0384168152602081018390527f8c32c568416fcf97be35ce5b27844cfddcd63a67a1a602c3595ba5dac38f303a910160405180910390a16107ef60018055565b5050565b6003602052600090815260409020805460018201805491929161081590610ed4565b80601f016020809104026020016040519081016040528092919081815260200182805461084190610ed4565b801561088e5780601f106108635761010080835404028352916020019161088e565b820191906000526020600020905b81548152906001019060200180831161087157829003601f168201915b50505060028401546003850154600486015460058701546006909701549596929591945092506001600160a01b0381169060ff600160a01b9091041688565b6108d5610b8f565b6001600160a01b0381166108ff57604051631e4fbdf760e01b8152600060048201526024016101ca565b61090881610c15565b50565b610913610bbc565b600081116109635760405162461bcd60e51b815260206004820152601d60248201527f416d6f756e74206d7573742062652067726561746572207468616e203000000060448201526064016101ca565b600083815260036020526040902054156109bf5760405162461bcd60e51b815260206004820152601760248201527f4f7264657220494420616c72656164792065786973747300000000000000000060448201526064016101ca565b6002546040516323b872dd60e01b8152336004820152306024820152604481018390526001600160a01b03909116906323b872dd906064016020604051808303816000875af1158015610a16573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a3a9190611096565b610a7d5760405162461bcd60e51b81526020600482015260146024820152731554d115081d1c985b9cd9995c8819985a5b195960621b60448201526064016101ca565b604080516101008101825284815260208082018581526000838501819052606084018190526080840181905260a084018690523360c085015260e0840181905287815260039092529290208151815591519091906001820190610ae09082610f5c565b506040828101516002830155606083015160038301556080830151600483015560a0830151600583015560c08301516006909201805460e0909401511515600160a01b026001600160a81b03199094166001600160a01b0390931692909217929092179055517fee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b590610b79908590859085903390611124565b60405180910390a1610b8a60018055565b505050565b6000546001600160a01b031633146106345760405163118cdaa760e01b81523360048201526024016101ca565b600260015403610c0e5760405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c0060448201526064016101ca565b6002600155565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b634e487b7160e01b600052604160045260246000fd5b600082601f830112610c8c57600080fd5b813567ffffffffffffffff811115610ca657610ca6610c65565b604051601f8201601f19908116603f0116810167ffffffffffffffff81118282101715610cd557610cd5610c65565b604052818152838201602001851015610ced57600080fd5b816020850160208301376000918101602001919091529392505050565b60008060008060808587031215610d2057600080fd5b843593506020850135925060408501359150606085013567ffffffffffffffff811115610d4c57600080fd5b610d5887828801610c7b565b91505092959194509250565b80356001600160a01b0381168114610d7b57600080fd5b919050565b60008060408385031215610d9357600080fd5b610d9c83610d64565b946020939093013593505050565b600060208284031215610dbc57600080fd5b5035919050565b6000815180845260005b81811015610de957602081850181015186830182015201610dcd565b506000602082860101526020601f19601f83011685010191505092915050565b88815261010060208201526000610e2461010083018a610dc3565b6040830198909852506060810195909552608085019390935260a08401919091526001600160a01b031660c0830152151560e09091015292915050565b600060208284031215610e7357600080fd5b610e7c82610d64565b9392505050565b600080600060608486031215610e9857600080fd5b83359250602084013567ffffffffffffffff811115610eb657600080fd5b610ec286828701610c7b565b93969395505050506040919091013590565b600181811c90821680610ee857607f821691505b602082108103610f0857634e487b7160e01b600052602260045260246000fd5b50919050565b601f821115610b8a57806000526020600020601f840160051c81016020851015610f355750805b601f840160051c820191505b81811015610f555760008155600101610f41565b5050505050565b815167ffffffffffffffff811115610f7657610f76610c65565b610f8a81610f848454610ed4565b84610f0e565b6020601f821160018114610fbe5760008315610fa65750848201515b600019600385901b1c1916600184901b178455610f55565b600084815260208120601f198516915b82811015610fee5787850151825560209485019460019092019101610fce565b508482101561100c5786840151600019600387901b60f8161c191681555b50505050600190811b01905550565b634e487b7160e01b600052601160045260246000fd5b818103818111156110445761104461101b565b92915050565b80820281158282048414176110445761104461101b565b60008261107e57634e487b7160e01b600052601260045260246000fd5b500490565b808201808211156110445761104461101b565b6000602082840312156110a857600080fd5b81518015158114610e7c57600080fd5b6020808252601c908201527f55534454207472616e7366657220746f2075736572206661696c656400000000604082015260600190565b6000600160ff1b82016111045761110461101b565b5060000390565b60006020828403121561111d57600080fd5b5051919050565b84815260806020820152600061113d6080830186610dc3565b6040830194909452506001600160a01b03919091166060909101529291505056fea2646970667358221220622013ae607897e52b6504cb63250462abcd2849be6b3bf3a2616da2a79771a164736f6c634300081d003300000000000000000000000036e6504c968f5c2a310b6af7b97bc22cdd3402cc"

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
func GetAirdropInput(_to common.Address, _amount *big.Int) ([]byte, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		log.Warnf("GetAirdropInput: abi is nil")
		return nil, err
	}
	if parsed == nil {
		log.Warnf("GetAirdropInput  parsed : abi is nil")
		return nil, errors.New("GetABI returned nil")
	}

	input, err := parsed.Pack("airdrop", _to, _amount)
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
	//fmt.Println("原始响应:", string(bodyBytes))

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
	//TODO 单价和gas规定标准进行修改
	//gasPrice := new(big.Int)
	//gasPrice.SetString("300000000000", 10) // 200 gwei
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error(" suggestGasPrice:", err)
		return "", "", err
	}

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
	//fmt.Println("strMillis:", strMillis)

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
	//fmt.Println("原始响应:", string(bodyBytes))

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
	// 自定义 HTTP 客户端，注入 x-api-key Header
	httpClient := &http.Client{
		Transport: &transportWithHeader{
			headers: map[string]string{
				"x-api-key": conf.X_Api_Key,
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
