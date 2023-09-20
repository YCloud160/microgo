package generator

import "sync/atomic"

var requestId uint32

func NextRequestId() uint32 {
	return atomic.AddUint32(&requestId, 1)
}
