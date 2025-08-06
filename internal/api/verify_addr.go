package api

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

func ValidateAvatarURL(avatar string) error {
	if avatar == "" {
		return nil // 不上传头像允许为空
	}

	u, err := url.Parse(avatar)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("头像链接格式错误")
	}

	// 限定只接受 http / https
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("头像链接必须以 http/https 开头")
	}

	// 可选：限制来源为你自己的 CDN
	if !strings.HasSuffix(u.Host, "upmpc-test.s3.ap-southeast-1.amazonaws.com") {
		return errors.New("头像必须来自平台图片服务器")
	}

	// 可选：限制图片后缀
	if !strings.HasSuffix(u.Path, ".png") && !strings.HasSuffix(u.Path, ".jpg") && !strings.HasSuffix(u.Path, ".jpeg") {
		return errors.New("仅支持 PNG、JPG 格式头像")
	}

	return nil
}

func NormalizeTwitter(twitter string) (string, error) {
	twitter = strings.TrimSpace(twitter)
	twitter = strings.TrimPrefix(twitter, "@")

	if twitter == "" {
		return "", nil
	}
	// 最简单的合法字符检查（只允许英文字母数字下划线）
	if !regexp.MustCompile(`^[a-zA-Z0-9_]{1,15}$`).MatchString(twitter) {
		return "", errors.New("输入不合法")
	}

	return "https://x.com/" + twitter, nil
}

func ValidateWebsite(website string) error {
	website = strings.TrimSpace(website)

	if website == "" {
		return nil // 允许用户不填
	}

	// 限制最大长度（例如 200 个字符）
	if utf8.RuneCountInString(website) > 200 {
		return errors.New("网站链接过长")
	}

	// 解析 URL
	u, err := url.Parse(website)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("请输入合法的网址，例如 https://example.com")
	}

	// 只允许 http / https
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("仅支持 http 和 https 协议")
	}

	// 可选：禁止某些危险域名
	blocked := []string{"example.local", "localhost", "127.0.0.1"}
	for _, b := range blocked {
		if strings.Contains(u.Host, b) {
			return errors.New("网址不允许指向本地地址")
		}
	}

	return nil
}
