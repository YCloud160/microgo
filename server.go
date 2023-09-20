package microgo

import "context"

type Server interface {
	Start() error
	Stop() error
	Name() string
}

type Call func(ctx context.Context, impl any, method string, input []byte) (output []byte, err error)
