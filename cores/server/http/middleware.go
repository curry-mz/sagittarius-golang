package http

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	gCtx "code.cd.local/sagittarius/sagittarius-golang/context"
	"code.cd.local/sagittarius/sagittarius-golang/cores/logger"

	"github.com/getsentry/sentry-go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

///////////////////////////////////////////
// 服务端中间件
///////////////////////////////////////////

func PanicHandler(lgr *logger.Logger) core {
	return func(c *Context) {
		defer func() {
			var rerr interface{}
			if rerr = recover(); rerr != nil {
				var buf [1 << 10]byte
				runtime.Stack(buf[:], true)
				lgr.Error(c.Ctx(), "http error, message:%v\n, stack:%s", rerr, string(buf[:]))

				hub := sentry.CurrentHub().Clone()
				hub.CaptureException(errors.New(string(buf[:])))
				hub.Flush(5 * time.Second)
			}
		}()
		c.Next()
	}
}

func TracingHandler(tracer opentracing.Tracer) core {
	return func(c *Context) {
		spanContext, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request().Header),
		)
		var opts []opentracing.StartSpanOption
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			opts = append(opts, opentracing.Tag{Key: string(ext.Component), Value: "http Server"})
		} else {
			opts = append(opts, opentracing.ChildOf(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "http Server"})
		}
		span := tracer.StartSpan(
			c.Request().URL.String(),
			opts...,
		)
		defer span.Finish()

		c.ctx = opentracing.ContextWithSpan(c.ctx, span)
		// context 写入上游服务信息
		sk := gCtx.GetUberHttpHeader(c.Request().Header)
		if sk != "" {
			ss := strings.Split(sk, ".")
			c.ctx = gCtx.NewClientContext(c.ctx, gCtx.TransData{
				Endpoint:    c.Request().RemoteAddr,
				Namespace:   ss[0],
				Product:     ss[1],
				ServiceName: strings.Join(ss[2:], "."),
			})
		}
		c.Next()
	}
}

func LogHandler(lgr *logger.Logger, requestEnable bool) core {
	return func(c *Context) {
		// 获取远端服务信息
		td, ok := gCtx.FromClientContext(c.ctx)
		if !ok {
			td = gCtx.TransData{}
		}
		// start时间
		start := time.Now().UnixMilli()

		defer func() {
			logData := map[string]interface{}{
				"Peer":   td,
				"Method": c.Request().URL.String(),
				"Cost":   fmt.Sprintf("%dms", time.Now().UnixMilli()-start),
			}
			if requestEnable {
				logData["Request"] = c.reqData
			}
			if c.respData != nil {
				logData["Response"] = c.respData
			}
			bs, e := json.Marshal(logData)
			if e != nil {
				return
			}
			lgr.Write(c.Ctx(), "%s", string(bs))
		}()
		c.Next()
	}
}
