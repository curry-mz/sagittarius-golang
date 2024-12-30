package nacos

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"code.cd.local/sagittarius/sagittarius-golang/cores/logger"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Option func(*options)

type options struct {
	namespace  string
	product    string
	name       string
	timeOut    uint64
	logger     *logger.Logger
	accessKey  string
	secretKey  string
	serverPath string
	runEnv     string
	userName   string
	password   string
}

// WithNamespace namespace
func WithNamespace(namespace string) Option {
	return func(o *options) {
		o.namespace = namespace
	}
}

// WithProduct product
func WithProduct(product string) Option {
	return func(o *options) {
		o.product = product
	}
}

// WithName service name
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithTimeOut client time out
func WithTimeOut(timeOut uint64) Option {
	return func(o *options) {
		o.timeOut = timeOut
	}
}

// WithLogger 日志
func WithLogger(logger *logger.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithAccessKey access-key 鉴权
func WithAccessKey(accessKey string) Option {
	return func(o *options) {
		o.accessKey = accessKey
	}
}

// WithSecretKey secret-key 鉴权
func WithSecretKey(secretKey string) Option {
	return func(o *options) {
		o.secretKey = secretKey
	}
}

// WithServerPath nacos server addr
func WithServerPath(path string) Option {
	return func(o *options) {
		o.serverPath = path
	}
}

// WithRunEnv current - env online or testing
func WithRunEnv(runEnv string) Option {
	return func(o *options) {
		o.runEnv = runEnv
	}
}

// WithUserName nacos username
func WithUserName(userName string) Option {
	return func(o *options) {
		o.userName = userName
	}
}

// WithPassword users password
func WithPassword(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

type Client struct {
	cli      config_client.IConfigClient
	opts     *options
	changeCh chan string
}

func NewClient(opts ...Option) *Client {
	o := &options{
		timeOut: 5000,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.namespace == "" || o.product == "" || o.name == "" {
		panic("service undefined")
	}
	clientConfig := constant.ClientConfig{
		TimeoutMs:           o.timeOut,
		NamespaceId:         o.namespace,
		AppName:             fmt.Sprintf("%s.%s", o.product, o.name),
		NotLoadCacheAtStart: true,
	}
	if o.secretKey != "" {
		clientConfig.SecretKey = o.secretKey
	}
	if o.accessKey != "" {
		clientConfig.AccessKey = o.accessKey
	}
	if o.logger != nil {
		clientConfig.CustomLogger = &Logger{
			Logger: o.logger,
		}
	}
	if o.userName != "" {
		clientConfig.Username = o.userName
	}
	if o.password != "" {
		clientConfig.Password = o.password
	}
	us, err := url.Parse(o.serverPath)
	if err != nil {
		panic(err)
	}
	if us.Scheme == "" || us.Host == "" {
		panic("nacos-server config center path error")
	}
	ss := strings.Split(us.Host, ":")
	if len(ss) != 2 {
		panic("nacos-server config center host error")
	}
	addr := ss[0]
	port, err := strconv.ParseInt(ss[1], 10, 64)
	if err != nil {
		panic(err)
	}
	if us.Path == "" {
		us.Path = "/nacos"
	} else {
		us.Path = strings.TrimRight(us.Path, "/")
	}
	serverConfig := []constant.ServerConfig{
		{
			Scheme:      us.Scheme,
			IpAddr:      addr,
			Port:        uint64(port),
			ContextPath: us.Path,
		},
	}
	cli, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfig,
	})
	if err != nil {
		panic(err)
	}
	return &Client{cli: cli, opts: o, changeCh: make(chan string)}
}

func (c *Client) onChange(namespace, group, dataId, data string) {
	c.changeCh <- data
}

func (c *Client) ListenConfig() chan string {
	return c.changeCh
}

func (c *Client) GetJsonConfig(name string, v interface{}) (string, error) {
	vs, err := c.cli.GetConfig(vo.ConfigParam{
		DataId: name,
		Group:  c.opts.runEnv,
		Type:   vo.JSON,
	})
	if err != nil {
		return "", err
	}
	if len(vs) == 0 {
		return "", errors.New("config file does not exist")
	}
	if err = c.cli.ListenConfig(vo.ConfigParam{
		DataId:   name,
		Group:    c.opts.runEnv,
		Type:     vo.JSON,
		OnChange: c.onChange,
	}); err != nil {
		return "", err
	}
	return vs, json.Unmarshal([]byte(vs), v)
}

func (c *Client) GetYamlConfig(name string, v interface{}) (string, error) {
	vs, err := c.cli.GetConfig(vo.ConfigParam{
		DataId: name,
		Group:  c.opts.runEnv,
		Type:   vo.YAML,
	})
	if err != nil {
		return "", err
	}
	if len(vs) == 0 {
		return "", errors.New("config file does not exist")
	}
	if err = c.cli.ListenConfig(vo.ConfigParam{
		DataId:   name,
		Group:    c.opts.runEnv,
		Type:     vo.YAML,
		OnChange: c.onChange,
	}); err != nil {
		return "", err
	}
	return vs, yaml.Unmarshal([]byte(vs), v)
}

func (c *Client) GetXmlConfig(name string, v interface{}) (string, error) {
	vs, err := c.cli.GetConfig(vo.ConfigParam{
		DataId: name,
		Group:  c.opts.runEnv,
		Type:   vo.XML,
	})
	if err != nil {
		return "", err
	}
	if len(vs) == 0 {
		return "", errors.New("config file does not exist")
	}
	if err = c.cli.ListenConfig(vo.ConfigParam{
		DataId:   name,
		Group:    c.opts.runEnv,
		Type:     vo.XML,
		OnChange: c.onChange,
	}); err != nil {
		return "", err
	}
	return vs, xml.Unmarshal([]byte(vs), v)
}
