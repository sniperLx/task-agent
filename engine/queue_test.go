package engine

import (
	"fmt"
	"sync"
	"testing"
)

type testTask struct {
	key string
}

func (tt *testTask) GetId() string {
	return tt.key
}

func TestBlockingQueue_Poll(t *testing.T) {
	bq := NewBlockingQueue()
	bq.Offer(&testTask{key: "hello"})
	bq.Offer(&testTask{key: "world"})

	wg := sync.WaitGroup{}
	wg.Add(3)

	fmt.Printf("size: %v\n", bq.Size())

	go func() {
		t.Logf("1: %v\n", bq.Poll())
		wg.Done()
	}()

	go func() {
		t.Logf("2: %v\n", bq.Poll())
		wg.Done()
	}()

	go func() {
		t.Logf("3: %v\n", bq.Poll())
		wg.Done()
	}()

	wg.Wait()
	if bq.Size() != 0 {
		t.Errorf("size should be 0, but get %d\n", bq.Size())
	}
}

func TestBlockingQueue_Clear(t *testing.T) {
	bq := NewBlockingQueue()
	bq.Offer(&testTask{key: "hello"})
	bq.Offer(&testTask{key: "world"})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		bq.Clear()
		wg.Done()
	}()

	wg.Wait()

	if bq.Size() != 0 {
		t.Errorf("after Clear was called, size should be 0, but get %d\n", bq.Size())
	}
}
