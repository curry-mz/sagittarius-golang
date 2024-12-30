package balancer

import (
	"context"

	"code.cd.local/sagittarius/sagittarius-golang/cores/registry"
)

type Balancer interface {
	Pick(context.Context) (*registry.Service, error)
	Update(context.Context, []*registry.Service)
}

type Builder interface {
	Build() Balancer
}
