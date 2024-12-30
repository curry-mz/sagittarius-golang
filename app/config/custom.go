package config

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/curry-mz/sagittarius-golang/env"
	"github.com/curry-mz/sagittarius-golang/logger"
	"github.com/curry-mz/sagittarius-golang/nacos"
)

type Custom struct {
	cli  *nacos.Client
	name string
}

func New(ns string, pd string, sn string, name string) (*Custom, error) {
	path, accessKey, secretKey, _, userName, password := env.GetNacos()
	if path == "" {
		return nil, errors.New("nacos-server config center path undefined")
	}
	// 创建nacos客户端
	ncopts := []nacos.Option{
		nacos.WithNamespace(ns),
		nacos.WithProduct(pd),
		nacos.WithName(sn),
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
	cli := nacos.NewClient(ncopts...)
	return &Custom{
		cli:  cli,
		name: name,
	}, nil
}

func (c *Custom) GetJsonConfig(ctx context.Context, v interface{}) error {
	_, err := c.cli.GetJsonConfig(c.name, v)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-c.cli.ListenConfig():
				_ = json.Unmarshal([]byte(s), v)
			}
		}
	}()
	return nil
}
