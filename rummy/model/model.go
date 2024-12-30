package model

import "github.com/shopspring/decimal"

const (
	ReqLogin    = 501
	AckLogin    = 502
	ReqTables   = 503
	AckTables   = 504
	ReqRankList = 505
	AckRankList = 506
)

type ProxyMessage struct {
	App       int32  `json:"app"`
	MessageID int32  `json:"messageID"`
	Body      []byte `json:"body"`
}

type DisconnectRequest struct {
	App      int32 `json:"app"`
	PlayerID int64 `json:"playerID"`
}
type KafkaOrderInfo struct {
	Action        string          `json:"action"`
	PlayerId      int32           `json:"player_id"`       // 玩家id
	MerId         int32           `json:"mer_id"`          // 商户id
	MerName       string          `json:"mer_name"`        // 商户名称
	GameId        int32           `json:"game_id"`         // 游戏id
	GameName      string          `json:"game_name"`       // 游戏名称
	GameRound     string          `json:"game_round"`      // 期号
	Currency      string          `json:"currency"`        // 币种
	BetAmount     decimal.Decimal `json:"bet_amount"`      // 下注金额
	SettleAmount  decimal.Decimal `json:"settle_amount"`   // 结算金额
	JackpotAmount decimal.Decimal `json:"jackpot_amount"`  // jackpot
	Status        int8            `json:"status"`          //订单状态1未开奖2已开奖3回滚 预留
	SettleStatus  int8            `json:"settle_status"`   //结算状态1未结算2赢3输4平  预留
	BetTime       int64           `json:"bet_time"`        //下注时间
	SettleTime    int64           `json:"settle_time"`     //结算时间
	ItemId        int64           `json:"item_id;"`        //下注类型 返回0，1，2，3
	ServiceCharge decimal.Decimal `json:"service_charge;"` //服务费
	OrderNo       string          `json:"order_no"`
}
