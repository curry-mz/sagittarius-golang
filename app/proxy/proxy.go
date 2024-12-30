package proxy

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/curry-mz/sagittarius-golang/app"
	"github.com/curry-mz/sagittarius-golang/app/config"
	"github.com/curry-mz/sagittarius-golang/cores/client/http"
	"github.com/curry-mz/sagittarius-golang/cores/client/rpc"
	"github.com/curry-mz/sagittarius-golang/env"
	"github.com/curry-mz/sagittarius-golang/logger"
	"github.com/curry-mz/sagittarius-golang/mq/kafka"
	kfkCore "github.com/curry-mz/sagittarius-golang/mq/kafka/core"
	"github.com/curry-mz/sagittarius-golang/mq/rocket/consumer"
	"github.com/curry-mz/sagittarius-golang/mq/rocket/producer"
	"github.com/curry-mz/sagittarius-golang/mysql"
	"github.com/curry-mz/sagittarius-golang/redis"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	_sqlClient      = sync.Map{}
	_redisClient    = sync.Map{}
	_client         = sync.Map{}
	_rocketProducer = sync.Map{}
	_rocketConsumer = sync.Map{}
	_kafkaProducer  = sync.Map{}
	_kafkaConsumer  = sync.Map{}

	_sqlMutex            = sync.Mutex{}
	_redisMutex          = sync.Mutex{}
	_clientMutex         = sync.Mutex{}
	_rocketProducerMutex = sync.Mutex{}
	_rocketConsumerMutex = sync.Mutex{}
	_kafkaProducerMutex  = sync.Mutex{}
	_kafkaConsumerMutex  = sync.Mutex{}
)

// InitSqlClient 初始化mysql客户端
func InitSqlClient(name string, opts ...mysql.Option) (*mysql.Client, error) {
	_sqlMutex.Lock()
	defer _sqlMutex.Unlock()
	if c, has := _sqlClient.Load(name); has {
		return c.(*mysql.Client), nil
	}
	config := app.Router().Config().GetDatabase(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init sql client, config is nil, name:%s", name))
	}
	if config.Master == "" {
		return nil, errors.New(fmt.Sprintf("app init sql client, master is nil, name:%s", name))
	}
	if config.MaxIdle > 0 {
		opts = append(opts, mysql.MaxIdle(config.MaxIdle))
	}
	if config.MaxOpen > 0 {
		opts = append(opts, mysql.MaxOpen(config.MaxOpen))
	}
	if config.MaxLifeTime != "" {
		td, err := time.ParseDuration(config.MaxLifeTime)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("app init sql client, config maxlifetime, value:%s", config.MaxLifeTime))
		}
		opts = append(opts, mysql.MaxLifeTime(td))
	}
	if config.MaxIdleTime != "" {
		td, err := time.ParseDuration(config.MaxIdleTime)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("app init sql client, config maxidletime, value:%s", config.MaxIdleTime))
		}
		opts = append(opts, mysql.MaxLifeTime(td))
	}
	opts = append(opts, mysql.Logger(logger.GetLogger()))
	c, err := mysql.NewClient(config.Master, config.Slaves, opts...)
	if err != nil {
		return nil, err
	}
	_sqlClient.Store(name, c)
	return c, nil
}

// InitRedisClient 初始化redis客户端
func InitRedisClient(name string, opts ...redis.Option) (*redis.Client, error) {
	_redisMutex.Lock()
	defer _redisMutex.Unlock()
	if c, has := _redisClient.Load(name); has {
		return c.(*redis.Client), nil
	}
	config := app.Router().Config().GetRedis(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init redis client, config is nil, name:%s", name))
	}
	if config.Addr == "" {
		return nil, errors.New(fmt.Sprintf("app init redis client, addr is nil, name:%s", name))
	}
	addrs := strings.Split(config.Addr, ",")
	opts = append(opts, redis.Addrs(addrs), redis.Name(name))
	if config.Model != "" {
		opts = append(opts, redis.Model(config.Model))
	}
	if config.ReadTimeout > 0 {
		opts = append(opts, redis.ReadTimeout(config.ReadTimeout))
	}
	if config.WriteTimeout > 0 {
		opts = append(opts, redis.WriteTimeout(config.WriteTimeout))
	}
	if config.IdleTimeout > 0 {
		opts = append(opts, redis.IdleTimeout(config.IdleTimeout))
	}
	if config.DataBase >= 0 {
		opts = append(opts, redis.DB(config.DataBase))
	}
	if config.MaxRetry > 0 {
		opts = append(opts, redis.Retry(config.MaxRetry))
	}
	if config.PoolSize > 0 {
		opts = append(opts, redis.PoolSize(config.PoolSize))
	}
	if config.MinIdleConn > 0 {
		opts = append(opts, redis.MinIdleConn(config.MinIdleConn))
	}
	if config.Password != "" {
		opts = append(opts, redis.Password(config.Password))
	}
	c, err := redis.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	_redisClient.Store(name, c)
	return c, nil
}

