package microgo

import "context"

type Server interface {
	Start() error
	Stop() error
	Name() string
	Addr() string
}

type Call func(ctx context.Context, impl any, enc Encoder, method string, input []byte) (output []byte, err error)
