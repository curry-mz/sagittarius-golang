package game

import (
	"time"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	RoleFarmer   = 0
	RoleLandlord = 1
)

type UserId int64

type UserInfo struct {
	UserId   UserId `json:"user_id"`
	Username string `json:"username"`
}

type Client struct {
	UserInfo   *UserInfo
	Room       *Room
	Table      *Table
	HandPokers []int64
	Ready      bool
	Next       *Client //链表
	IsRobot    bool
	toRobot    chan []interface{} //发送给robot的消息
	toServer   chan []interface{} //robot发送给服务器
}

// 重置状态
