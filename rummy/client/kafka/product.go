package kafka

import (
	"context"
	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/logger"
	"github.com/curry-mz/sagittarius-golang/mq/kafka"
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