// InitRPCClient 初始化grpc client
func InitRPCClient(ctx context.Context, name string, opts ...rpc.ClientOption) (*grpc.ClientConn, error) {
	_clientMutex.Lock()
	defer _clientMutex.Unlock()
	fullKey, config := app.Router().Config().GetClient(name, env.ProtoRPC)
	if c, has := _client.Load(fullKey); has {
		return c.(*grpc.ClientConn), nil
	}
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init rpc client, config is nil, name:%s", name))
	}
	if config.UnUseDiscovery && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	if app.Router().Discovery() == nil && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	if config.EndPoints != "" {
		eps := strings.Split(config.EndPoints, ",")
		opts = append(opts, rpc.WithEps(eps...))
	}
	if !config.UnUseDiscovery && app.Router().Discovery() != nil {
		// 开始服务发现
		watcher, err := app.Router().Discovery().Watcher(ctx, config.Namespace, config.Product, config.ServiceName, env.ProtoRPC)
		if err != nil {
			return nil, err
		}
		opts = append(opts, rpc.WithWatcher(watcher))
	}
	var timeout time.Duration
	var err error
	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}
	opts = append(opts, rpc.WithUnaryInterceptor(
		rpc.RetryClientUnaryInterceptor(config.Retry),
		rpc.TimeoutClientUnaryInterceptor(timeout),
		rpc.TracingClientUnaryInterceptor(app.Router().Ctx(), app.Router().Tracer()),
		grpc_prometheus.UnaryClientInterceptor),
	)
	c, err := rpc.DialContext(ctx, opts...)
	if err != nil {
		return nil, err
	}
	_client.Store(fullKey, c)
	return c, nil
}

func InitHttpClientUseConfig(ctx context.Context, config *config.ClientConfig) (*http.Client, error) {
	_clientMutex.Lock()
	defer _clientMutex.Unlock()

	fullName := strings.TrimLeft(
		fmt.Sprintf("%s.%s.%s", config.Namespace, config.Product, config.ServiceName),
		".",
	)
	fullKey := fmt.Sprintf("%s-%s", fullName, env.ProtoHttp)
	if c, has := _client.Load(fullKey); has {
		return c.(*http.Client), nil
	}
	if config.UnUseDiscovery && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	if app.Router().Discovery() == nil && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	var opts []http.Option
	if config.EndPoints != "" {
		eps := strings.Split(config.EndPoints, ",")
		opts = append(opts, http.WithEps(eps...))
	}
	if config.Timeout == "" {
		config.Timeout = "5s"
	}
	td, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, err
	}
	opts = append(opts, http.WithTimeout(td))
	if !config.UnUseDiscovery && app.Router().Discovery() != nil {
		// 开始服务发现
		watcher, err := app.Router().Discovery().Watcher(ctx, config.Namespace, config.Product, config.ServiceName, env.ProtoHttp)
		if err != nil {
			return nil, err
		}
		opts = append(opts, http.WithWatcher(watcher))
	}
	if config.Retry > 0 {
		opts = append(opts, http.WithRetry(config.Retry))
	}
	opts = append(opts, http.WithInterceptors(
		http.TracingInterceptor(app.Router().Ctx(), app.Router().Tracer()),
	))
	c := http.NewClient(ctx, opts...)
	_client.Store(fullKey, c)
	return c, nil
}

// InitHttpClient 初始化http client
func InitHttpClient(ctx context.Context, name string, opts ...http.Option) (*http.Client, error) {
	_clientMutex.Lock()
	defer _clientMutex.Unlock()
	fullKey, config := app.Router().Config().GetClient(name, env.ProtoHttp)
	if c, has := _client.Load(fullKey); has {
		return c.(*http.Client), nil
	}
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init http client, config is nil, name:%s", name))
	}
	if config.UnUseDiscovery && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	if app.Router().Discovery() == nil && len(strings.Split(config.EndPoints, ",")) == 0 {
		return nil, errors.New("client endpoints is nil")
	}
	if config.EndPoints != "" {
		eps := strings.Split(config.EndPoints, ",")
		opts = append(opts, http.WithEps(eps...))
	}
	if config.Timeout == "" {
		config.Timeout = "5s"
	}
	td, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, err
	}
	opts = append(opts, http.WithTimeout(td))
	if !config.UnUseDiscovery && app.Router().Discovery() != nil {
		// 开始服务发现
		watcher, err := app.Router().Discovery().Watcher(ctx, config.Namespace, config.Product, config.ServiceName, env.ProtoHttp)
		if err != nil {
			return nil, err
		}
		opts = append(opts, http.WithWatcher(watcher))
	}
	if config.Retry > 0 {
		opts = append(opts, http.WithRetry(config.Retry))
	}
	opts = append(opts, http.WithInterceptors(
		http.TracingInterceptor(app.Router().Ctx(), app.Router().Tracer()),
	))
	c := http.NewClient(ctx, opts...)
	_client.Store(fullKey, c)
	return c, nil
}

