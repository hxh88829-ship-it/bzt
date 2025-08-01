package api

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TransferEvent struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}
