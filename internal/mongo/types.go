package mongo

// Users 用户
type Users struct {
	Uid             string `bson:"uid" json:"uid"`
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
	Timestamp uint64 `bson:"timestamp" json:"timestamp"`
	Index     uint64 `bson:"index" json:"index"`
}

// Order 订单
type Order struct {
	Uid            string `bson:"uid" json:"uid"`
	OrderId        string `bson:"order_id" json:"order_id"`
	Symbol         string `bson:"symbol" json:"symbol"` // 质押币种
	OpenPrice      string `bson:"open_price" json:"open_price"`
	ClosePrice     string `bson:"close_price" json:"close_price"`
	ProfitLoss     string `bson:"profit_loss" json:"profit_loss"`
	Amount         string `bson:"amount" json:"amount"`
	UsersAddr      string `bson:"users_addr" json:"users_addr"`
	IsClosed       uint64 `bson:"is_closed" json:"is_closed"` //0待确定，1开仓，2关仓
	OrderStartTime uint64 `bson:"order_start_time" json:"order_start_time"`
	OrderEndTime   uint64 `bson:"order_end_time" json:"order_end_time"`
	OpenTxHash     string `bson:"open_tx_hash" json:"open_tx_hash"`
	CloseTxHash    string `bson:"close_tx_hash" json:"close_tx_hash"`
	Status         uint64 `bson:"status" json:"status"` // 0 代表正常， 1代表异常
}

type RewardAmount struct {
	Symbol      string `bson:"symbol" json:"symbol"`
	UpdateAt    uint64 `bson:"update_at" json:"update_at"`
	TotalAmount string `bson:"total_amount" json:"total_amount"` // 总量
}

type UserAmount struct {
	Symbol       string `bson:"symbol" json:"symbol"`
	UpdateAt     int64  `bson:"update_at" json:"update_at"`
	LossAmount   string `bson:"loss_amount" json:"loss_amount"`
	Profit       string `bson:"profit" json:"profit"`
	UserAddr     string `bson:"user_addr" json:"user_addr"`
	ClaimAirdrop string `bson:"claim_airdrop" json:"claim_airdrop"`
}

// Airdrop 链上扫块空投
type Airdrop struct {
	Uid         string `bson:"uid" json:"uid"`
	OrderId     string `bson:"order_id" json:"order_id"`
	ToAddr      string `bson:"to_addr" json:"to_addr"`
	Amount      string `bson:"amount" json:"amount"`
	Symbol      string `bson:"symbol" json:"symbol"`             // 空投币种
	AirdropTime string `bson:"airdrop_time" json:"airdrop_time"` // 空投时间
	TxHash      string `bson:"tx_hash" json:"tx_hash"`           // 交易哈希
	Status      uint64 `bson:"status" json:"status"`             //0 不成功 1成功
}

// 。。。
type DailyAirdropTrade struct {
	Symbol    string `bson:"symbol" json:"symbol"`
	Reward    string `bson:"reward" json:"reward"`
	Date      string `bson:"date" json:"date"`
	PoolTotal string `bson:"pool_total" json:"pool_total"`
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
	TotalFee        uint64 `bson:"total_fee" json:"total_fee"`
	TransactionType string `bson:"transaction_type" json:"transaction_type"`
	Status          uint64 `bson:"status" json:"status"`
}

// ScanBlock 扫块记录
type ScanBlock struct {
	NetWork     uint64 `bson:"netWork" json:"netWork"`         //Dtt
	LatestBlock uint64 `bson:"latestBlock" json:"latestBlock"` //当前
	Time        int64  `bson:"time" json:"time"`
}

type LossBlock struct {
	NetWork uint64 `bson:"netWork" json:"netWork"`
	BlockNr uint64 `bson:"blockNr" json:"blockNr"`
	Time    int64  `bson:"time" json:"time"`
	Reason  string `bson:"reason" json:"reason"`
}

