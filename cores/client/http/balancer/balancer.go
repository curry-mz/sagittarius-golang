package balancer

import (
	"context"

	"github.com/curry-mz/sagittarius-golang/cores/registry"
)

type Balancer interface {
	Pick(context.Context) (*registry.Service, error)
	Update(context.Context, []*registry.Service)
}

type Builder interface {
	Build() Balancer
}
