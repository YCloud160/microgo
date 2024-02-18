package meta

import "context"

type metaRequestKey struct{}

type metaKey struct{}

func NewOutRequestContext(ctx context.Context, meta map[string]string) context.Context {
	return context.WithValue(ctx, metaRequestKey{}, meta)
}

func FromOutRequestContext(ctx context.Context) (map[string]string, bool) {
	val := ctx.Value(metaRequestKey{})
	meta, ok := val.(map[string]string)
	if ok {
		return meta, true
	}
	return map[string]string{}, false
}

func FromOutContext(ctx context.Context) map[string]string {
	val := ctx.Value(metaKey{})
	meta, ok := val.(map[string]string)
	if ok {
		return meta
	}
	return map[string]string{}
}
