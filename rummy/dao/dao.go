package dao

import (
	"context"
	"strconv"
	"time"

	"code.cd.local/games-go/mini/rummy.busi/dao/db"
	"code.cd.local/games-go/mini/rummy.busi/dao/model"
	sudRds "code.cd.local/games-go/mini/rummy.busi/dao/redis"

	"github.com/curry-mz/sagittarius-golang/redis"
)

type SudDao struct {
	rds *sudRds.SudRedis
	db  *db.SudDB
}

func NewDao() (*SudDao, error) {
	racingRedis, err := sudRds.New()
	if err != nil {
		return nil, err
	}
	db, err := db.New()
	if err != nil {
		return nil, err
	}
	return &SudDao{rds: racingRedis, db: db}, nil
}

func (dao *SudDao) CreateMutex(name string, expire time.Duration) *redis.Mutex {
	return dao.rds.CreateMutex(name, expire)
}

func (dao *SudDao) CreateMutexWithExtend(name string, expire time.Duration) *redis.Mutex {
	return dao.rds.CreateMutexWithExtend(name, expire)
}

func (dao *SudDao) GetStock(ctx context.Context, roomID int32) (*model.StockPO, error) {
	return dao.db.GetStock(ctx, roomID)
}

func (dao *SudDao) IncrStock(ctx context.Context, roomID int32, income, chips int64) error {
	return dao.db.IncrStock(ctx, roomID, income, chips)
}

func (dao *SudDao) CreateLotteryResult(ctx context.Context, po *model.LotteryPO) error {
	return dao.db.CreateLotteryResult(ctx, po)
}

func (dao *SudDao) UpdateLotteryResult(ctx context.Context,
	roomID int32, roundNo string, attr map[string]interface{}) error {
	if len(attr) == 0 {
		return nil
	}
	return dao.db.UpdateLotteryResult(ctx, roomID, roundNo, attr)
}

func (dao *SudDao) PageFindLotteryResult(ctx context.Context,
	roomID int32, pageNo int, size int) ([]*model.LotteryPO, error) {
	return dao.db.PageFindLotteryResult(ctx, roomID, pageNo, size)
}

func (dao *SudDao) FindLotteryResult(ctx context.Context, roomID int32, roundNo string) (*model.LotteryPO, error) {
	return dao.db.FindLotteryResult(ctx, roomID, roundNo)
}

func (dao *SudDao) CreateOrder(ctx context.Context,
	orderNo string,
	playerID int64,
	roomID int32,
	roundNo string,
	itemID int64,
	chips int64,
	currency string,
	timestamp int64) error {
	//if err := dao.rds.DelPlayerRoundBet(ctx, roomID, roundNo, playerID); err != nil {
	//	return err
	//}
	return dao.db.CreateOrder(ctx, orderNo, playerID, roomID, roundNo, itemID, chips, currency, timestamp)
}

func (dao *SudDao) GetFailedOrder(ctx context.Context) ([]*model.OrderPO, error) {
	return dao.db.GetFailedOrder(ctx)
}

func (dao *SudDao) CreateBetHistory(ctx context.Context, pos []*model.HistoryPO) error {
	if len(pos) == 0 {
		return nil
	}
	return dao.db.CreateBetHistory(ctx, pos)
}

func (dao *SudDao) GetRank(ctx context.Context, channelCode string) ([]string, error) {
	return dao.rds.GetRank(ctx, channelCode)
}
func (dao *SudDao) AddRank(ctx context.Context, channelCode string, uesrId int64, score int64) error {
	return dao.rds.ZAddUserRankScore(ctx, channelCode, uesrId, score)
}
func (dao *SudDao) GetRankMinOne(ctx context.Context, channelCode string) ([]string, error) {
	return dao.rds.GetRankMinOne(ctx, channelCode)
}
func (dao *SudDao) GetUserWinCoin(ctx context.Context, channelCode string, userId int64) (int64, error) {
	winCoin := dao.rds.GetUserWinCoin(ctx, channelCode, userId)
	winCoinInt, err := strconv.Atoi(winCoin)
	return int64(winCoinInt), err
}
func (dao *SudDao) SetUserWinCoin(ctx context.Context, channelCode string, userId int64, score int64) error {
	return dao.rds.SetUserWinCoin(ctx, channelCode, userId, score)
}
func (dao *SudDao) DeleteUserRank(ctx context.Context, channelCode string, userId int64) error {
	return dao.rds.ZDeleteUserRankScore(ctx, channelCode, userId)
}
func (dao *SudDao) MatchGame(ctx context.Context, channelCode string, roomId int64, uesrId int64, score int64) error {
	return dao.rds.ZAddMatch(ctx, channelCode, roomId, uesrId, score)
}
