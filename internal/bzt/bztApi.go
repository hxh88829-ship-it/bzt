package bzt

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
	"valueguard/internal/api"
)

var bztAddr = "0x0d7a5cD806536Fa7c3bA8f580D7dB7144253dE4a"

func GetOwner() (string, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztCaller(con, cli)
	if err != nil {
		return "", err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	addr, err := ca.Owner(&opt)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func GetTokenBalance() (*big.Int, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztCaller(con, cli)
	if err != nil {
		return nil, err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	ba, err := ca.GetContractBalance(&opt)
	if err != nil {
		return nil, err
	}
	return ba, nil
}

func GetUsdToken() (string, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztCaller(con, cli)
	if err != nil {
		return "", err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	addr, err := ca.UsdtToken(&opt)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func GetOrders(order int64) (interface{}, error) {
	var out interface{}

	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztCaller(con, cli)
	if err != nil {
		return nil, err
	}
	opt := bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	out, err = ca.Orders(&opt, big.NewInt(order))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetAirdrop(toAddr string, amount int64) (*types.Transaction, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztTransactor(con, cli)
	if err != nil {
		return nil, err
	}
	pri, err := crypto.HexToECDSA("272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4")
	if err != nil {
		return nil, err
	}
	opt, err := bind.NewKeyedTransactorWithChainID(pri, big.NewInt(10086))
	if err != nil {
		return nil, err
	}
	tx, err := ca.Airdrop(opt, common.HexToAddress(toAddr), big.NewInt(amount))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetOpenOrder() (*types.Transaction, error) {
	return nil, nil
}

func GetCloseOrder(orderId, openPrice, closePrice int64, tokenName string) (*types.Transaction, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztTransactor(con, cli)
	if err != nil {
		return nil, err
	}
	pri, err := crypto.HexToECDSA("272fe71819fa8d8957737986b05535b72ae43ca17e71bbc22c97e04b3d9b78e4")
	if err != nil {
		return nil, err
	}
	opt, err := bind.NewKeyedTransactorWithChainID(pri, big.NewInt(10086))
	if err != nil {
		return nil, err
	}
	tx, err := ca.CloseOrder(opt, big.NewInt(orderId), big.NewInt(openPrice), big.NewInt(closePrice), tokenName)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetParseAirdrop(receipt *types.Receipt) (*BztAirdrop, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztFilterer(con, cli)
	if err != nil {
		return nil, err
	}
	ev := crypto.Keccak256Hash([]byte("Airdrop(address,uint256)"))
	var event *BztAirdrop
	for i, vlog := range receipt.Logs {
		if vlog == nil {
			continue
		}
		if len(vlog.Topics) > 0 && vlog.Topics[0] == ev && vlog.Address == con {
			event, err = ca.ParseAirdrop(*vlog)
			if err != nil {
				fmt.Printf(" Failed to parse log %d: %+v\n", i, vlog)
				return nil, err
			}
		}
	}
	return event, nil
}

func GetParseOrderClosed(receipt *types.Receipt) (*BztOrderClosed, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztFilterer(con, cli)
	if err != nil {
		log.Error("NewBztFilterer error", "err", err)
		return nil, err
	}
	if ca == nil {
		return nil, errors.New("BztFilterer is nil")
	}
	ev := crypto.Keccak256Hash([]byte("OrderClosed(uint256,int256,uint256,uint256,address)"))

	var result *BztOrderClosed
	if receipt == nil || len(receipt.Logs) == 0 {
		return nil, errors.New("empty or nil receipt")
	}

	for i, vlog := range receipt.Logs {
		if vlog == nil {
			continue
		}
		if len(vlog.Topics) > 0 && vlog.Topics[0] == ev && vlog.Address == con {
			event, err := ca.ParseOrderClosed(*vlog)
			if err != nil {
				fmt.Printf(" Failed to parse log %d: %+v\n", i, vlog)
				return nil, err
			}
			// ✅ 分字段打印
			fmt.Println("✅ Parsed OrderClosed Event:")
			fmt.Println("  📦 Order ID     :", event.OrderId.String())
			fmt.Println("  📈 Profit/Loss  :", event.ProfitLoss.String())
			fmt.Println("  🟢 Open Price   :", event.OpenPrice.String())
			fmt.Println("  🔴 Close Price  :", event.ClosePrice.String())
			fmt.Println("  👤 User         :", event.User.Hex())
			//0x06a2f7fd54de050efdb547068782f039f9d20511970de8f48241f64625f52d96
			result = event
			break // 假设每个交易只触发一次事件
		}
	}

	if result == nil {
		return nil, errors.New("OrderClosed event not found")
	}
	return result, nil
}

func GetParseOrderOpened(receipt *types.Receipt) (*BztOrderOpened, error) {
	cli := api.Client
	con := common.HexToAddress(bztAddr)
	ca, err := NewBztFilterer(con, cli)
	if err != nil {
		log.Error("NewBztFilterer error", "err", err)
		return nil, err
	}
	if ca == nil {
		return nil, errors.New("BztFilterer is nil")
	}
	ev := crypto.Keccak256Hash([]byte("OrderOpened(uint256,string,uint256,address)"))

	var result *BztOrderOpened
	if receipt == nil || len(receipt.Logs) == 0 {
		return nil, errors.New("empty or nil receipt")
	}

	for i, vlog := range receipt.Logs {
		if vlog == nil {
			continue
		}
		if len(vlog.Topics) > 0 && vlog.Topics[0] == ev && vlog.Address == con {
			event, err := ca.ParseOrderOpened(*vlog)
			if err != nil {
				fmt.Printf(" Failed to parse log %d: %+v\n", i, vlog)
				return nil, err
			}
			// ✅ 分字段打印
			fmt.Println("✅ Parsed OrderClosed Event:")
			fmt.Println("  📦 Order ID     :", event.OrderId.String())
			fmt.Println("  📈 TokenName    :", event.TokenName)
			fmt.Println("  🟢 Amount       :", event.Amount)
			fmt.Println("  👤 User         :", event.User.Hex())
			fmt.Println("transfer:", event.Raw.Topics[0].String())
			fmt.Println("transfer:", vlog.Topics[0].String())
			result = event
			//	0xee570f04775e144993314e5a0a45e525633d3c8d528ed5fa6fc49eb7bee161b5
			break
		}
	}

	if result == nil {
		return nil, errors.New("OrderClosed event not found")
	}
	return result, nil
}
