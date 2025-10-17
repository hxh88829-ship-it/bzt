package conf

/*
apikey        签名服务的apikey
baseurl       签名服务的ip
key_id        owner钱包对应的kms keyid
owner_address owner钱包对应的钱包地址
rpc_url       节点url
x_api_key     节点请求需要的key
hmacKey       hash消息认证码
*/

var (
	Apikey               string
	BaseUrl              string
	KeyId                string
	OwnerAddress         string
	HmacKey              string
	RpcUrl               string
	ContractBztAddr      string
	ContractDusdtAddress string
	X_Api_Key            string
	Secret               string
	BinanceApikey        string
	BinanceSecretKey     string
)

// XApiKey      string
