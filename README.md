# 基础框架golang版本

## 环境变量说明
NACOS
> NACOS_SERVER_PATH - 配置中心server地址(必要)  
> NACOS_ACCESS - 配置中心鉴权accessKey  
> NACOS_SECRET - 配置中心鉴权secretKey  
> NACOS_USERNAME - 配置中心Username  
> NACOS_PASSWORD - 配置中心password  
> NACOS_CONFIG_FORMAT - 配置文件格式(json/xml/yaml)

ETCD(服务发现)
> ETCD_ENDPOINTS - etcd地址

Consul(服务发现)
> CONSUL_HTTP_ADDR - consul地址

Sentry
> SENTRY_DNS - sentry错误报警服务地址  

Jaeger
> JAEGER_ADDR - jaeger链路追踪信息收集服务地址  

其他配置
> METRIC_DISABLE - 是否关闭本地监控(runtime/pprof) true:关闭(不推荐)  
> SERVICE_ENV - online:线上环境 testing:测试环境 (默认值为testing,该项必须保证准确性)  
> LOG_PATH - 日志保存路径

## 集成中间件
Mysql : gorm  
Redis : go-redis & go-redsync/redsync  
服务发现: etcd,consul  
配置中心: nacos  
消息队列: rocketMQ  
链路追踪: jaeger  
监控: pprof+runtime  
错误监控: sentry  

## 支持服务类型
http & https & grpc & websocket & socket.io  

## 配置说明

### 日志配置
rotation - 日志分割方式(day/hour) 默认day  
saveDays - 日志保存天数 默认3  
level - 日志级别(debug/info/warn/error) 默认debug  
format - 日志格式(console/json) 默认console  
```json
{
  "rotation": "hour",
  "saveDays": 7,
  "level": "debug",
  "format": "json"
}
```

### 启动服务配置
proto - 协议类型(http/rpc/websocket)  
port - 端口号
```json
[
  {
    "proto": "http",
    "port": 11001
  },
  {
    "proto": "websocket",
    "port": 11002
  }
]
```

### 服务发现配置
used - 服务发现方式(etcd/consul) 需配合对应的环境变量配置  
```json
{
  "used": "etcd"
}
```

### 数据库配置
name - 名称 配置检索使用  
master - 主库dns 主从配置将进行读写分离  
slaves - 从库dns  
maxOpen - 链接池最大连接数 配置前请通知DBA  
maxIdle - 链接池最大空闲连接数 配置前请通知DBA  
maxLifeTime - 链接可复用最大时间  
maxIdleTime - 链接池链接最大空闲时长  
```json
[
  {
    "name":"racing", 
    "master":"app:6oXXByzlNoexPye7@tcp(192.168.0.19:3306)/racing?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s&readTimeout=5s", 
    "slaves":["app:6oXXByzlNoexPye7@tcp(192.168.0.19:3306)/racing?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s&readTimeout=5s"], 
    "maxOpen": 100, 
    "maxIdle": 50, 
    "maxLifeTime": "15m", 
    "maxIdleTime": "1h"
  }
]
```

### redis配置
name - 名称 配置检索使用  
model - redis模式(singleton/sentinel/cluster) 默认singleton  
addr - 地址 多地址则用','分割  
database - db库  
maxRetry - 最大重试次数  
idleTimeout - 客户端关闭空闲连接的时间 应该小于服务器的超时时间 默认为300秒 -1禁用空闲超时检查  
readTimeout - 命令读超时 默认3(秒单位)  
writeTimeout - 命令写超时 默认等于读超时(秒单位)
poolSize - 连接池大小 默认100 配置前请通知DBA  
minIdleConn - 最小连接数量 默认35 配置前请通知DBA  
password - 密码  
```json
[
  {
    "name": "racing",
    "model": "singleton",
    "addr": "192.168.0.19:6379",
    "database": 0,
    "maxRetry": 3,
    "idleTimeout": 600,
    "readTimeout": 5,
    "writeTimeout": 3,
    "poolSize": 150,
    "minIdleConn": 50,
    "password": "BGGjeI"
  }
]
```

