package db

import (
	"code.cd.local/games-go/mini/rummy.busi/conf"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"code.cd.local/games-go/mini/rummy.busi/dao/model"
	"code.cd.local/games-go/mini/rummy.busi/utils"

	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/logger"
	"github.com/curry-mz/sagittarius-golang/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SudDB struct {
	sql *mysql.Client
}

func New() (*SudDB, error) {
	cli, err := proxy.InitSqlClient(conf.ExtraConfig().Databases[0].Name)
	if err != nil {
		return nil, err
	}
	return &SudDB{sql: cli}, nil
}

func (db *SudDB) getOrderTableName() string {
	now := time.Now()
	return fmt.Sprintf(model.OrderTable, now.Year(), now.Month())
}

func (db *SudDB) getHistoryTableName() string {
	now := time.Now()
	return fmt.Sprintf(model.HistoryTable, now.Year(), now.Month())
}

func (db *SudDB) GetStock(ctx context.Context, roomID int32) (*model.StockPO, error) {
	var po model.StockPO
	result := db.sql.DB(ctx).Table(model.StockTable).
		Where("`room_id` = ?", roomID).
		Last(&po)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound && result.Error != sql.ErrNoRows {
			return nil, result.Error
		}
		return nil, nil
	}
	return &po, nil
}

func (db *SudDB) IncrStock(ctx context.Context, roomID int32, income, chips int64) error {
	po := model.StockPO{
		RoomID:      roomID,
		TotalIncome: income,
		TotalAccept: chips,
	}
	return db.sql.DB(ctx).Table(model.StockTable).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "room_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_accept": gorm.Expr("`total_accept` + ?", chips),
			"total_income": gorm.Expr("`total_income` + ?", income),
		}),
	}).Create(&po).Error
}

func (db *SudDB) CreateLotteryResult(ctx context.Context, po *model.LotteryPO) error {
	return db.sql.DB(ctx).Table(model.LotteryTable).Create(po).Error
}

func (db *SudDB) UpdateLotteryResult(ctx context.Context,
	roomID int32, roundNo string, attr map[string]interface{}) error {
	return db.sql.DB(ctx).Table(model.LotteryTable).
		Where("`room_id` = ? and `round_no` = ?", roomID, roundNo).
		Updates(attr).Error
}

func (db *SudDB) PageFindLotteryResult(ctx context.Context,
	roomID int32, pageNo int, size int) ([]*model.LotteryPO, error) {
	var pos []*model.LotteryPO
	result := db.sql.DB(ctx).Table(model.LotteryTable).
		Where("`room_id` = ?", roomID).
		Order("`id` desc").
		Limit(size).
		Offset((pageNo - 1) * size).
		Find(&pos)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound && result.Error != sql.ErrNoRows {
			return nil, result.Error
		}
		return nil, nil
	}
	return pos, nil
}

func (db *SudDB) FindLotteryResult(ctx context.Context, roomID int32, roundNo string) (*model.LotteryPO, error) {
	var po model.LotteryPO
	result := db.sql.DB(ctx).Table(model.LotteryTable).Where("`room_id` = ? and `round_no` = ?", roomID, roundNo).Find(&po)
	if result.Error != nil {
		return nil, result.Error
	}
	return &po, nil
}

func (db *SudDB) CreateOrder(ctx context.Context,
	orderNo string,
	playerID int64,
	roomID int32,
	roundNo string,
	itemID int64,
	chips int64,
	currency string,
	timestamp int64) error {
	po := model.OrderPO{
		OrderNo:   orderNo,
		PlayerID:  playerID,
		RoomID:    roomID,
		RoundNo:   roundNo,
		ItemID:    itemID,
		BetChips:  chips,
		Currency:  currency,
		Prize:     0,
		Timestamp: timestamp,
	}
	return db.sql.DB(ctx).Table(db.getOrderTableName()).Create(&po).Error
}

func (db *SudDB) GetRoundOrder(ctx context.Context, roomID int32, roundNo string) ([]*model.OrderPO, error) {
	var pos []*model.OrderPO
	tableName := db.getOrderTableName()
	result := db.sql.DB(ctx).Table(tableName).
		Where("`room_id` = ? and `round_no` = ?", roomID, roundNo).
		Order("`id` desc").
		Find(&pos)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return pos, nil
}

