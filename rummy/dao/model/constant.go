package model

const (
	RoundPointsKey          = "gg:tpv2:kv:rp:%d:%s" // 房间本轮骰子值信息
	ChannelWinRankKey       = "gg:rm:zset:%s"       //渠道下面的赢家排行榜
	ChannelRoomMatchKey     = "gg:rm:zset:%s:%d"    //某个渠道下的某个场次匹配池
	ChannelWinRankKeyExpire = -1
	UserWinCoinKey          = "gg:rm:kv:%s:%d" //某个渠道下的某个用户赢的钱
)

const (
	OrderTable   = "order_%04d%02d"
	StockTable   = "stock"
	LotteryTable = "lottery_result"
	HistoryTable = "history_%04d%02d"
)

const (
	OrderStateCreate       = 0
	OrderStateSettleOK     = 1
	OrderStateSettleFailed = 2
)

const (
	UseAlgorithm int8 = 1
	UseRandom    int8 = 2
)
