package utils

import (
	"code.cd.local/games-go/mini/rummy.busi/client/gateway"
	"code.cd.local/games-go/mini/rummy.busi/conf"
	"context"
	"github.com/curry-mz/sagittarius-golang/logger"
	"google.golang.org/protobuf/proto"
)

func PushMessage(ctx context.Context, msgID int32, msg proto.Message, mode int, group string, except []int64, to ...int64) error {
	// 做成pb bytes
	body, err := proto.Marshal(msg)
	if err != nil {
		logger.Error(ctx, "pushMessage proto.Marshal err:%v", err)
		return err
	}

	// 推送给客户端
	req := gateway.PushMessageRequest{
		App:    conf.ExtraConfig().AppID,
		Mode:   mode,
		Group:  group,
		To:     to,
		Except: except,
		Data: struct {
			MessageID int32  `json:"messageID"`
			Body      []byte `json:"body"`
		}{
			MessageID: msgID,
			Body:      body,
		},
	}
	if err = gateway.PushMessage(ctx, &req); err != nil {
		logger.Error(ctx, "pushMessage gateway.PushMessage err:%v, data:%v", err, req)
		return err
	}

	return nil
}