func (db *SudDB) GetFailedOrder(ctx context.Context) ([]*model.OrderPO, error) {
	var pos []*model.OrderPO
	tableName := db.getOrderTableName()
	result := db.sql.DB(ctx).Table(tableName).
		Where("`status` = ?", model.OrderStateSettleFailed).
		Order("`id` desc").
		Find(&pos)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return pos, nil
}

func (db *SudDB) GetOrderWithPlayersByRound(ctx context.Context,
	roomID int32, roundNo string, playerIDs []int64) ([]*model.OrderPO, error) {
	var pos []*model.OrderPO
	tableName := db.getOrderTableName()
	result := db.sql.DB(ctx).Table(tableName).
		Where("`room_id` = ? and `round_no` = ? and `player_id` in ?", roomID, roundNo, playerIDs).
		Order("`id` desc").
		Find(&pos)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return pos, nil
}

func (db *SudDB) MUpdateOrder(ctx context.Context, pos []*model.OrderPO, updateColumn []string) error {
	tableName := db.getOrderTableName()
	querys, err := utils.GenBatchUpdateSQLWithPK(tableName, pos, "id", updateColumn)
	if err != nil {
		return err
	}
	var errors []error
	var wg sync.WaitGroup
	for _, query := range querys {
		wg.Add(1)
		go func(qr string) {
			defer wg.Done()

			result := db.sql.DB(ctx).Exec(qr)
			if result.Error != nil {
				if result.Error != gorm.ErrRecordNotFound && result.Error != sql.ErrNoRows {
					errors = append(errors, result.Error)
					logger.Error(ctx, "MUpdateOrder, query:%s, err:%v", qr, result.Error)
				}
			}
		}(query)
	}
	wg.Wait()
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (db *SudDB) CreateBetHistory(ctx context.Context, pos []*model.HistoryPO) error {
	tableName := db.getHistoryTableName()
	return db.sql.DB(ctx).Table(tableName).Create(&pos).Error
}

func (db *SudDB) PageFindBetHistory(ctx context.Context,
	roomID int32, playerID int64, offset int64) ([]*model.HistoryPO, error) {
	var pos []*model.HistoryPO
	tableName := db.getHistoryTableName()
	var result *gorm.DB
	if offset == 0 {
		result = db.sql.DB(ctx).Table(tableName).
			Where("`room_id` = ? and `player_id` = ?", roomID, playerID).
			Order("`id` desc").
			Limit(10).
			Find(&pos)
	} else {
		result = db.sql.DB(ctx).Table(tableName).
			Where("`id` < ? and `room_id` = ? and `player_id` = ?", offset, roomID, playerID).
			Order("`id` desc").
			Limit(10).
			Find(&pos)
	}
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return pos, nil
}
func (db *SudDB) UpdateStock(ctx context.Context,
	roomID int32, attr map[string]interface{}) error {
	return db.sql.DB(ctx).Table(model.StockTable).
		Where("`room_id` = ?", roomID).
		Updates(attr).Error
}
func (db *SudDB) PageFindBetHistoryByUsser(ctx context.Context, playerID int64, offset int64) ([]*model.UserHistoryPO, error) {
	var pos []*model.UserHistoryPO
	tableName := db.getHistoryTableName()
	var result *gorm.DB
	joinSql := fmt.Sprintf("left join lottery_result on lottery_result.round_no = %s.round_no and lottery_result.room_id = %s.room_id", tableName, tableName)
	selectField := fmt.Sprintf("%s.*,lottery_result.winner,lottery_result.red_cards,lottery_result.red_type,lottery_result.black_cards,lottery_result.black_type,lottery_result.settle_timestamp", tableName)
	if offset == 0 {
		result = db.sql.DB(ctx).Table(tableName).Joins(joinSql).
			Where("`player_id` = ?", playerID).
			Order("`id` desc").
			Select(selectField).
			Limit(100).
			Find(&pos)
	} else {
		result = db.sql.DB(ctx).Table(tableName).
			Where("`id` < ? and `player_id` = ?", offset, playerID).
			Order("`id` desc").
			Select(selectField).
			Limit(100).
			Find(&pos)
	}
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return pos, nil
}
