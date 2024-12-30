package finance

const (
	Name = "zues.hermes.finance.busi"
)

const (
	queryUrl  = "/server_api/finance/v1/user/balance"
	betUrl    = "/server_api/finance/v1/user/bet"
	settleUrl = "/server_api/finance/v1/user/settle"
)

type QueryRequest struct {
	PlayerIDs []int64 `json:"userIDs"`
}

type QueryResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []*struct {
		PlayerID int64 `json:"userID"`
		Chips    int64 `json:"balance"`
	} `json:"data"`
}

type BetRequest struct {
	GameCode  int32  `json:"gameCode"`
	PlayerID  int64  `json:"userID"`
	OrderNo   string `json:"orderNum"`
	Chips     int64  `json:"amount"`
	Currency  string `json:"currency"`
	Timestamp int64  `json:"timestamp"`
}

type BetResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SettleRequest struct {
	GameCode int32         `json:"gameCode"`
	Orders   []SettleOrder `json:"orders"`
}

type SettleOrder struct {
	PlayerID int64  `json:"userID"`
	OrderNo  string `json:"orderNum"`
	Chips    int64  `json:"amount"`
	Currency string `json:"currency"`
}

type SettleResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    SettleData `json:"data"`
}

type SettleData struct {
	Users map[int64]struct {
		Balance  int64  `json:"balance"`
		Currency string `json:"currency"`
	} `json:"users"`
	Orders map[string]struct{} `json:"orders"`
}
