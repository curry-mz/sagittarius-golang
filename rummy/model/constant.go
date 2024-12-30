package model

// 房间状态
const (
	RoomStateReady    = 0
	RoomStateBegin    = 1
	RoomStateBetBegin = 2
	RoomStateBetOver  = 3
	RoomStateLottery  = 4
	RoomStateSettle   = 5
)

// 下注区域
const (
	ItemRed     int64 = 1
	ItemBlack         = 2
	ItemSpecial       = 11
)

var (
	Items   = []int64{ItemRed, ItemBlack}
	ItemAll = []int64{ItemRed, ItemBlack, ItemSpecial}

	ItemMap = map[int64]struct{}{
		ItemRed:     struct{}{},
		ItemBlack:   struct{}{},
		ItemSpecial: struct{}{},
	}
)

// 特殊牌型
const (
	HighCard = 0  // 高牌
	Pair     = 1  // 小对子[2-8]
	BigPair  = 11 // 大对子[9-A]
	Color    = 12 // 同花
	Run      = 13 // 顺子
	PureRun  = 14 // 同花顺
	Trail    = 15 // 三条
)

var (
	CardTypes = []int{HighCard, Pair, BigPair, Color, Run, PureRun, Trail}
)

var RateMap = map[int64]int64{
	ItemRed:   2,
	ItemBlack: 2,
}

var SpecialRateMap = map[int]int{
	BigPair: 2,
	Color:   3,
	Run:     4,
	PureRun: 12,
	Trail:   20,
}

//var WeightMap = map[int]int{
//	HighCard: 500,
//	Pair:     200,
//	BigPair:  200,
//	Color:    100,
//	Run:      50,
//	PureRun:  5,
//	Trail:    1,
//}
