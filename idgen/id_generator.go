package idgen

import (
	"sync/atomic"
)

type IDGenerator struct {
	counter uint64
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		counter: 0,
	}
}

func (g *IDGenerator) Generate() uint64 {

	count := atomic.AddUint64(&g.counter, 1)

	return count
}

func (g *IDGenerator) Reset() {
	atomic.StoreUint64(&g.counter, 0)
}
