package conf

import (
	dkafka "code.cd.local/games-go/mini/rummy.busi/client/kafka"
	"code.cd.local/games-go/mini/rummy.busi/global"
	"code.cd.local/sagittarius/sagittarius-golang/app"
	"code.cd.local/sagittarius/sagittarius-golang/app/config"
	"code.cd.local/sagittarius/sagittarius-golang/logger"
	"context"
)

type Config struct {
	config.ServiceConfig
	AppID    int32 `json:"appID"`
	Channels []struct {
		ChannelID   int    `json:"channelID"`
		ChannelCode string `json:"channelCode"`
		Currency    string `json:"currency"`
		Rooms       []struct {
			RoomID     int32  `json:"roomID"`
			RoomName   string `json:"roomName"`
			MinAccount int64  `json:"min_account"`
			MaxAccount int64  `json:"max_account"`
			BaseScore  int64  `json:"base_score"`
			Fee        int64  `json:"fee"`
		} `json:"rooms"`
	} `json:"channels"`
	KafkaProducers []*KafkaProducers `json:"kafkaProducers"`
	Online         []OnlineItem      `json:"online"`
}

type KafkaProducers struct {
	Name          string `json:"name"`
	Brokers       string `json:"brokers"`
	Topics        string `json:"topics"`
	Group         string `json:"group"`
	OffsetInitial int64  `json:"offsetInitial"`
}
type OnlineItem struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Min   int `json:"min"`
	Max   int `json:"max"`
}

var conf *Config

func InitAppConfig() {
	// 初始化配置
	conf = new(Config)
	err := app.Router().ExtraJsonConfig(conf)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case <-app.Router().Ctx().Done():
				return
			case <-app.Router().OnExtraConfigChange():
				var cfg Config
				if err = app.Router().ExtraJsonConfig(&cfg); err != nil {
					logger.Error(app.Router().Ctx(), "nacos config reload error:%v", err)
					continue
				}
				logger.Debug(app.Router().Ctx(), "nocos reload cfg:%v", cfg)
				conf = &cfg
			}
		}
	}()
}
func InitKafka() {
	var err error
	ctx := context.Background()
	global.GnKafkaPro, err = dkafka.ProducerInit(ctx)
	if err != nil {
		logger.Error(ctx, "InitKafka ProducerInit, err:%v", err)
	}
}
func ExtraConfig() *Config {
	return conf
}
