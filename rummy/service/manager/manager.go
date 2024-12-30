package manager

import (
	"code.cd.local/games-go/mini/rummy.busi/client/gateway"
	"code.cd.local/games-go/mini/rummy.busi/client/user"
	"code.cd.local/games-go/mini/rummy.busi/code"
	"code.cd.local/games-go/mini/rummy.busi/conf"
	"code.cd.local/games-go/mini/rummy.busi/dao"
	"code.cd.local/games-go/mini/rummy.busi/model"
	pb "code.cd.local/games-go/mini/rummy.busi/model/pb"
	"code.cd.local/games-go/mini/rummy.busi/service/game"
	"code.cd.local/games-go/mini/rummy.busi/utils"
	"code.cd.local/sagittarius/sagittarius-golang/cores/server/http"
	"code.cd.local/sagittarius/sagittarius-golang/logger"
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"sync"
)

type Manager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	dao       *dao.SudDao
	sfNode    *snowflake.Node
	rooms     sync.Map // roomID- *room
	aiCenters sync.Map // roomID - *ai.Center
	Tables    map[string]*game.Table
}

const GAMENAME = "rummy"

func New(ctx context.Context) (*Manager, error) {
	m := Manager{}
	d, err := dao.NewDao()
	if err != nil {
		return nil, err
	}
	m.dao = d
	m.ctx, m.cancel = context.WithCancel(ctx)
	// 雪花算法节点
	sfNode, err := utils.CreateSnowFlakeNode()
	if err != nil {
		return nil, err
	}
	m.sfNode = sfNode
	for _, channel := range conf.ExtraConfig().Channels {
		for _, rc := range channel.Rooms {
			if _, has := m.rooms.Load(rc.RoomID); !has {
				r, err := game.NewRoom(ctx, rc.RoomID, rc.RoomName, m.dao, rc.MinAccount, rc.MaxAccount, rc.Fee, rc.BaseScore)
				if err != nil {
					logger.Error(ctx, "watchGame game.NewRoom err:%v", err)
					continue
				}
				m.rooms.Store(rc.RoomID, r)
			}
		}
	}
	return &m, nil
}

func (m *Manager) Stop() {
	//m.rooms.Range(func(key, value any) bool {
	//	r := value.(*game.Room)
	//	r.Stop()
	//
	//	return true
	//})
	m.cancel()
}

func (m *Manager) ProxyMessageHandler(c *http.Context) {
	var req model.ProxyMessage
	if err := c.Bind(nil, &req); err != nil {
		logger.Error(c.Ctx(), "ProxyMessageHandler bind data err:%v", err)
		_ = c.JsonErr(code.ParamError)
		return
	}
	if req.App == 0 || req.MessageID == 0 {
		_ = c.JsonErr(code.ParamError)
		return
	}
	switch req.MessageID {
	case model.ReqLogin:
		var data pb.ReqLogin
		if err := proto.Unmarshal(req.Body, &data); err != nil {
			logger.Error(c.Ctx(), "ProxyMessageHandler proto.Unmarshal err:%v", err)
			_ = c.JsonErr(err)
			return
		}
		if err := m.loginHandler(c.Ctx(), &data); err != nil {
			logger.Error(c.Ctx(), "loginHandler err:%v", err)
			_ = c.JsonErr(err)
			return
		}
	case model.ReqTables:
		var data pb.ReqTables
		if err := proto.Unmarshal(req.Body, &data); err != nil {
			logger.Error(c.Ctx(), "ProxyMessageHandler proto.Unmarshal err:%v", err)
			_ = c.JsonErr(err)
			return
		}
		if err := m.tableListHandler(c.Ctx(), &data); err != nil {
			logger.Error(c.Ctx(), "tableListHandler err:%v", err)
			_ = c.JsonErr(err)
			return
		}
	case model.ReqRankList:
		var data pb.ReqRanks
		if err := proto.Unmarshal(req.Body, &data); err != nil {
			logger.Error(c.Ctx(), "ProxyMessageHandler proto.Unmarshal err:%v", err)
			_ = c.JsonErr(err)
			return
		}
		if err := m.rankListHandler(c.Ctx(), &data); err != nil {
			logger.Error(c.Ctx(), "tableListHandler err:%v", err)
			_ = c.JsonErr(err)
			return
		}
	default:
		_ = c.JsonErr(code.MessageIllegalError)
		logger.Error(c.Ctx(), "ProxyMessageHandler messageID not find, id:%d", req.MessageID)
		return
	}
}