// InitRocketProducer 初始化rocket producer
func InitRocketProducer(ctx context.Context, name string, opts ...producer.Option) (*producer.Producer, error) {
	_rocketProducerMutex.Lock()
	defer _rocketProducerMutex.Unlock()
	if c, has := _rocketProducer.Load(name); has {
		return c.(*producer.Producer), nil
	}
	config := app.Router().Config().RocketProducer(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init rocket producer, config is nil, name:%s", name))
	}
	if config.Brokers == "" {
		return nil, errors.New("rocket brokers is nil")
	}
	bks := strings.Split(config.Brokers, ",")
	opts = append(opts, producer.WithNameServer(bks))
	if config.Timeout != "" {
		config.Timeout = "5s"
	}
	td, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, err
	}
	opts = append(opts, producer.WithTimeout(td))
	if config.MaxRetry == 0 {
		config.MaxRetry = 2
	}
	opts = append(opts, producer.WithRetry(config.MaxRetry))
	if config.AccessKey != "" || config.SecretKey != "" || config.SecurityToken != "" {
		opts = append(opts, producer.WithCredentials(primitive.Credentials{
			AccessKey:     config.AccessKey,
			SecretKey:     config.SecretKey,
			SecurityToken: config.SecurityToken,
		}))
	}
	opts = append(opts,
		producer.WithTracer(app.Router().Tracer()),
		producer.WithInterceptors([]primitive.Interceptor{producer.LogInterceptor(logger.GetLogger())}),
	)
	p, err := producer.NewProducer(ctx, opts...)
	if err != nil {
		return nil, err
	}
	_rocketProducer.Store(name, p)
	return p, nil
}

// InitRocketConsumer 初始化rocket consumer
func InitRocketConsumer(ctx context.Context, name string, opts ...consumer.Option) (*consumer.PushConsumer, error) {
	_rocketConsumerMutex.Lock()
	defer _rocketConsumerMutex.Unlock()
	if c, has := _rocketConsumer.Load(name); has {
		return c.(*consumer.PushConsumer), nil
	}
	config := app.Router().Config().RocketConsumer(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init rocket consumer, config is nil, name:%s", name))
	}
	if config.Brokers == "" {
		return nil, errors.New("rocket brokers is nil")
	}
	bks := strings.Split(config.Brokers, ",")
	opts = append(opts, consumer.WithNameServer(bks))
	if config.ConsumeTimeout == "" {
		config.ConsumeTimeout = "30m"
	}
	td, err := time.ParseDuration(config.ConsumeTimeout)
	if err != nil {
		return nil, err
	}
	opts = append(opts, consumer.WithConsumeTimeout(td))
	if config.MaxRetry == 0 {
		config.MaxRetry = 2
	}
	opts = append(opts, consumer.WithRetry(config.MaxRetry))
	if config.AccessKey != "" || config.SecretKey != "" || config.SecurityToken != "" {
		opts = append(opts, consumer.WithCredentials(primitive.Credentials{
			AccessKey:     config.AccessKey,
			SecretKey:     config.SecretKey,
			SecurityToken: config.SecurityToken,
		}))
	}
	if config.MaxReconsumeTimes > 0 {
		opts = append(opts, consumer.WithMaxReconsumeTimes(config.MaxReconsumeTimes))
	}
	if config.Expression == "" {
		config.Expression = "*"
	}
	opts = append(opts,
		consumer.WithTracer(app.Router().Tracer()),
		consumer.WithInterceptors([]primitive.Interceptor{consumer.LogInterceptor(logger.GetLogger())}),
		consumer.WithFrom(config.From),
		consumer.WithGoroutineNums(runtime.NumCPU()*5),
		consumer.WithGroupName(config.GroupName),
		consumer.WithModel(config.Mode),
		consumer.WithTagExpression(config.Expression),
	)
	con, err := consumer.NewPushConsumer(ctx, opts...)
	if err != nil {
		return nil, err
	}
	_rocketConsumer.Store(name, con)
	return con, nil
}

