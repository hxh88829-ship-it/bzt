package mongo

var (
	MonCli *MongoClient
)

const (
	DatabaseNameForChain = "bzt_hx"
	user                 = "user"
	scanBlock            = "scanBlock"
	newPrice             = "newPrice"
	airdrop              = "airdrop"
	rewardPool           = "rewardPool"
	lossAmount           = "lossAmount"
	order                = "order"
	lossBlock            = "lossBlock"
	transaction          = "transaction"
	dailyAirdrops        = "dailyAirdrops"
	bztDapp              = "bztDapp"
	deployContract       = "deployContract"
	orderSwitch          = "orderSwitch"
	kLineByOneDay        = "kLineByOneDay"
	kLineByThreeDay      = "kLineByThreeDay"
	kLineByOneHour       = "kLineByOneHour"
	kLineByFourHour      = "kLineByFourHour"
)
