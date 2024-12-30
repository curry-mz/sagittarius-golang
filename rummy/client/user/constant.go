package user

const (
	Name = "zues.hera.center.user"
)

const (
	getAIUrl    = "/server_api/user/v1/ai/list"
	getUsersUrl = "/server_api/user/v1/user/infos"
)

type AIData struct {
	PlayerID int64  `json:"playerID"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type GetAIResponse struct {
	Status  int       `json:"status"`
	Message string    `json:"message"`
	Data    []*AIData `json:"data"`
}

type UserDetail struct {
	UserID     int64  `json:"userID"`
	ChannelID  string `json:"channelID"`
	Player     string `json:"player"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	CreateTime int64  `json:"createTime"`
}

type UserInfosReq struct {
	UserIDs []int64 `json:"userIDs"`
}
type GetUserInfosResponse struct {
	Status  int           `json:"status"`
	Message string        `json:"message"`
	Data    []*UserDetail `json:"data"`
}
