package context

import (
	"net/http"

	"google.golang.org/grpc/metadata"
)

type Metadata struct {
	metadata.MD
}

func (m Metadata) ForeachKey(handler func(key, val string) error) error {
	for k, values := range m.MD {
		for _, v := range values {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m Metadata) Set(key, val string) {
	m.MD[key] = append(m.MD[key], val)
}

func (m Metadata) Get(key string) string {
	if _, has := m.MD[key]; !has {
		return ""
	}
	if len(m.MD[key]) == 0 {
		return ""
	}
	return m.MD[key][0]
}

const (
	_uberCtxServiceKey = "_uber_ctx_service_key"
)

func GetUberMeta(md Metadata) string {
	return md.Get(_uberCtxServiceKey)
}

func SetUberMeta(md Metadata, sk string) {
	md.Set(_uberCtxServiceKey, sk)
}

func GetUberHttpHeader(h http.Header) string {
	return h.Get(_uberCtxServiceKey)
}

func SetUberHttpHeader(h http.Header, sk string) {
	h.Set(_uberCtxServiceKey, sk)
}
