package mongo

// Users 用户
type Users struct {
	Uid             string `bson:"uid" json:"uid"`
	Name            string `bson:"name" json:"name"`
	Email           string `bson:"email" json:"email"`
	Address         string `bson:"address" json:"address"`
	CreateTimeAt    int64  `bson:"create_time_at" json:"create_time_at"`
	OriginalMessage string `bson:"original_message" json:"original_message"`
	Status          string `bson:"status" json:"status"` // 0.正常 1.异常
}

// 钱包
type Wallet struct {
	Address string `bson:"address" json:"address"`
	Balance string `bson:"balance" json:"balance"` //可用资产
	Symbol  string `bson:"symbol" json:"symbol"`
	Pledge  string `bson:"pledge" json:"pledge"` //质押中资产
}

// CoinPrice 行情价格
type CoinPrice struct {
	Symbol    string `bson:"symbol" json:"symbol"`
	Price     string `bson:"price" json:"price"`
	Timestamp int64  `bson:"timestamp" json:"timestamp"`
	Index     int    `bson:"index" json:"index"`
}

// Order 订单
type Order struct {
	OrderId        string `bson:"order_id" json:"order_id"`
	Symbol         string `bson:"symbol" json:"symbol"` // 质押币种
	OpenPrice      string `bson:"open_price" json:"open_price"`
	ClosePrice     string `bson:"close_price" json:"close_price"`
	ProfitLoss     string `bson:"profit_loss" json:"profit_loss"`
	Amount         string `bson:"amount" json:"amount"`
	UsersAddr      string `bson:"users_addr" json:"users_addr"`
	IsClosed       bool   `bson:"is_closed" json:"is_closed"` //nil表示待确定，false开仓中，true关仓
	OrderStartTime uint64 `bson:"order_start_time" json:"order_start_time"`
	OrderEndTime   uint64 `bson:"order_end_time" json:"order_end_time"`
}

type LossAmount struct {
	UserAddr string `bson:"user_addr" json:"user_addr"`
	Symbol   string `bson:"symbol" json:"symbol"`
	Amount   string `bson:"amount" json:"amount"`
	Counts   string `bson:"counts" json:"counts"`
	UpdateAt int64  `bson:"update_at" json:"update_at"`
}

// Airdrop 空投
type Airdrop struct {
	ToAddr      string `bson:"to_addr" json:"to_addr"`
	Amount      string `bson:"amount" json:"amount"`
	Symbol      string `bson:"symbol" json:"symbol"`             // 空投币种
	Status      uint64 `bson:"status" json:"status"`             // 状态 (0:失败 1:成功)
	AirdropTime uint64 `bson:"airdrop_time" json:"airdrop_time"` // 空投时间
	TxHash      string `bson:"tx_hash" json:"tx_hash"`           // 交易哈希
}

type WithdrawalAddress struct {
	UserAddr  string `bson:"user_addr" json:"user_addr"`   // 用户地址
	Symbol    string `bson:"symbol" json:"symbol"`         // 适用币种
	Amount    string `bson:"amount" json:"amount"`         // 金额
	TxHash    string `bson:"tx_hash" json:"tx_hash"`       // 交易哈希
	Status    int    `bson:"status" json:"status"`         // 状态 (0:失败 1:成功 )
	CreatedAt int64  `bson:"created_at" json:"created_at"` // 添加时间
}

type DepositRecord struct {
	UserAddr   string `bson:"user_addr" json:"user_addr"`     // 用户地址
	FromAddr   string `bson:"from_addr" json:"from_addr"`     // 来源地址
	Symbol     string `bson:"symbol" json:"symbol"`           // 币种
	Amount     string `bson:"amount" json:"amount"`           // 充值数量
	TxHash     string `bson:"tx_hash" json:"tx_hash"`         // 交易哈希
	Status     int    `bson:"status" json:"status"`           // 状态 (0:失败 1:成功 )
	CreateTime int64  `bson:"create_time" json:"create_time"` // 创建时间
}

// Transaction 交易记录
type Transaction struct {
	TxHash          string `bson:"tx_hash" json:"tx_hash"`
	From            string `bson:"from" json:"from"`
	To              string `bson:"to" json:"to"`
	Nonce           uint64 `bson:"nonce" json:"nonce"`
	Data            string `bson:"data" json:"data"`
	Time            uint64 `bson:"time" json:"time"`
	Number          uint64 `bson:"number" json:"number"`
	Value           string `bson:"value" json:"value"`
	Gas             uint64 `bson:"gas" json:"gas"`
	GasPrice        string `bson:"gas_price" json:"gas_price"`
	TransactionType string `bson:"transaction_type" json:"transaction_type"` //1。充值 2.提现。 3。转账
	Status          uint64 `bson:"status" json:"status"`
}

// ScanBlock 扫块记录
type ScanBlock struct {
	NetWork     uint64 `bson:"netWork" json:"netWork"`         //DTC
	LatestBlock uint64 `bson:"latestBlock" json:"latestBlock"` //当前
	Time        int64  `bson:"time" json:"time"`
}

type LossBlock struct {
	NetWork uint64 `bson:"netWork" json:"netWork"`
	BlockNr uint64 `bson:"blockNr" json:"blockNr"`
	Time    int64  `bson:"time" json:"time"`
	Reason  string `bson:"reason" json:"reason"`
}
