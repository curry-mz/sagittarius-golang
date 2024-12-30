package game

import (
	"code.cd.local/games-go/mini/rummy.busi/client/finance"
	"code.cd.local/games-go/mini/rummy.busi/code"
	"code.cd.local/games-go/mini/rummy.busi/conf"
	"code.cd.local/games-go/mini/rummy.busi/dao"
	"code.cd.local/games-go/mini/rummy.busi/utils"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"time"
)

type timeRange struct {
	Start time.Duration // 起始时间
	End   time.Duration // 结束时间
	Min   int           // 最小随机数
	Max   int           // 最大随机数
}
type Room struct {
	Ctx        context.Context
	RoomID     int32
	RoomName   string
	MinAccount int64
	MaxAccount int64
	BaseScore  int64
	Fee        int64
	dao        *dao.SudDao
}
type BetLimitsRate struct {
	Limit  int64
	Weight int64
}

func NewRoom(ctx context.Context, roomID int32, roomName string, dao *dao.SudDao, min, max, fee, baseScore int64) (*Room, error) {
	r := &Room{
		Ctx:        ctx,
		RoomID:     roomID,
		RoomName:   roomName,
		dao:        dao,
		Fee:        fee,
		MinAccount: min,
		MaxAccount: max,
		BaseScore:  baseScore,
	}
	return r, nil
}

func (r *Room) ID() int32 {
	return r.RoomID
}

func (r *Room) RoomIDStr() string {
	return strconv.Itoa(int(r.ID()))
}

func (r *Room) InsertDayRank(channelId string, userId int64, winCoin int64) (err error) {
	userIds, err := r.dao.GetRank(r.Ctx, channelId)
	if err != nil {

		return err
	}
	// 获取今天已经赢的钱
	uWinCoin, err := r.dao.GetUserWinCoin(r.Ctx, channelId, userId)
	if err != nil {
		uWinCoin = 0
	}

	totalWinCoin := uWinCoin + winCoin
	r.dao.SetUserWinCoin(r.Ctx, channelId, userId, totalWinCoin)
	if len(userIds) < 100 {
		// 小于100，直接插入排行榜
		r.dao.AddRank(r.Ctx, channelId, userId, totalWinCoin)
		return
	}
	// 获取排行榜最小的人
	minUser, err := r.dao.GetRankMinOne(r.Ctx, channelId)
	if len(minUser) != 0 {
		// 获取今天已经赢的钱
		if minUser[0] != "" {
			minUid, err := strconv.Atoi(minUser[0])
			if err != nil {
				return err
			}
			minUserCoin, err := r.dao.GetUserWinCoin(r.Ctx, channelId, userId)
			if err != nil {
				return err
			}
			if totalWinCoin > minUserCoin || minUid == int(userId) {
				// 删除最后一名
				if !utils.InArrayString(strconv.Itoa(int(userId)), userIds) {
					r.dao.DeleteUserRank(r.Ctx, channelId, int64(minUid))
				}
				// 插入加入排行榜的人
				r.dao.AddRank(r.Ctx, channelId, userId, totalWinCoin)
			}
		}
	}
	return
}

func (r *Room) GetRoomNum() int64 {
	trs := []*timeRange{}
	online := conf.ExtraConfig().Online
	for _, tr := range online {
		nTimeRange := new(timeRange)
		nTimeRange.Start = time.Duration(tr.Start) * time.Hour
		nTimeRange.End = time.Duration(tr.End) * time.Hour
		nTimeRange.Max = tr.Max
		nTimeRange.Min = tr.Min
		trs = append(trs, nTimeRange)
	}
	hourOfDay := time.Now().Hour()

	// 根据当前时间确定时间段
	min := 0
	max := 0
	for _, tr := range trs {
		if tr.Start.Hours() <= float64(hourOfDay) && float64(hourOfDay) < tr.End.Hours() {
			min = tr.Min
			max = tr.Max
			break
		}
	}

	if max == 0 && min == 0 {
		fmt.Println("No matching time range found for current time.")
		return 0
	}

	// 生成随机数
	rand.Seed(time.Now().UnixNano())
	userNum := int64(rand.Intn(max-min+1) + min)
	return userNum
}
func (r *Room) MatchGame(channelId string, roomId int64, userId int64) (err error) {
	pcMap, err := finance.Query(context.Background(), &finance.QueryRequest{
		PlayerIDs: []int64{userId},
	})
	if err != nil {
		return errors.WithMessage(code.ServerError, "|finance.Query")
	}
	chips := pcMap[userId]
	if chips < r.MinAccount {
		return errors.WithMessage(code.AmountInsufficientError, "|amount insufficient")
	}
	timeStamp := time.Now().UnixNano()
	err = r.dao.MatchGame(r.Ctx, channelId, roomId, userId, timeStamp)
	return
}