### 下游服务配置
namespace - 下游服务namespace  
product - 下游服务产品  
serviceName - 服务名称 namespace.namespace.serviceName为服务发现名  
proto - 下游服务协议(rpc/http)  
endpoints - endpoints 多个','分割 如果使用服务发现则该配置为兜底 如果未使用服务发现则该配置生效  
unUseDiscovery - 是否禁止服务发现 默认false 如果服务本身未配置服务发现则调用下游服务无法使用服务发现功能  
retry - 重试次数  
timeout - 超时时间 注意这里的超时时间为单次请求超时 所以接口的最坏超时需要乘以retry  
```json
[
  {
    "serviceName": "pt",
    "proto":"http",
    "endpoints":"https://api.testing.dream22.xyz",
    "unUseDiscovery":true,
    "retry":2,
    "timeout":"3s"
  },
  {
    "namespace": "aries",
    "product": "common",
    "serviceName": "push.link",
    "proto": "http",
    "retry": 2,
    "timeout":"3s"
  }
]
```

### rocketmq producer配置
name - 名称 配置检索使用  
brokers - brokers 使用','分割  
timeout - 发送超时时间  
accessKey - 鉴权用accessKey  
secretKey - 鉴权用secretKey  
securityToken - 鉴权用securityToken  
maxRetry - 写入最大重试次数  
```json
[
  {
    "name": "racing_producer",
    "brokers": "192.168.0.19:9876",
    "timeout": "5s",
    "maxRetry": 1
  }
]
```

### rocketmq consumer配置
name - 名称 配置检索使用  
brokers - brokers 使用','分割  
consumeTimeout - 队列消息超时时间  
accessKey - 鉴权用accessKey  
secretKey - 鉴权用secretKey  
securityToken - 鉴权用securityToken  
maxRetry - 读取最大重试次数  
mode - 消费模式 0:broadcasting 1:clustering  
from - 消费起点 0:最近一次提交 1:记录的最早一次提交 2:当前时间  
groupName - 消费组名称  
maxReconsumeTimes - 消费重试次数  
expression - Tag标签过滤器  
```json
[
  {
    "name": "racing_consumer",
    "brokers": "192.168.0.19:9876",
    "consumeTimeout": "60m",
    "maxRetry": 5,
    "mode": 1,
    "from": 0,
    "groupName": "racing_busi_settle-topic",
    "maxReconsumeTimes": 1,
    "expression": "*"
  }
]
```

### 其他
accessRequestDisable - access日志禁止输出request信息 默认false  

### 自定义配置
配置文件中除标准配置外可增加自定义配置
```json
{
  "groups":[
    {
      "groupID":1,
      "groupName":"racing_group_90s",
      "rtp":90,
      "bettingSec":55,
      "racingSec1":25,
      "racingSec2":5,
      "lotterySec":5,
      "text": "1.5min"
    },
    {
      "groupID":2,
      "groupName":"racing_group_180s",
      "rtp":90,
      "bettingSec":145,
      "racingSec1":25,
      "racingSec2":5,
      "lotterySec":5,
      "text": "3min"
    }
  ]
}
```
标准配置由框架读取，自定义配置需要自行解析
```go
package conf

import "github.com/curry-mz/sagittarius-golang/app"

type Config struct {
	Groups []struct {
		GroupID    int    `json:"groupID"`
		GroupName  string `json:"groupName"`
		RTP        int    `json:"rtp"`
		BettingSec int64  `json:"bettingSec"`
		RacingSec1 int64  `json:"racingSec1"`
		RacingSec2 int64  `json:"racingSec2"`
		LotterySec int64  `json:"lotterySec"`
		Text       string `json:"text"`
		Image      string `json:"image"`
	} `json:"groups"`
}

var conf *Config

func InitAppConfig() {
	var cfg Config
	// 解析自定义配置
	// err := app.Router().ExtraXmlConfig
	// err := app.Router().ExtraYamlConfig
	err := app.Router().ExtraJsonConfig(&cfg)
	if err != nil {
		panic(err)
	}
	conf = &cfg
}
```