func (m *Manager) tableListHandler(ctx context.Context, req *pb.ReqTables) error {
	var tag string
	if req.PtCode == "" && req.Currency == "" {
		tag = "common"
	} else {
		tag = fmt.Sprintf("%s_%s", req.PtCode, req.Currency)
	}

	var tables []*pb.Table
	for _, rc := range conf.ExtraConfig().Channels {
		channelTag := fmt.Sprintf("%s_%s", rc.ChannelCode, rc.Currency)
		if strings.Contains(channelTag, tag) {
			for _, room := range rc.Rooms {
				tables = append(tables, &pb.Table{
					TableId:    room.RoomID,
					RoomName:   room.RoomName,
					MinAccount: room.MinAccount,
					MaxAccount: room.MaxAccount,
					Fee:        room.Fee,
					BaseScore:  room.BaseScore,
				})
			}
		}
	}

	// 做成应答
	msg := pb.AckTables{
		UserId: req.UserId,
		Tables: tables,
	}
	_ = utils.PushMessage(ctx, model.AckTables, &msg, gateway.PushModeSpecifyTo, "", nil, req.UserId)

	return nil
}

func (m *Manager) rankListHandler(ctx context.Context, req *pb.ReqRanks) error {

	var ranks []*pb.Rank
	userIds, err := m.dao.GetRank(ctx, req.ChannelCode)
	if err != nil {
		ranks = nil
	}
	if len(userIds) > 0 {
		var userIdsInt []int64
		for _, userStr := range userIds {
			uid, err := strconv.Atoi(userStr)
			if err != nil {
				userIdsInt = append(userIdsInt, int64(uid))
			}
		}
		if len(userIds) > 0 {
			userInfos, err := user.GetUserInfos(ctx, userIdsInt)
			if err != nil {
				return err
			}
			for key, userInfo := range userInfos {
				rankInfo := new(pb.Rank)
				rankInfo.Rank = int64(key) + 1
				rankInfo.Avatar = userInfo.Avatar
				rankInfo.UserId = userInfo.UserID
				rankInfo.Nickname = userInfo.Nickname
				uWinScore, err := m.dao.GetUserWinCoin(ctx, req.ChannelCode, userInfo.UserID)
				if err != nil {
					uWinScore = 0
				}
				rankInfo.Score = uWinScore
			}
		}

	}
	// 做成应答
	msg := pb.AckRanks{
		UserId:      req.UserId,
		ChannelCode: req.ChannelCode,
		Ranks:       ranks,
	}
	_ = utils.PushMessage(ctx, model.AckRankList, &msg, gateway.PushModeSpecifyTo, "", nil, req.UserId)

	return nil
}

func (m *Manager) loginHandler(ctx context.Context, req *pb.ReqLogin) error {
	// 缓存检查玩家是否在桌子中
	var UserRoom string
	for tableId, table := range m.Tables {
		for userId, _ := range table.TableClients {
			if int64(userId) == req.UserId {
				UserRoom = tableId
			}
		}
	}
	// 做成应答
	msg := pb.AckLogin{
		UserId:       req.UserId,
		PlayingTable: UserRoom,
	}
	// 做成pb bytes
	body, err := proto.Marshal(&msg)
	if err != nil {
		logger.Error(ctx, "loginHandler proto.Marshal err:%v", err)
		return err
	}
	// 推送给客户端
	pushReq := gateway.PushMessageRequest{
		App:  conf.ExtraConfig().AppID,
		Mode: gateway.PushModeSpecifyTo,
		To:   []int64{req.UserId},
		Data: struct {
			MessageID int32  `json:"messageID"`
			Body      []byte `json:"body"`
		}{
			MessageID: model.AckLogin,
			Body:      body,
		},
	}
	if err = gateway.PushMessage(ctx, &pushReq); err != nil {
		logger.Error(ctx, "loginHandler gateway.PushMessage err:%v, data:%v", err, req)
		return err
	}
	return nil
}
