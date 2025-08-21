package api

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

func VerifyForAddress(OriginalMessage, signature string) (string, error) {
	Hash := ComputeMessageHash(OriginalMessage)

	if len(signature) != 132 {
		return "", errors.New("signature length is wrong")
	}

	Sign, err := hexutil.Decode(signature)
	if err != nil {
		return "", err
	}

	if len(Sign) != 65 {
		return "", fmt.Errorf("signature must be 65 bytes long")
	}
	// see crypto.Ecrecover description
	if Sign[64] == 27 || Sign[64] == 28 {
		Sign[64] -= 27
	}

	pubKey, err := crypto.Ecrecover(Hash, Sign) //crypto.Keccak256(),
	if err != nil {
		return "", err
	}

	if len(pubKey) < 13 {
		return "", errors.New("pubKey length is wrong")
	}

	var addr common.Address
	copy(addr[:], crypto.Keccak256(pubKey[1:])[12:])

	if len(addr.String()) != 42 {
		return "", errors.New("address length is wrong")
	}

	return strings.ToLower(addr.String()), nil
}

func HashData(message string) []byte {
	data := common.FromHex(message)
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// 计算消息哈希 (与MetaMask一致)
func ComputeMessageHash(message string) []byte {
	formatted := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256([]byte(formatted))
	return hash // 0x...格式
}
