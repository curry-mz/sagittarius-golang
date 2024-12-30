package core

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

/////////////////////////////////////////
// 生产者相关
/////////////////////////////////////////

var (
	_asyncChanSize = 256
)

// 生产者接口

type IProducer interface {
	SendMessage(ctx context.Context, topic string, key []byte, data []byte)
	Success() chan *sarama.ProducerMessage
	Error() chan *sarama.ProducerError
}

// 同步生产者相关

type SyncProducer struct {
	ctx      context.Context
	builder  IMessageBuilder
	sp       sarama.SyncProducer
	kafkaVer sarama.KafkaVersion
	errChan  chan *sarama.ProducerError
	succChan chan *sarama.ProducerMessage
}

func NewSyncProducer(ctx context.Context, brokers []string, builder IMessageBuilder, cfg *sarama.Config) (*SyncProducer, error) {
	// 初始化同步生产者
	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new sync producer")
	}
	sp := SyncProducer{
		ctx:      ctx,
		sp:       p,
		kafkaVer: cfg.Version,
		builder:  builder,
		errChan:  make(chan *sarama.ProducerError, 1),
		succChan: make(chan *sarama.ProducerMessage, 1),
	}
	go func(producer *SyncProducer) {
		select {
		case <-producer.ctx.Done():
			producer.sp.Close()
		}
	}(&sp)
	return &sp, nil
}

func (sp *SyncProducer) SendMessage(ctx context.Context, topic string, key []byte, data []byte) {
	// 构建消息
	msg := sp.builder.ProducerMessage(ctx, topic, key, data, sp.kafkaVer)
	_, _, err := sp.sp.SendMessage(msg.msg)
	if err != nil {
		sp.errChan <- &sarama.ProducerError{Msg: msg.msg, Err: err}
	} else {
		sp.succChan <- msg.msg
	}
}

func (sp *SyncProducer) Success() chan *sarama.ProducerMessage {
	return sp.succChan
}

func (sp *SyncProducer) Error() chan *sarama.ProducerError {
	return sp.errChan
}

// 异步生产者

type AsyncProducer struct {
	ctx      context.Context
	builder  IMessageBuilder
	ap       sarama.AsyncProducer
	kafkaVer sarama.KafkaVersion
	errChan  chan *sarama.ProducerError
	succChan chan *sarama.ProducerMessage
}

func NewAsyncProducer(ctx context.Context, brokers []string, builder IMessageBuilder, cfg *sarama.Config) (*AsyncProducer, error) {
	// 初始化异步生产者
	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new async producer")
	}
	ap := AsyncProducer{
		ctx:      ctx,
		builder:  builder,
		ap:       p,
		kafkaVer: cfg.Version,
		errChan:  make(chan *sarama.ProducerError, _asyncChanSize),
		succChan: make(chan *sarama.ProducerMessage, _asyncChanSize),
	}
	// 必要监听
	go func(producer *AsyncProducer) {
		for {
			select {
			case e := <-producer.ap.Errors():
				producer.errChan <- e
			case msg := <-producer.ap.Successes():
				producer.succChan <- msg
			case <-producer.ctx.Done():
				producer.ap.Close()
				return
			}
		}
	}(&ap)
	return &ap, nil
}

func (ap *AsyncProducer) SendMessage(ctx context.Context, topic string, key []byte, data []byte) {
	// 构建消息
	msg := ap.builder.ProducerMessage(ctx, topic, key, data, ap.kafkaVer)
	ap.ap.Input() <- msg.msg
}

func (ap *AsyncProducer) Success() chan *sarama.ProducerMessage {
	return ap.succChan
}

func (ap *AsyncProducer) Error() chan *sarama.ProducerError {
	return ap.errChan
}
