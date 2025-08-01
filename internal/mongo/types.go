package mongo

// 用户
type Users struct {
	Uid             string `bson:"uid" json:"uid"`
	Name            string `bson:"name" json:"name"`
	Email           string `bson:"email" json:"email"`
	Avatar          string `bson:"avatar" json:"avatar"`
	Address         string `bson:"address" json:"address"`
	CreateTimeAt    int64  `bson:"create_time_at" json:"create_time_at"`
	OriginalMessage string `bson:"original_message" json:"original_message"`
}

type CoinPrice struct {
	Symbol    string `bson:"symbol" json:"symbol"`
	Price     string `bson:"price" json:"price"`
	Timestamp int64  `bson:"timestamp" json:"timestamp"`
	Index     int    `bson:"index" json:"index"`
}

type Transaction struct {
	TxHash   string        `bson:"tx_hash" json:"tx_hash"`
	From     string        `bson:"from" json:"from"`
	To       string        `bson:"to" json:"to"`
	Cost     string        `bson:"cost" json:"cost"`
	Nonce    uint64        `bson:"nonce" json:"nonce"`
	Data     string        `bson:"data" json:"data"`
	Time     uint64        `bson:"time" json:"time"`
	Number   uint64        `bson:"number" json:"number"`
	Value    string        `bson:"value" json:"value"`
	GasLimit uint64        `bson:"gas_limit" json:"gas_limit"`
	GasPrice string        `bson:"gas_price" json:"gas_price"`
	Logs     TransferEvent `bson:"logs" json:"logs"`
	ChainID  uint64        `bson:"chain_id" json:"chain_id"`
	Method   string        `bson:"method" json:"method"`
	Status   uint64        `bson:"status" json:"status"`
}

type TransferEvent struct {
	ToAddress string `bson:"to_address" json:"to_address"`
	TokenId   string `bson:"token_id" json:"token_id"`
	Contract  string `bson:"contract" json:"contract"`
}

// 扫块记录
type ScanBlock struct {
	NetWork     uint64 `bson:"netWork" json:"netWork"`         //DTC
	LatestBlock uint64 `bson:"latestBlock" json:"latestBlock"` //当前
	Time        int64  `bson:"time" json:"time"`
}
