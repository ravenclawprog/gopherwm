package idgen

import (
	"sync"
	"testing"
)

func TestIDGenerator_Generate_Sequential(t *testing.T) {
	gen := NewIDGenerator()

	// check id is 1
	if got := gen.Generate(); got != 1 {
		t.Errorf("Generate() = %d; want 1", got)
	}

	// check id is 2
	if got := gen.Generate(); got != 2 {
		t.Errorf("Generate() = %d; want 2", got)
	}

	// check id is 3
	if got := gen.Generate(); got != 3 {
		t.Errorf("Generate() = %d; want 3", got)
	}
}

func TestIDGenerator_Generate_Concurrent(t *testing.T) {
	gen := NewIDGenerator()

	const goroutinesCount = 100
	const iterationsPerGoroutine = 1000
	totalExpectedIDs := goroutinesCount * iterationsPerGoroutine

	// unique id and wg
	var generatedIDs sync.Map
	var wg sync.WaitGroup

	// run goroutine
	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				id := gen.Generate()

				_, loaded := generatedIDs.LoadOrStore(id, true)
				if loaded {
					t.Errorf("Found duplicate ID: %d", id)
				}
			}
		}()
	}

	wg.Wait()

	if got := gen.Generate(); got != uint64(totalExpectedIDs+1) {
		t.Errorf("After concurrent test, next ID = %d; want %d", got, totalExpectedIDs+1)
	}
}

func TestIDGenerator_Reset(t *testing.T) {
	gen := NewIDGenerator()

	gen.Generate()
	gen.Generate()
	gen.Generate() // counter is 3

	gen.Reset()

	if got := gen.Generate(); got != 1 {
		t.Errorf("After Reset(), Generate() = %d; want 1", got)
	}
}
