package redis

import (
	"code.cd.local/games-go/mini/rummy.busi/conf"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.cd.local/games-go/mini/rummy.busi/dao/model"

	"code.cd.local/sagittarius/sagittarius-golang/app/proxy"
	"code.cd.local/sagittarius/sagittarius-golang/redis"

	redisgo "github.com/go-redis/redis/v8"
)

type SudRedis struct {
	rds *redis.Client
}

func New() (*SudRedis, error) {
	cli, err := proxy.InitRedisClient(conf.ExtraConfig().Rds[0].Name)
	if err != nil {
		return nil, err
	}
	return &SudRedis{rds: cli}, nil
}

func (rr *SudRedis) CreateMutex(name string, expire time.Duration) *redis.Mutex {
	return rr.rds.NewMutex(name, expire)
}

func (rr *SudRedis) CreateMutexWithExtend(name string, expire time.Duration) *redis.Mutex {
	return rr.rds.NewMutexWithExtend(name, expire)
}

func (rr *SudRedis) GetRoundPoints(ctx context.Context, roomID int32, roundNo string) ([]int32, error) {
	key := fmt.Sprintf(model.RoundPointsKey, roomID, roundNo)
	reply, err := rr.rds.Get(ctx, key).Result()
	if err != nil {
		if err == redisgo.Nil {
			return nil, nil
		}
		return nil, err
	}
	if reply == "" {
		return nil, nil
	}
	ss := strings.Split(reply, ",")
	var points []int32
	for _, s := range ss {
		v, rerr := strconv.Atoi(s)
		if rerr != nil {
			return nil, rerr
		}
		points = append(points, int32(v))
	}
	return points, nil
}

func (rr *SudRedis) GetRank(ctx context.Context, channelCode string) (i []string, err error) {
	key := fmt.Sprintf(model.ChannelWinRankKey, channelCode)
	return rr.rds.ZRevRangeByScore(ctx, key, &redisgo.ZRangeBy{
		"-inf",
		"+inf",
		0,
		100,
	}).Result()
}
func (rr *SudRedis) ZAddUserRankScore(ctx context.Context, channelCode string, userId int64, score int64) error {
	key := fmt.Sprintf(model.ChannelWinRankKey, channelCode)
	return rr.rds.ZAdd(ctx, key, &redisgo.Z{
		Score:  float64(score),
		Member: userId,
	}).Err()
}
func (rr *SudRedis) GetRankMinOne(ctx context.Context, channelCode string) (i []string, err error) {
	key := fmt.Sprintf(model.ChannelWinRankKey, channelCode)
	return rr.rds.ZRangeByScore(ctx, key, &redisgo.ZRangeBy{
		"-inf",
		"+inf",
		0,
		1,
	}).Result()
}

func (rr *SudRedis) GetUserWinCoin(ctx context.Context, channelCode string, userId int64) string {
	key := fmt.Sprintf(model.UserWinCoinKey, channelCode, userId)
	return rr.rds.Get(ctx, key).Val()
}

func (rr *SudRedis) SetUserWinCoin(ctx context.Context, channelCode string, userId int64, score int64) error {
	key := fmt.Sprintf(model.UserWinCoinKey, channelCode, userId)
	return rr.rds.Set(ctx, key, score, -1).Err()
}
func (rr *SudRedis) ZDeleteUserRankScore(ctx context.Context, channelCode string, userId int64) error {
	key := fmt.Sprintf(model.ChannelWinRankKey, channelCode)
	return rr.rds.ZRem(ctx, key, userId).Err()
}

func (rr *SudRedis) ZAddMatch(ctx context.Context, channelCode string, roomId int64, userId int64, score int64) error {
	key := fmt.Sprintf(model.ChannelRoomMatchKey, channelCode, roomId)
	return rr.rds.ZAdd(ctx, key, &redisgo.Z{
		Score:  float64(score),
		Member: userId,
	}).Err()
}
