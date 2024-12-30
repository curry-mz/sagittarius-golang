package game

import (
	"context"
	"github.com/curry-mz/sagittarius-golang/logger"
	"landlord/common"
	"sync"
	"time"
)

const (
	GameWaitting = iota
	GamePlaying
	GameEnd
)

type Table struct {
	Lock         sync.RWMutex
	TableId      string
	State        int
	Creator      *Client
	TableClients map[UserId]*Client
	GameManage   GameManage
}

type GameManage struct {
	Turn           *Client
	LastShotClient *Client
	AnPokers       []int64 //暗牌区
	MingPokers     []int64 //名牌区
	LastShotPoker  int64   //上次出牌
	LastShotTime   int64
}

func (table *Table) joinTable(c *Client) {
	table.Lock.Lock()
	defer table.Lock.Unlock()
	if len(table.TableClients) > 2 {
		logger.Error(context.Background(), "Player[%d] JOIN Table[%d] FULL", c.UserInfo.UserId, table.TableId)
		return
	}
	logger.Debug(context.Background(), "[%v] user [%v] request join table", c.UserInfo.UserId, c.UserInfo.Username)
	if _, ok := table.TableClients[c.UserInfo.UserId]; ok {
		logger.Error(context.Background(), "[%v] user [%v] already in this table", c.UserInfo.UserId, c.UserInfo.Username)
		return
	}

	c.Table = table
	c.Ready = true
	if len(table.TableClients) == 1 {
		table.Creator = c
	}
	for _, client := range table.TableClients {
		if client.Next == nil {
			client.Next = c
			break
		}
	}
	table.TableClients[c.UserInfo.UserId] = c
	if len(table.TableClients) == 3 {
		c.Next = table.Creator
		table.State = GamePlaying
		table.dealPoker()
	}
}
func (table *Table) dealPoker() {
	table.GameManage.Pokers = make([]int, 0)
	for i := 0; i < 54; i++ {
		table.GameManage.Pokers = append(table.GameManage.Pokers, i)
	}
	table.ShufflePokers()
	for i := 0; i < 13; i++ {
		for _, client := range table.TableClients {
			client.HandPokers = append(client.HandPokers, table.GameManage.Pokers[len(table.GameManage.Pokers)-1])
			table.GameManage.Pokers = table.GameManage.Pokers[:len(table.GameManage.Pokers)-1]
		}
	}
	response := make([]interface{}, 0, 3)
	response = append(append(append(response, common.ResDealPoker), table.GameManage.FirstCallScore.UserInfo.UserId), nil)
	for _, client := range table.TableClients {
		response[len(response)-1] = client.HandPokers
		client.sendMsg(response)
	}
}
func (table *Table) ShufflePokers() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := len(table.GameManage.Pokers)
	for i > 0 {
		randIndex := r.Intn(i)
		table.GameManage.Pokers[i-1], table.GameManage.Pokers[randIndex] = table.GameManage.Pokers[randIndex], table.GameManage.Pokers[i-1]
		i--
	}
}
