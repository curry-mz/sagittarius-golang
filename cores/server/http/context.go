package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"

	gErrors "github.com/curry-mz/sagittarius-golang/cores/errors"

	"github.com/go-playground/form/v4"
	"github.com/pkg/errors"
)

type core func(*Context)

type Context struct {
	ctx      context.Context
	r        *http.Request
	rp       *http.Response
	w        http.ResponseWriter
	h        http.Handler
	reqBody  []byte
	reqData  interface{}
	respData interface{}

	cores []core
	index int8
	srv   *Engine
}

func newContext() *Context {
	c := &Context{
		index:    0,
		r:        nil,
		w:        nil,
		cores:    nil,
		srv:      nil,
		respData: nil,
		reqData:  nil,
		reqBody:  nil,
		ctx:      context.TODO(),
	}
	return c
}

func (c *Context) reset() *Context {
	c.index = 0
	c.cores = nil
	c.srv = nil
	c.ctx = context.TODO()
	c.respData = nil
	c.reqData = nil
	c.reqBody = nil
	return c
}

func (c *Context) do() {
	for c.index < int8(len(c.cores)) {
		c.cores[c.index](c)
		c.index++
	}
}

func (c *Context) Ctx() context.Context {
	return c.ctx
}

func (c *Context) Writer() http.ResponseWriter {
	return c.w
}

func (c *Context) WithValue(key any, value any) {
	c.ctx = context.WithValue(c.ctx, key, value)
}

func (c *Context) Handler() http.Handler {
	return c.h
}

func (c *Context) Body() []byte {
	return c.reqBody
}

// SetBody 入参格式和header的Content-Type保持一只
func (c *Context) SetBody(body []byte) {
	c.reqBody = body
}

func (c *Context) Request() *http.Request {
	return c.r
}

func (c *Context) Response() *http.Response {
	return c.rp
}

func (c *Context) ResponseData() interface{} {
	return c.respData
}

func (c *Context) Path() string {
	return c.r.URL.Path
}

func (c *Context) Abort() {
	c.index = int8(len(c.cores))
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.cores)) {
		c.cores[c.index](c)
		c.index++
	}
}

func (c *Context) Bind(atom interface{}, v interface{}) error {
	// 解析query
	queryBinder := _binders["query"]
	if queryBinder != nil {
		if atom != nil {
			err := queryBinder.Unmarshal(c, atom)
			if err != nil {
				return err
			}
		}
		if v != nil {
			err := queryBinder.Unmarshal(c, v)
			if err != nil {
				return err
			}
		}
	}
	// 解析body
	if len(c.reqBody) > 0 {
		var accept string
		for _, ct := range c.Request().Header["Content-Type"] {
			if ct != "" {
				accept = ct
				break
			}
		}
		if accept == "" {
			return errors.New("request head Content-Type not support")
		}
		if binder, has := _binders[accept]; has {
			err := binder.Unmarshal(c, v)
			if err != nil {
				return err
			}
		}
	}
	if v != nil {
		c.reqData = v
	}
	return nil
}

func (c *Context) HttpError(code int, message string) error {
	c.w.Header().Add("Content-Type", "text/plain")
	c.w.WriteHeader(code)

	c.buildRespData(code, -1, message, nil)
	_, err := c.w.Write([]byte(message))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) JsonCustom(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/json")
	c.w.WriteHeader(http.StatusOK)

	if data == nil {
		return nil
	}
	c.respData = data
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) FormCustom(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/x-www-form-urlencoded")
	c.w.WriteHeader(http.StatusOK)

	if data == nil {
		return nil
	}
	c.respData = data
	e := form.NewEncoder()
	vs, err := e.Encode(data)
	if err != nil {
		return err
	}
	_, err = c.w.Write([]byte(vs.Encode()))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) XmlCustom(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/xml")
	c.w.WriteHeader(http.StatusOK)

	if data == nil {
		return nil
	}
	c.respData = data
	bs, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) JsonOK(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/json")
	c.w.WriteHeader(http.StatusOK)

	body := map[string]interface{}{
		"status":  0,
		"message": "success",
	}
	if data != nil {
		body["data"] = data
	}
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, 0, "", data)
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) FormOK(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/x-www-form-urlencoded")
	c.w.WriteHeader(http.StatusOK)

	body := map[string]interface{}{
		"status":  0,
		"message": "success",
	}
	if data != nil {
		body["data"] = data
	}
	e := form.NewEncoder()
	vs, err := e.Encode(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, 0, "", data)
	_, err = c.w.Write([]byte(vs.Encode()))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) XmlOK(data interface{}) error {
	c.w.Header().Add("Content-Type", "application/xml")
	c.w.WriteHeader(http.StatusOK)

	body := map[string]interface{}{
		"status":  0,
		"message": "success",
	}
	if data != nil {
		body["data"] = data
	}
	bs, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, 0, "", data)
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) JsonErr(err error) error {
	c.w.Header().Add("Content-Type", "application/json")
	c.w.WriteHeader(http.StatusOK)

	ge := gErrors.Cause(err)
	body := map[string]interface{}{
		"status":  ge.Code(),
		"message": ge.Message(),
	}
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, ge.Code(), ge.Message(), nil)
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) FormErr(err error) error {
	c.w.Header().Add("Content-Type", "application/x-www-form-urlencoded")
	c.w.WriteHeader(http.StatusOK)

	ge := gErrors.Cause(err)
	body := map[string]interface{}{
		"status":  ge.Code(),
		"message": ge.Message(),
	}
	e := form.NewEncoder()
	vs, err := e.Encode(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, ge.Code(), ge.Message(), nil)
	_, err = c.w.Write([]byte(vs.Encode()))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) XmlErr(err error) error {
	c.w.Header().Add("Content-Type", "application/xml")
	c.w.WriteHeader(http.StatusOK)

	ge := gErrors.Cause(err)
	body := map[string]interface{}{
		"status":  ge.Code(),
		"message": ge.Message(),
	}
	bs, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	c.buildRespData(http.StatusOK, ge.Code(), ge.Message(), nil)
	_, err = c.w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) buildRespData(httpCode int, status int, message string, body interface{}) {
	data := map[string]interface{}{
		"httpCode": httpCode,
	}
	if httpCode != http.StatusOK {
		data["message"] = message
	} else {
		bd := map[string]interface{}{
			"status":  status,
			"message": message,
		}
		if status == 0 && body != nil {
			bd["data"] = body
		}
		data["body"] = bd
	}
	c.respData = data
}