type BztDapp struct {
	AppId         int64  `bson:"app_id" json:"app_id"`
	DappIcon      string `bson:"dapp_icon" json:"dapp_icon"`
	DappIntroduce string `bson:"dapp_introduce" json:"dapp_introduce"`
	DappName      string `bson:"dapp_name" json:"dapp_name"`
	DappUrl       string `bson:"dapp_url" json:"dapp_url"`
}

type DeployTransaction struct {
	TxHash   string `bson:"tx_hash" json:"tx_hash"`
	Nonce    string `bson:"nonce" json:"nonce"`
	Data     string `bson:"data" json:"data"`
	Gas      uint64 `bson:"gas" json:"gas"`
	GasPrice string `bson:"gas_price" json:"gas_price"`
}

type OrderSwitch struct {
	Status  uint64 `bson:"status" json:"status"`
	ChainId uint64 `bson:"chain_id" json:"chain_id"`
	Types   string `bson:"types" json:"types"`
}

type Kline struct {
	OpenTime                 int64  `bson:"open_time" json:"open_time"`
	OpenPrice                string `bson:"open_price" json:"open_price"`
	HighPrice                string `bson:"high_price" json:"high_price"`
	LowPrice                 string `bson:"low_price" json:"low_price"`
	ClosePrice               string `bson:"close_price" json:"close_price"`
	Volume                   string `bson:"volume" json:"volume"`
	CloseTime                int64  `bson:"close_time" json:"close_time"`
	QuoteAssetVolume         string `bson:"quote_asset_volume" json:"quote_asset_volume"`
	NumberOfTrades           int    `bson:"number_of_trades" json:"number_of_trades"`
	TakerBuyBaseAssetVolume  string `bson:"taker_buy_base_asset_volume" json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume string `bson:"taker_buy_quote_asset_volume" json:"taker_buy_quote_asset_volume"`
	Ignore                   string `bson:"ignore" json:"ignore"`
	DataType                 string `bson:"data_type" json:"data_type"` //k线类别，日k或时k
	Symbol                   string `bson:"symbol" json:"symbol"`
}

type BinanceOrder struct {
	Address                 string `bson:"address" json:"address"`
	TxHash                  string `bson:"tx_hash" json:"tx_hash"`
	BztOrderId              string `bson:"bzt_order_id" json:"bzt_order_id"`
	Symbol                  string `bson:"symbol" json:"symbol"`
	OrderId                 int64  `bson:"order_id" json:"orderId"`
	OrderListId             int64  `bson:"order_list_id" json:"orderListId"`
	ClientOrderId           string `bson:"client_order_id" json:"clientOrderId"`
	Price                   string `bson:"price" json:"price"`
	OrigQty                 string `bson:"orig_qty" json:"origQty"`
	ExecutedQty             string `bson:"executed_qty" json:"executedQty"`
	CummulativeQuoteQty     string `bson:"cumulative_quote_qty" json:"cummulativeQuoteQty"`
	Status                  string `bson:"status" json:"status"`
	TimeInForce             string `bson:"time_in_force" json:"timeInForce"`
	Type                    string `bson:"type" json:"type"`
	Side                    string `bson:"side" json:"side"`
	StopPrice               string `bson:"stop_price" json:"stopPrice"`
	IcebergQty              string `bson:"iceberg_qty" json:"icebergQty"`
	Time                    int64  `bson:"time" json:"time"`
	UpdateTime              int64  `bson:"update_time" json:"updateTime"`
	IsWorking               bool   `bson:"is_working" json:"isWorking"`
	WorkingTime             int64  `bson:"working_time" json:"workingTime"`
	OrigQuoteOrderQty       string `bson:"orig_quote_order_qty" json:"origQuoteOrderQty"`
	SelfTradePreventionMode string `bson:"self_trade_prevention_mode" json:"selfTradePreventionMode"`
}
