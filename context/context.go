package context

import "context"

type (
	serverTransportKey struct{}
	clientTransportKey struct{}
	forContextKey      struct{}
)

func ForCtx(ctx context.Context) context.Context {
	return context.WithValue(context.Background(), forContextKey{}, ctx)
}

func AsCtx(ctx context.Context) context.Context {
	if ctx.Value(forContextKey{}) == nil {
		return ctx
	}
	if _, ok := ctx.Value(forContextKey{}).(context.Context); !ok {
		return ctx
	}
	return ctx.Value(forContextKey{}).(context.Context)
}

type TransData struct {
	Endpoint    string `json:"host"`
	Namespace   string `json:"namespace"`
	Product     string `json:"product"`
	ServiceName string `json:"serviceName"`
}

func NewServerContext(ctx context.Context, td TransData) context.Context {
	return context.WithValue(ctx, serverTransportKey{}, td)
}

func FromServerContext(ctx context.Context) (TransData, bool) {
	td, ok := ctx.Value(serverTransportKey{}).(TransData)
	return td, ok
}

func NewClientContext(ctx context.Context, td TransData) context.Context {
	return context.WithValue(ctx, clientTransportKey{}, td)
}

func FromClientContext(ctx context.Context) (TransData, bool) {
	td, ok := ctx.Value(clientTransportKey{}).(TransData)
	return td, ok
}
