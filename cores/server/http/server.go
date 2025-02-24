package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/curry-mz/sagittarius-golang/cores/crypto"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Option func(*Engine)

func Addr(addr string) Option {
	return func(e *Engine) {
		e.addr = addr
	}
}

func TLS(cfg *tls.Config) Option {
	return func(e *Engine) {
		e.tlsCfg = cfg
	}
}

func Crypto(c crypto.ICrypto) Option {
	return func(e *Engine) {
		e.crypto = c
	}
}

func UseH2C(h2c bool) Option {
	return func(e *Engine) {
		e.UseH2C = h2c
	}
}

func OnStop(f func()) Option {
	return func(e *Engine) {
		e.onStop = f
	}
}

type Engine struct {
	*Group
	*http.Server

	addr   string
	UseH2C bool
	pool   sync.Pool
	tree   trees
	tlsCfg *tls.Config
	crypto crypto.ICrypto
	onStop func()
}

func New(opts ...Option) *Engine {
	e := &Engine{
		tree: newTree(),
	}
	group := &Group{
		svr: e,
	}
	e.Group = group
	e.pool.New = func() interface{} {
		return newContext()
	}
	for _, opt := range opts {
		opt(e)
	}
	e.Server = &http.Server{
		TLSConfig: e.tlsCfg,
		Handler:   e,
	}
	return e
}

func (e *Engine) addRoute(method string, path string, cores ...core) {
	e.tree.addRoute(method, path, cores...)
}

func (e *Engine) NewGroup(basePath string) *Group {
	return e.Group.Group(basePath)
}

func (e *Engine) Handler() http.Handler {
	if !e.UseH2C {
		return e
	}

	h2s := &http2.Server{}
	return h2c.NewHandler(e, h2s)
}

func (e *Engine) Start(ctx context.Context) error {
	e.Server.Addr = e.addr
	var err error
	e.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	if e.tlsCfg != nil {
		err = e.Server.ListenAndServeTLS("", "")
	} else {
		err = e.Server.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (e *Engine) Stop(ctx context.Context) error {
	if e.onStop != nil {
		e.onStop()
	}
	return e.Shutdown(ctx)
}

func (e *Engine) handleHTTPRequest(c *Context) {
	method := c.r.Method
	path := c.r.URL.Path
	if method == "OPTIONS" {
		c.Writer().WriteHeader(http.StatusOK)
		return
	}
	if e.tree[method] == nil {
		_ = c.HttpError(404, "page not found")
		return
	}
	root := e.tree[method]
	if len(path) == 1 && path[0] == '/' {
		c.cores = root.cores
	} else {
		ss := strings.Split(path, "/")
		var ns []string
		for _, s := range ss {
			if s != "" {
				ns = append(ns, s)
			}
		}
		current := e.tree[method]
		for _, s := range ns {
			if _, has := current.children[s]; !has && method != "OPTIONS" {
				_ = c.HttpError(404, "page not found")
				return
			}
			current = current.children[s]
		}
		c.cores = current.cores
	}
	// 提前解析body
	data, err := io.ReadAll(c.Request().Body)
	if err != nil {
		_ = c.HttpError(499, fmt.Sprintf("request body decode error:%v", err.Error()))
		return
	}
	// Reset resp.Body so it can be use again
	c.Request().Body = io.NopCloser(bytes.NewBuffer(data))
	if len(data) != 0 {
		if e.crypto != nil {
			var s string
			s, err = e.crypto.Decrypt(string(data))
			if err != nil {
				_ = c.HttpError(499, fmt.Sprintf("request body decrypt error:%v", err.Error()))
				return
			}
			data = []byte(s)
		}
		c.reqBody = data
	}
	c.do()
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Context)
	c.w = w
	c.r = req
	c.reset()

	e.handleHTTPRequest(c)

	e.pool.Put(c)
}