// InitKafkaProducer 初始化kafka生产者
func InitKafkaProducer(name string, opts ...kafka.ProducerOption) (*kafka.Producer, error) {
	_kafkaProducerMutex.Lock()
	defer _kafkaProducerMutex.Unlock()
	if c, has := _kafkaProducer.Load(name); has {
		return c.(*kafka.Producer), nil
	}
	config := app.Router().Config().KafkaProducer(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init kafka producer, config is nil, name:%s", name))
	}
	if config.Brokers == "" {
		return nil, errors.New(fmt.Sprintf("app init kafka producer, brokers is nil, name:%s", name))
	}
	t := app.Router().Tracer()
	if t == nil {
		return nil, errors.New("app init kafka producer, tracer is nil")
	}
	opts = append(opts, kafka.ProducerMessageBuilder(kfkCore.NewMessageBuilder(t)))
	opts = append(opts, kafka.ProducerNotifyDisable(config.DisableNotify))
	brokers := strings.Split(config.Brokers, ",")
	if config.Timeout != "" {
		td, err := time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("app init kafka producer, config timeout, value:%s", config.Timeout))
		}
		opts = append(opts, kafka.ProducerTimeout(td))
	}
	if config.MaxMessageBytes > 0 {
		opts = append(opts, kafka.ProducerMaxMessageBytes(config.MaxMessageBytes))
	}
	if config.MaxRetry > 0 {
		opts = append(opts, kafka.ProducerRetry(config.MaxRetry))
	}
	if config.Mode != "" {
		opts = append(opts, kafka.ProducerModel(config.Mode))
	}
	p, err := kafka.NewProducer(app.Router().Ctx(), brokers, opts...)
	if err != nil {
		return nil, err
	}
	_kafkaProducer.Store(name, p)
	return p, nil
}

// InitKafkaConsumer 初始化kafka消费者
func InitKafkaConsumer(name string, opts ...kafka.ConsumerOption) (*kafka.Consumer, error) {
	_kafkaConsumerMutex.Lock()
	defer _kafkaConsumerMutex.Unlock()
	//TODO 从_kafkaProducer修改为_kafkaConsumer
	if c, has := _kafkaConsumer.Load(name); has {
		return c.(*kafka.Consumer), nil
	}
	config := app.Router().Config().KafkaConsumer(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init kafka consumer, config is nil, name:%s", name))
	}
	if config.Brokers == "" {
		return nil, errors.New(fmt.Sprintf("app init kafka consumer, brokers is nil, name:%s", name))
	}
	if config.Topics == "" {
		return nil, errors.New(fmt.Sprintf("app init kafka consumer, topics is nil, name:%s", name))
	}
	if config.Group == "" {
		return nil, errors.New(fmt.Sprintf("app init kafka consumer, group is nil, name:%s", name))
	}
	t := app.Router().Tracer()
	if t == nil {
		return nil, errors.New("app init kafka consumer, tracer is nil")
	}
	opts = append(opts, kafka.ConsumerMessageBuilder(kfkCore.NewMessageBuilder(t)))
	brokers := strings.Split(config.Brokers, ",")
	if config.MaxWaitTime != "" {
		td, err := time.ParseDuration(config.MaxWaitTime)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("app init kafka consumer, config maxwaittime, value:%s", config.MaxWaitTime))
		}
		opts = append(opts, kafka.ConsumerMaxWaitTime(td))
	}
	if config.CommitRetry > 0 {
		opts = append(opts, kafka.ConsumerCommitRetry(config.CommitRetry))
	}
	if config.OffsetInitial != 0 {
		opts = append(opts, kafka.ConsumerOffsetInitial(config.OffsetInitial))
	}
	if config.ReBalance != "" {
		rbs := strings.Split(config.ReBalance, ",")
		if len(rbs) > 0 && rbs[0] != "" {
			opts = append(opts, kafka.ConsumerReBalance(rbs))
		}
	}
	if config.TopicCreateEnable {
		opts = append(opts, kafka.ConsumerTopicCreateEnable(config.TopicCreateEnable))
	}
	if !config.AutoCommit {
		opts = append(opts, kafka.ConsumerAutoCommit(config.AutoCommit))
	}
	c, err := kafka.NewConsumer(app.Router().Ctx(), config.Group, brokers, config.Topics, opts...)
	if err != nil {
		return nil, err
	}
	_kafkaConsumer.Store(name, c)
	return c, nil
}

// InitSqlClient 初始化openGuess客户端
func InitPostgresql(name string) (*mysql.Client, error) {
	_sqlMutex.Lock()
	defer _sqlMutex.Unlock()
	if c, has := _sqlClient.Load(name); has {
		return c.(*mysql.Client), nil
	}
	config := app.Router().Config().GetPostgresql(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("app init sql client, config is nil, name:%s", name))
	}

	//初始化Postgresql连接
	host := config.Host
	user := config.User
	password := config.Password
	dbname := config.Dbname
	port := config.Port
	sslmode := config.Sslmode
	//dsn := "host=localhost user=postgres password=YOUR_PASSWORD dbname=YOUR_DBNAME port=5432 sslmode=disable"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	c, err := mysql.NewPostgresqlClient(dsn)
	if err != nil {
		return nil, err
	}
	_sqlClient.Store(name, c)
	return c, nil
}
