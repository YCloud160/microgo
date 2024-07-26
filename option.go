package microgo

type Option func()

func WithBeforeClose(f func()) Option {
	return func() {
		beforeClose = f
	}
}