### 整体例子
```json
{
  "accessRequestDisable": true,
  "log": {
    "rotation": "day",
    "saveDays": 3,
    "level": "debug"
  },
  "servers":[{"proto":"http","port":11001}],
  "discovery":{
    "used": "consul"
  },
  "databases":[
    {
      "name":"racing",
      "master":"app:6oXXByzlNoexPye7@tcp(192.168.0.19:3306)/racing?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s&readTimeout=5s",
      "slaves":["app:6oXXByzlNoexPye7@tcp(192.168.0.19:3306)/racing?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s&readTimeout=5s"],
      "maxOpen": 100,
      "maxIdle": 50,
      "maxLifeTime": "15m",
      "maxIdleTime": "1h"
    }
  ],
  "redis":[
    {
      "name": "racing",
      "model": "singleton",
      "addr": "192.168.0.19:6379",
      "database": 0,
      "maxRetry": 3,
      "password": "BGGjeI"
    }
  ],
  "clients":[
    {
      "serviceName": "pt",
      "proto":"http",
      "endpoints":"https://api.testing.dream22.xyz",
      "unUseDiscovery":true,
      "retry":2,
      "timeout":"3s"
    },
    {
      "namespace": "aries",
      "product": "common",
      "serviceName": "push.link",
      "proto": "http",
      "retry": 2,
      "timeout":"3s"
    }
  ],
  "producers": [
    {
      "name": "racing_producer",
      "brokers": "192.168.0.19:9876",
      "timeout": "5s",
      "maxRetry": 1
    }
  ],
  "consumers": [
    {
      "name": "racing_consumer",
      "brokers": "192.168.0.19:9876",
      "consumeTimeout": "60m",
      "maxRetry": 5,
      "mode": 1,
      "from": 0,
      "groupName": "racing_busi_settle-topic",
      "maxReconsumeTimes": 1,
      "expression": "*"
    }
  ],
  "groups":[
    {
      "groupID":1,
      "groupName":"racing_group_90s",
      "rtp":90,
      "bettingSec":55,
      "racingSec1":25,
      "racingSec2":5,
      "lotterySec":5,
      "text": "1.5min"
    },
    {
      "groupID":2,
      "groupName":"racing_group_180s",
      "rtp":90,
      "bettingSec":145,
      "racingSec1":25,
      "racingSec2":5,
      "lotterySec":5,
      "text": "3min"
    }
  ]
}
```

## 代码示例

### 框架启动
```go
package main

import (
	"flag"

	"code.cd.local/games-go/racing/racing.busi/conf"
	"code.cd.local/games-go/racing/racing.busi/server"

	"github.com/curry-mz/sagittarius-golang/app"
	"github.com/curry-mz/sagittarius-golang/app/config"
	"github.com/curry-mz/sagittarius-golang/env"
)

var (
	confPath = flag.String("confPath", "", "confPath is none")
)

func init() {
	flag.Parse()

	if *confPath != "" && env.IsTesting() {
		// 仅测试环境支持通过启动项传入配置文件地址
		config.WithConfigPath(*confPath)
	}
	// 初始化框架&服务定义
	app.InitRouter(&config.ServiceDefine{
		Namespace:   Namespace,
		Product:     Product,
		ServiceName: ServiceName,
	})
}

func main() {
	// 自定义配置解析
	conf.InitAppConfig()
	// 初始化启动服务
	srv := server.NewHttp()
	// 将server绑定到框架
	app.Router().BindServer(srv)
	// 框架启动
	app.Run()
}
```

### 下游服务调用
```go
package push

import (
	"context"
	"fmt"
	netHttp "net/http"

	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/cores/client/http"
	gErrors "github.com/curry-mz/sagittarius-golang/cores/errors"
	"github.com/curry-mz/sagittarius-golang/logger"

	"github.com/pkg/errors"
)

const (
	Name = "aries.common.push.link"
)

const (
	MessagePushUrl = "/server_api/v1/link/push/message"
)

type RequestPushMessage struct {
	App     string      `json:"app"`  // required
	Mode    int         `json:"mode"` // 1:app内广播 2:组内广播 3:指定id推送
	Group   string      `json:"group,omitempty"`
	To      []string    `json:"to,omitempty"`
	Version string      `json:"version,omitempty"`
	Tag     string      `json:"tag,omitempty"`
	SubType int32       `json:"subType"`
	Payload interface{} `json:"payload,omitempty"`
}

type ResponsePushMessage struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func MessagePush(ctx context.Context, req *RequestPushMessage) error {
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	var rsp ResponsePushMessage
	httpRsp, err := c.JsonPost(http.Request(ctx, MessagePushUrl), req, &rsp)
	if err != nil {
		return errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return errors.New(fmt.Sprintf("post %s, httpCode:%d", MessagePushUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return gErrors.New(rsp.Status, rsp.Message)
	}
	return nil
}
```

### 数据库使用
```go
package db

import (
	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/mysql"
)

const dbName = "racing"

type RacingDB struct {
	sql *mysql.Client
}

func New() (*RacingDB, error) {
	cli, err := proxy.InitSqlClient(dbName)
	if err != nil {
		return nil, err
	}
	return &RacingDB{sql: cli}, nil
}
```

### redis使用
```go
package redis

import (
	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/redis"
)

const redisName = "racing"

type RacingRedis struct {
	rds *redis.Client
}

func New() (*RacingRedis, error) {
	cli, err := proxy.InitRedisClient(redisName)
	if err != nil {
		return nil, err
	}
	return &RacingRedis{rds: cli}, nil
}
```
