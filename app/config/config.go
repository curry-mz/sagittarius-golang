package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/curry-mz/sagittarius-golang/env"
	"github.com/curry-mz/sagittarius-golang/logger"
	"github.com/curry-mz/sagittarius-golang/nacos"

	"github.com/pkg/errors"
)

/////////////////////////////////////////////////
// 服务强制定义项
/////////////////////////////////////////////////

// 服务发现明完整为namespace.product.name

// ServiceDefine 服务定义信息
type ServiceDefine struct {
	Namespace   string // 所属命名空间
	Product     string // 产品 product
	ServiceName string // 服务名(两段式推荐)
}

/////////////////////////////////////////////////

// LogConfig 服务日志配置
type LogConfig struct {
	// 日志分割方式
	Rotation string `yaml:"rotation" json:"rotation" xml:"rotation"`
	// 日志保存天数
	SaveDays int `yaml:"saveDays" json:"saveDays" xml:"saveDays"`
	// 日志级别
	Level string `yaml:"level" json:"level" xml:"level"`
	// 日志格式
	Format string `yaml:"format" json:"format" xml:"format"`
}
type KkplusConfig struct {
	// 服务发现方式 目前只有etcd
	Host      string `yaml:"host" json:"host" xml:"host"`
	MerNo     string `yaml:"mer_no" json:"mer_no" xml:"mer_no"`
	NotifyUrl string `yaml:"notify_url" json:"notify_url" xml:"notify_url"`
}

// ServerConfig 启动服务配置
type ServerConfig struct {
	// 协议类型 http/rpc/websocket
	Proto string `yaml:"proto" json:"proto" xml:"proto"`
	// 启动端口
	Port int `yaml:"port" json:"port" xml:"port"`
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	// 服务发现方式 目前只有etcd
	Used string `yaml:"used" json:"used" xml:"used"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" xml:"name"`
	// 主库dns
	Master string `yaml:"master" json:"master" xml:"master"`
	// 从库dns
	Slaves []string `yaml:"slaves" json:"slaves" xml:"slaves"`
	// 最大连接数
	MaxOpen int `yaml:"maxOpen" json:"maxOpen" xml:"maxOpen"`
	// 最大空闲连接数
	MaxIdle int `yaml:"maxIdle" json:"maxIdle" xml:"maxIdle"`
	// 链接可复用最大时间
	MaxLifeTime string `yaml:"maxLifeTime" json:"maxLifeTime" xml:"maxLifeTime"`
	// 链接池链接最大空闲时长
	MaxIdleTime string `yaml:"maxIdleTime" json:"maxIdleTime" xml:"maxIdleTime"`
}

// PostgresqlConfig 数据库配置
type PostgresqlConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" xml:"name"`
	// 地址
	Host string `yaml:"host" json:"host" xml:"host"`
	// 端口号
	Port string `yaml:"port" json:"port" xml:"port"`
	// 用户名
	User string `yaml:"user" json:"user" xml:"user"`
	// 密码
	Password string `yaml:"password" json:"password" xml:"password"`
	// 连接数据库名
	Dbname string `yaml:"dbname" json:"dbname" xml:"dbname"`
	// 禁用模式
	Sslmode string `yaml:"sslmode" json:"sslmode" xml:"sslmode"`
}

// RedisConfig redis缓存配置
type RedisConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" xml:"name"`
	// 模式 singleton/sentinel/cluster 默认singleton
	Model string `yaml:"model" json:"model" xml:"model"`
	// 地址 多地址则用','分割
	Addr string `yaml:"addr" json:"addr" xml:"addr"`
	// redis db
	DataBase int `yaml:"database" json:"database" xml:"database"`
	// 最大重试次数
	MaxRetry int `yaml:"maxRetry" json:"maxRetry" xml:"maxRetry"`
	// 客户端关闭空闲连接的时间。应该小于服务器的超时时间。默认为5分钟。-1禁用空闲超时检查。
	IdleTimeout int `yaml:"idleTimeout" json:"idleTimeout" xml:"idleTimeout"`
	// 命令读超时，默认3(秒单位)
	ReadTimeout int `yaml:"readTimeout" json:"readTimeout" xml:"readTimeout"`
	// 命令写超时，默认等于读超时(秒单位)
	WriteTimeout int `yaml:"writeTimeout" json:"writeTimeout" xml:"writeTimeout"`
	// 连接池大小 默认100
	PoolSize int `yaml:"poolSize" json:"poolSize" xml:"poolSize"`
	// 最小连接数量 默认35
	MinIdleConn int `yaml:"minIdleConn" json:"minIdleConn" xml:"minIdleConn"`
	// 密码
	Password string `yaml:"password" json:"password" xml:"password"`
}

