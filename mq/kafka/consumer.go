package kafka

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"

	"code.cd.local/sagittarius/sagittarius-golang/mq/kafka/core"
)

type ConsumerOption func(*consumerOption)

type consumerOption struct {
	reBalance         []string
	offsetInitial     int64
	commitRetry       int
	maxWaitTime       time.Duration
	topicCreateEnable bool // 是否允许自动创建不存在的topic 生产环境不建议使用true
	autoCommit        bool
	builder           core.IMessageBuilder
}

func ConsumerReBalance(reBalance []string) ConsumerOption {
	return func(o *consumerOption) {
		o.reBalance = reBalance
	}
}

func ConsumerOffsetInitial(offsetInitial int64) ConsumerOption {
	return func(o *consumerOption) {
		o.offsetInitial = offsetInitial
	}
}

func ConsumerMaxWaitTime(maxWaitTime time.Duration) ConsumerOption {
	return func(o *consumerOption) {
		o.maxWaitTime = maxWaitTime
	}
}

func ConsumerCommitRetry(commitRetry int) ConsumerOption {
	return func(o *consumerOption) {
		o.commitRetry = commitRetry
	}
}

func ConsumerTopicCreateEnable(topicCreateEnable bool) ConsumerOption {
	return func(o *consumerOption) {
		o.topicCreateEnable = topicCreateEnable
	}
}

func ConsumerAutoCommit(autoCommit bool) ConsumerOption {
	return func(o *consumerOption) {
		o.autoCommit = autoCommit
	}
}

func ConsumerMessageBuilder(builder core.IMessageBuilder) ConsumerOption {
	return func(o *consumerOption) {
		o.builder = builder
	}
}

type Consumer struct {
	gc      *core.GroupConsumer
	Version sarama.KafkaVersion
	msgChan chan *core.ConsumerMessage
	errChan chan error
}

func (c *Consumer) Message() chan *core.ConsumerMessage {
	return c.msgChan
}

func (c *Consumer) Error() chan error {
	return c.errChan
}

func NewConsumer(ctx context.Context, groupName string, brokers []string, topics string, opts ...ConsumerOption) (*Consumer, error) {
	cfg := sarama.NewConfig()
	// TODO 根据这个ID做唯一识别
	cfg.ClientID = groupName + "_" + uuid.New().String()
	// 设置sarama日志
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	// 设定版本
	cfg.Version = Version
	// 返回设置
	cfg.Consumer.Return.Errors = true
	// 其他设置
	option := consumerOption{
		offsetInitial:     sarama.OffsetNewest,
		commitRetry:       defaultConsumerCommitRetryTimes,
		reBalance:         []string{defaultConsumerRebalanceStrategy},
		maxWaitTime:       defaultConsumerMaxWaitTime,
		topicCreateEnable: false,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&option)
		}
	}
	if option.builder == nil {
		return nil, fmt.Errorf("kafka consumer, message builder is nil")
	}
	// 分区分配策略
	for _, rb := range option.reBalance {
		switch rb {
		case "range":
			cfg.Consumer.Group.Rebalance.GroupStrategies = append(cfg.Consumer.Group.Rebalance.GroupStrategies,
				sarama.NewBalanceStrategyRange())
		case "roundrobin":
			cfg.Consumer.Group.Rebalance.GroupStrategies = append(cfg.Consumer.Group.Rebalance.GroupStrategies,
				sarama.NewBalanceStrategyRoundRobin())
		default:
			cfg.Consumer.Group.Rebalance.GroupStrategies = append(cfg.Consumer.Group.Rebalance.GroupStrategies,
				sarama.NewBalanceStrategySticky())
		}
	}

	// 初始消费
	if option.offsetInitial != sarama.OffsetNewest && option.offsetInitial != sarama.OffsetOldest {
		option.offsetInitial = sarama.OffsetNewest
	}
	cfg.Consumer.Offsets.Retry.Max = option.commitRetry
	cfg.Consumer.Offsets.Initial = option.offsetInitial
	cfg.Consumer.MaxWaitTime = option.maxWaitTime
	cfg.Metadata.AllowAutoTopicCreation = option.topicCreateEnable
	if !option.autoCommit {
		cfg.Consumer.Offsets.AutoCommit.Enable = false
	}
	gc, err := core.NewGroupConsumer(ctx, cfg, groupName, brokers, option.topicCreateEnable,
		option.autoCommit, option.builder, topics)
	if err != nil {
		return nil, err
	}
	c := &Consumer{
		gc:      gc,
		Version: Version,
		msgChan: make(chan *core.ConsumerMessage),
		errChan: make(chan error),
	}
	go func() {
		for {
			select {
			case msg, ok := <-gc.Message():
				if ok {
					c.msgChan <- msg
				}
			case e := <-gc.Error():
				c.errChan <- e
			}
		}
	}()
	return c, nil
}
