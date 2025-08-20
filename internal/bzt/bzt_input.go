package bzt

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
	"valueguard/internal/api"
)

func GetOpenOrderInput(_orderId *big.Int, _tokenName string, _amount *big.Int) ([]byte, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errors.New("GetABI returned nil")
	}
	input, err := parsed.Pack("openOrder", _orderId, _tokenName, _amount)
	if err != nil {
		return nil, err
	}
	return input, nil
}
func GetCloseOrderInput(orderId, openPrice, closePrice *big.Int, tokenName string) ([]byte, error) {

	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errors.New("GetABI returned nil")
	}

	input, err := parsed.Pack("closeOrder", orderId, openPrice, closePrice, tokenName)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func GetAirdropInput(_to common.Address, _amount *big.Int) ([]byte, error) {
	parsed, err := BztMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errors.New("GetABI returned nil")
	}

	input, err := parsed.Pack("airdrop", _to, _amount)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func SendTransaction(
	private string, //钱包私钥
	cli *ethclient.Client, //节点client
	gasLimit uint64, //gas数量
	data []byte,
	to common.Address, //接受地址
	// value *big.Int,
) (string, error) {

	//构造签名相关参数
	opts, err := NewTransferOpt(private, api.ChainId)
	log.Info(api.ChainId)
	if err != nil {
		return "", err
	}

	//获取单价
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	//获取nonce
	nonce, err := cli.PendingNonceAt(context.Background(), common.HexToAddress("0xc020e62ce44297e86dA12CF15CfDc20B83eF3b72"))
	if err != nil {
		return "", err
	}

	//构造交易
	rawTx := types.NewTx(&types.LegacyTx{
		//To:      nil,
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		//Value:    value,
		Data: data,
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