// ClientConfig 下游客户端配置
type ClientConfig struct {
	// 下游服务namespace
	Namespace string `yaml:"namespace" json:"namespace" xml:"namespace"`
	// 下游服务产品
	Product string `yaml:"product" json:"product" xml:"product"`
	// 服务名称
	ServiceName string `yaml:"serviceName" json:"serviceName" xml:"serviceName"`
	// 下游服务协议 rpc/http
	Proto string `yaml:"proto" json:"proto" xml:"proto"`
	// endpoints 多个','分割
	EndPoints string `yaml:"endpoints" json:"endpoints" xml:"endpoints"`
	// 是否禁止服务发现
	UnUseDiscovery bool `yaml:"unUseDiscovery" json:"unUseDiscovery" xml:"unUseDiscovery"`
	// 重试次数
	Retry int `yaml:"retry" json:"retry" xml:"retry"`
	// 超时时间
	Timeout string `yaml:"timeout" json:"timeout" xml:"timeout"`
}

// RocketProducerConfig rocket producer配置
type RocketProducerConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" xml:"name"`
	// brokers 使用','分割
	Brokers string `yaml:"brokers" json:"brokers" xml:"brokers"`
	// 发送超时时间
	Timeout string `yaml:"timeout" json:"timeout" xml:"timeout"`
	// 鉴权用accessKey
	AccessKey string `yaml:"accessKey" json:"accessKey" xml:"accessKey"`
	// 鉴权用secretKey
	SecretKey string `yaml:"secretKey" json:"secretKey" xml:"secretKey"`
	// 鉴权用securityToken
	SecurityToken string `yaml:"securityToken" json:"securityToken" xml:"securityToken"`
	// 写入最大重试次数
	MaxRetry int `yaml:"maxRetry" json:"maxRetry" xml:"maxRetry"`
}

// RocketConsumerConfig rocket consumer配置
type RocketConsumerConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" xml:"name"`
	// brokers 使用','分割
	Brokers string `yaml:"brokers" json:"brokers" xml:"brokers"`
	// 队列消息超时时间
	ConsumeTimeout string `yaml:"consumeTimeout" json:"consumeTimeout" xml:"consumeTimeout"`
	// 鉴权用accessKey
	AccessKey string `yaml:"accessKey" json:"accessKey" xml:"accessKey"`
	// 鉴权用secretKey
	SecretKey string `yaml:"secretKey" json:"secretKey" xml:"secretKey"`
	// 鉴权用securityToken
	SecurityToken string `yaml:"securityToken" json:"securityToken" xml:"securityToken"`
	// 读取最大重试次数
	MaxRetry int `yaml:"maxRetry" json:"maxRetry" xml:"maxRetry"`
	// 消费模式 0:broadcasting 1:clustering
	Mode int `yaml:"mode" json:"mode" xml:"mode"`
	// 消费起点 0:最近一次提交 1:记录的最早一次提交 2:当前时间
	From int `yaml:"from" json:"from" xml:"from"`
	// 消费组名称
	GroupName string `yaml:"groupName" json:"groupName" xml:"groupName"`
	// 消费重试次数
	MaxReconsumeTimes int32 `yaml:"maxReconsumeTimes" json:"maxReconsumeTimes" xml:"maxReconsumeTimes"`
	// Tag标签过滤器
	Expression string `yaml:"expression" json:"expression" xml:"expression"`
}

