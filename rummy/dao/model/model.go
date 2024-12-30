package model

import (
	"code.cd.local/games-go/mini/rummy.busi/model"
	"encoding/json"
	"strconv"
	"strings"
)

type OrderPO struct {
	ID        int64  `gorm:"column:id"`
	OrderNo   string `gorm:"column:order_no"`
	PlayerID  int64  `gorm:"column:player_id"`
	RoomID    int32  `gorm:"column:room_id"`
	RoundNo   string `gorm:"column:round_no"`
	ItemID    int64  `gorm:"column:item_id"`
	BetChips  int64  `gorm:"column:bet_chips"`
	Currency  string `gorm:"column:currency"`
	Prize     int64  `gorm:"column:prize"`
	Timestamp int64  `gorm:"column:timestamp"`
	Status    int8   `gorm:"column:status"` // 0下单成功 1已结算 2结算失败
}

type StockPO struct {
	ID              int64 `gorm:"column:id"`
	RoomID          int32 `gorm:"column:room_id"`
	Number          int64 `gorm:"column:number"`
	TotalAccept     int64 `gorm:"column:total_accept"`
	TotalIncome     int64 `gorm:"column:total_income"`
	TotalPrize      int64 `gorm:"column:total_prize"`
	TotalCommission int64 `gorm:"column:total_commission"`
}

type LotteryPO struct {
	ID              int64  `gorm:"column:id"`
	RoomID          int32  `gorm:"column:room_id"`
	RoundNo         string `gorm:"column:round_no"`
	ResultMode      int8   `gorm:"column:result_mode"` // 这里0表示随机 1表示杀率控制
	Winner          int64  `gorm:"column:winner"`
	RedCards        string `gorm:"column:red_cards"`
	RedType         int    `gorm:"column:red_type"`
	BlackCards      string `gorm:"column:black_cards"`
	BlackType       int    `gorm:"column:black_type"`
	TotalBet        int64  `gorm:"column:total_bet"`
	TotalPrize      int64  `gorm:"column:total_prize"`
	SettleTimestamp int64  `gorm:"column:settle_timestamp"`
	HitRiskControl  bool   `gorm:"column:hit_risk_control"`
	Stock           int64  `gorm:"column:stock"`
	Krate           int64  `gorm:"column:krate"`
	HitKrate        bool   `gorm:"column:hit_krate"`
}

func (lp *LotteryPO) GetRedCards() (redCards []int32) {
	rs := strings.Split(lp.RedCards, ",")
	for _, s := range rs {
		v, _ := strconv.Atoi(s)
		redCards = append(redCards, int32(v))
	}

	return
}
func (lp *LotteryPO) GetBlockCards() (blockCards []int32) {
	bs := strings.Split(lp.BlackCards, ",")
	for _, s := range bs {
		v, _ := strconv.Atoi(s)
		blockCards = append(blockCards, int32(v))
	}
	return
}

func (lp *LotteryPO) GetWin() (winItem int64, winCards []int32, winType int) {
	winItem = lp.Winner
	if lp.Winner == model.ItemRed {
		winCards = lp.GetRedCards()
		winType = lp.RedType
	} else {
		winCards = lp.GetBlockCards()
		winType = lp.BlackType
	}
	return
}

type HistoryPO struct {
	ID         int64  `gorm:"column:id"`
	PlayerID   int64  `gorm:"column:player_id"`
	RoomID     int32  `gorm:"column:room_id"`
	RoundNo    string `gorm:"column:round_no"`
	BeginTime  int64  `gorm:"column:begin_time"`
	EndTime    int64  `gorm:"column:end_time"`
	ItemID     int64  `gorm:"column:item_id"`
	TotalBet   int64  `gorm:"column:total_bet"`
	TotalPrize int64  `gorm:"column:total_prize"`
}

type AIPlayer struct {
	PlayerID int64  `json:"playerID"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (aip *AIPlayer) ToJson() ([]byte, error) {
	return json.Marshal(aip)
}

type CacheRoundBet struct {
	ItemID   int64   `json:"itemID"`
	Chips    []int64 `json:"chips"`
	Currency string  `json:"currency"`
}
type UserHistoryPO struct {
	ID              int64  `gorm:"column:id"`
	PlayerID        int64  `gorm:"column:player_id"`
	RoomID          int32  `gorm:"column:room_id"`
	RoundNo         string `gorm:"column:round_no"`
	BeginTime       int64  `gorm:"column:begin_time"`
	EndTime         int64  `gorm:"column:end_time"`
	ItemID          int64  `gorm:"column:item_id"`
	TotalBet        int64  `gorm:"column:total_bet"`
	TotalPrize      int64  `gorm:"column:total_prize"`
	RedCards        string `gorm:"column:red_cards"`
	RedType         int64  `gorm:"column:red_type"`
	BlackCards      string `gorm:"column:black_cards"`
	BlackType       int64  `gorm:"column:black_type"`
	Winner          int64  `gorm:"column:winner"`
	SettleTimestamp int64  `gorm:"column:settle_timestamp"`
}
