package generator

import (
	"sync/atomic"
)

var connId int64

func NextConnId() int64 {
	return atomic.AddInt64(&connId, 1)
}
