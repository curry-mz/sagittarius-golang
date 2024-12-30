package gateway

const (
	Name = "zues.base.gateway.websocket"
)

const (
	PushMessageUrl = "/server_api/v1/gw/ws/message"
)

const (
	PushModeBroadCast = 1 // app内广播
	PushModeGroup     = 2 // 组内广播
	PushModeSpecifyTo = 3 // 指定id推送
)

type PushMessageRequest struct {
	App    int32   `json:"app"`  // required
	Mode   int     `json:"mode"` // 1 2 3
	Group  string  `json:"group,omitempty"`
	Except []int64 `json:"except,omitempty"`
	To     []int64 `json:"to,omitempty"`
	Data   struct {
		MessageID int32  `json:"messageID"`
		Body      []byte `json:"body"`
	} `json:"data"`
}

type PushMessageResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
