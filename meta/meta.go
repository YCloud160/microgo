package meta

import "context"

type metaKey struct {
}

func NewOutContext(ctx context.Context, meta map[string]string) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}

func FromOutContext(ctx context.Context) (map[string]string, bool) {
	val := ctx.Value(metaKey{})
	meta, ok := val.(map[string]string)
	if ok {
		return meta, true
	}
	return map[string]string{}, false
}