// KafkaProducerConfig kafka producer配置
type KafkaProducerConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" toml:"name"`
	// brokers 使用','分割
	Brokers string `yaml:"brokers" json:"brokers" xml:"brokers"`
	// 禁止发送结果通知 默认false false情况下必须监听success和error
	DisableNotify bool `yaml:"disableNotify" json:"disableNotify" toml:"disableNotify"`
	// 超时时间
	Timeout string `yaml:"timeout" json:"timeout" toml:"timeout"`
	// 最大消息字节
	MaxMessageBytes int `yaml:"maxMessageBytes" json:"maxMessageBytes" toml:"maxMessageBytes"`
	// 写入最大重试次数
	MaxRetry int `yaml:"maxRetry" json:"maxRetry" toml:"maxRetry"`
	// 模式 sync/async 默认async
	Mode string `yaml:"mode" json:"mode" toml:"mode"`
}

// KafkaConsumerConfig kafka consumer配置
type KafkaConsumerConfig struct {
	// 名称
	Name string `yaml:"name" json:"name" toml:"name"`
	// brokers 使用','分割
	Brokers string `yaml:"brokers" json:"brokers" toml:"brokers"`
	// 组名
	Group string `yaml:"group" json:"group" toml:"group"`
	// topic 使用','分割
	Topics string `yaml:"topics" json:"topics" toml:"topics"`
	// 分区策略 使用','分割 sticky/roundrobin/range 默认sticky
	ReBalance string `yaml:"reBalance" json:"reBalance" toml:"reBalance"`
	// 消费offset最新或最早 -1/-2 默认-1最新
	OffsetInitial int64 `yaml:"offsetInitial" json:"offsetInitial" toml:"offsetInitial"`
	// 提交最大重试次数
	CommitRetry int `yaml:"commitRetry" json:"commitRetry" toml:"commitRetry"`
	// 返回消息最大等待时间
	MaxWaitTime string `yaml:"maxWaitTime" json:"maxWaitTime" toml:"maxWaitTime"`
	// 是否允许自动创建不存在的topic 默认false
	TopicCreateEnable bool `yaml:"topicCreateEnable" json:"topicCreateEnable" toml:"topicCreateEnable"`
	// 是否允许自动创提交offset
	AutoCommit bool `yaml:"autoCommit" json:"autoCommit" toml:"autoCommit"`
}

type ServiceConfig struct {
	// access日志禁止输出request信息 默认false
	AccessRequestDisable bool `yaml:"accessRequestDisable" json:"accessRequestDisable" xml:"accessRequestDisable"`
	// 日志配置
	Log *LogConfig `yaml:"log" json:"log" xml:"log"`
	// 启动服务配置
	Svrs []*ServerConfig `yaml:"servers" json:"servers" xml:"servers"`
	// 服务发现配置
	Discovery *DiscoveryConfig `yaml:"discovery" json:"discovery" xml:"discovery"`
	// 数据库配置
	Databases []*DatabaseConfig `yaml:"databases" json:"databases" xml:"databases"`
	//postgresql数据库配置
	Postgresql []*PostgresqlConfig `yaml:"postgresql" json:"postgresql" xml:"postgresql"`
	// redis配置
	Rds []*RedisConfig `yaml:"redis" json:"redis" xml:"redis"`
	// 下游服务配置
	Clients []*ClientConfig `yaml:"clients" json:"clients" xml:"clients"`
	// rocket producer配置
	RocketProducers []*RocketProducerConfig `yaml:"rocketProducers" json:"rocketProducers" xml:"rocketProducers"`
	// rocket consumer配置
	RocketConsumers []*RocketConsumerConfig `yaml:"rocketConsumers" json:"rocketConsumers" xml:"rocketConsumers"`
	// kafka producer配置
	KafkaProducers []*KafkaProducerConfig `yaml:"kafkaProducers" json:"kafkaProducers" xml:"kafkaProducers"`
	// kafka consumer配置
	KafkaConsumers []*KafkaConsumerConfig `yaml:"kafkaConsumers" json:"kafkaConsumers" xml:"kafkaConsumers"`
	//kkplus
	Kkplus *KkplusConfig `yaml:"kkplus" json:"kkplus" xml:"kkplus"`
}

func (c *ServiceConfig) GetDatabase(name string) *DatabaseConfig {
	for _, db := range c.Databases {
		if db.Name == name {
			return db
		}
	}
	return nil
}

