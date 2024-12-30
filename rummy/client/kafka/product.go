package kafka

import (
	"code.cd.local/sagittarius/sagittarius-golang/app/proxy"
	"code.cd.local/sagittarius/sagittarius-golang/logger"
	"code.cd.local/sagittarius/sagittarius-golang/mq/kafka"
	"context"
)

const Name = "zues.admin"

func ProducerInit(ctx context.Context) (*kafka.Producer, error) {
	producer, err := proxy.InitKafkaProducer(Name)
	if err != nil {
		logger.Error(ctx, "consumerKafka,err: %s", err.Error())
		return nil, err
	}
	return producer, nil
}
