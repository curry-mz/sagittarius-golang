package env

const (
	// NacosServerPath 配置中心nacos地址 必要
	NacosServerPath = "NACOS_SERVER_PATH"
	// NacosAccess accessKey
	NacosAccess = "NACOS_ACCESS"
	// NacosSecret secretKey
	NacosSecret = "NACOS_SECRET"
	// NacosUsername userName
	NacosUsername = "NACOS_USERNAME"
	// NacosPassword password
	NacosPassword = "NACOS_PASSWORD"
	// NacosConfigFormat 配置文件格式 json/xml/yaml
	NacosConfigFormat = "NACOS_CONFIG_FORMAT"
	// EtcdEndPoints Etcd endpoints
	EtcdEndPoints = "ETCD_ENDPOINTS"
	// SentryDNS sentry地址(错误报警) 可选
	SentryDNS = "SENTRY_DNS"
	// JaegerAddr jaeger地址(链路追踪收集) 可选
	JaegerAddr = "JAEGER_ADDR"
	// MetricDisable 监控开关 true为关闭 默认false
	MetricDisable = "METRIC_DISABLE"
	// ServiceEnv 环境 testing:测试环境 默认testing
	ServiceEnv = "SERVICE_ENV"
	// ConsulAddr Consul consul http地址
	ConsulAddr = "CONSUL_HTTP_ADDR"
	// LogPath 日志路径
	LogPath = "LOG_PATH"
)

const (
	TestingEnv = "testing"
)

const (
	TRUE  = "true"
	FALSE = "false"
)

const (
	ProtoRPC       = "rpc"
	ProtoHttp      = "http"
	ProtoWebsocket = "websocket"
	ProtoSocketIO  = "io"
)