func (c *ServiceConfig) GetPostgresql(name string) *PostgresqlConfig {
	for _, db := range c.Postgresql {
		if db.Name == name {
			return db
		}
	}
	return nil
}

func (c *ServiceConfig) GetRedis(name string) *RedisConfig {
	for _, rds := range c.Rds {
		if rds.Name == name {
			return rds
		}
	}
	return nil
}

func (c *ServiceConfig) GetClient(name string, proto string) (string, *ClientConfig) {
	for _, cli := range c.Clients {
		fullName := strings.TrimLeft(
			fmt.Sprintf("%s.%s.%s", cli.Namespace, cli.Product, cli.ServiceName),
			".",
		)
		if name == fullName && proto == strings.ToLower(cli.Proto) {
			return fmt.Sprintf("%s-%s", fullName, strings.ToLower(cli.Proto)), cli
		}
	}
	return "", nil
}

func (c *ServiceConfig) RocketProducer(name string) *RocketProducerConfig {
	for _, p := range c.RocketProducers {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (c *ServiceConfig) RocketConsumer(name string) *RocketConsumerConfig {
	for _, con := range c.RocketConsumers {
		if con.Name == name {
			return con
		}
	}
	return nil
}

func (c *ServiceConfig) KafkaProducer(name string) *KafkaProducerConfig {
	for _, p := range c.KafkaProducers {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (c *ServiceConfig) KafkaConsumer(name string) *KafkaConsumerConfig {
	for _, con := range c.KafkaConsumers {
		if con.Name == name {
			return con
		}
	}
	return nil
}

/////////////////////////////////////////////////

type Option func(*option)

type option struct {
	path   string
	prefix string
}

func WithPath(path string) Option {
	return func(o *option) {
		o.path = path
	}
}

func WithPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

func Initialize(sd *ServiceDefine, v interface{}, opts ...Option) (*nacos.Client, string, error) {
	var err error
	var cfgStr string
	var cli *nacos.Client
	o := option{}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.path != "" && env.IsTesting() {
		var bs []byte
		bs, err = ioutil.ReadFile(o.path)
		if err != nil {
			return nil, "", err
		}
		if len(bs) == 0 {
			return nil, "", errors.New("config file does not exist")
		}
		err = json.Unmarshal(bs, v)
		if err != nil {
			return nil, "", err
		}
		cfgStr = string(bs)
	} else {
		path, accessKey, secretKey, cfgFormat, userName, password := env.GetNacos()
		if path == "" {
			return nil, "", errors.New("nacos-server config center path undefined")
		}
		// 创建nacos客户端
		ncopts := []nacos.Option{
			nacos.WithNamespace(sd.Namespace),
			nacos.WithProduct(sd.Product),
			nacos.WithName(sd.ServiceName),
			nacos.WithRunEnv(env.GetRunEnv()),
			nacos.WithLogger(logger.GetGen()),
			nacos.WithServerPath(path),
		}
		if accessKey != "" {
			ncopts = append(ncopts, nacos.WithAccessKey(accessKey))
		}
		if secretKey != "" {
			ncopts = append(ncopts, nacos.WithSecretKey(secretKey))
		}
		if userName != "" {
			ncopts = append(ncopts, nacos.WithUserName(userName))
		}
		if password != "" {
			ncopts = append(ncopts, nacos.WithPassword(password))
		}
		cli = nacos.NewClient(ncopts...)
		name := fmt.Sprintf("%s.%s.config", sd.Product, sd.ServiceName)
		if o.prefix != "" {
			name = fmt.Sprintf("%s-%s", o.prefix, name)
		}
		// 读取配置格式
		switch strings.ToLower(cfgFormat) {
		case "yaml":
			cfgStr, err = cli.GetYamlConfig(name, v)
			if err != nil {
				return nil, "", err
			}
		case "xml":
			cfgStr, err = cli.GetXmlConfig(name, v)
			if err != nil {
				return nil, "", err
			}
		default:
			// 默认json
			cfgStr, err = cli.GetJsonConfig(name, v)
			if err != nil {
				return nil, "", err
			}
		}
	}
	return cli, cfgStr, nil
}
