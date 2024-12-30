package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type GroupConsumer struct {
	ctx        context.Context
	autoCommit bool
	builder    IMessageBuilder
	group      sarama.ConsumerGroup
	kafkaVer   sarama.KafkaVersion
	msgChan    chan *ConsumerMessage
	errChan    chan error
}

func (gc *GroupConsumer) Setup(sess sarama.ConsumerGroupSession) error {
	//fmt.Println("setup")
	//fmt.Println(sess.Claims())
	return nil
}

func (gc *GroupConsumer) Cleanup(sess sarama.ConsumerGroupSession) error {
	//fmt.Println("cleanup")
	//TODO 再平衡时也会调用，这个先直接去掉，不能直接关闭
	//gc.group.Close()
	return nil
}

func (gc *GroupConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			cm := gc.builder.ConsumerMessage(gc.ctx, msg, gc.kafkaVer)
			// 这里阻塞写入chan 因此为了效率 消费者应该异步处理
			cm.sess = sess
			gc.msgChan <- cm
			if gc.autoCommit {
				// 标记
				// sarama会自动进行提交 默认间隔1秒
				sess.MarkMessage(msg, "")
			}
		case <-sess.Context().Done():
			// 关闭时进行提交
			sess.Commit()
			return nil
		}
	}
}

func (gc *GroupConsumer) Message() chan *ConsumerMessage {
	return gc.msgChan
}

func (gc *GroupConsumer) Error() chan error {
	return gc.errChan
}

func (gc *GroupConsumer) start(topics []string, handler sarama.ConsumerGroupHandler) {
	go func() {
		for {
			err := gc.group.Consume(gc.ctx, topics, handler)
			if err != nil {
				switch err {
				case sarama.ErrClosedClient, sarama.ErrClosedConsumerGroup:
					// 退出
					return
				default:
					gc.errChan <- err
				}
			}
		}
	}()
	go func() {
		defer gc.group.Close()
		for {
			select {
			case err := <-gc.group.Errors():
				gc.errChan <- err
			case <-gc.ctx.Done():
				return
			}
		}
	}()
}

func NewGroupConsumer(
	ctx context.Context,
	cfg *sarama.Config,
	groupName string,
	brokers []string,
	topicCreateEnable bool,
	autoCommit bool,
	builder IMessageBuilder,
	topics string) (*GroupConsumer, error) {
	// 初始化client
	c, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return nil, err
	}
	// topic检查
	if !topicCreateEnable {
		partitionTopics, err := c.Topics()
		if err != nil {
			return nil, err
		}
		consumerTopics := strings.Split(topics, ",")
		// 开始检查
		var needCreate []string
		for _, ct := range consumerTopics {
			has := false
			for _, pt := range partitionTopics {
				if ct != pt {
					continue
				}
				has = true
				break
			}
			if !has {
				needCreate = append(needCreate, ct)
			}
		}
		// 判断结果
		if len(needCreate) > 0 {
			return nil, fmt.Errorf("kafka topic not find, topics:%s", strings.Join(needCreate, ","))
		}
	}
	// 根据client创建consumerGroup
	group, err := sarama.NewConsumerGroupFromClient(groupName, c)
	if err != nil {
		return nil, err
	}
	gc := GroupConsumer{
		ctx:        ctx,
		autoCommit: autoCommit,
		builder:    builder,
		group:      group,
		kafkaVer:   cfg.Version,
		msgChan:    make(chan *ConsumerMessage),
		errChan:    make(chan error),
	}
	gc.start(strings.Split(topics, ","), &gc)
	return &gc, nil
}
