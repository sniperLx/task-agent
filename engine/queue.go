package engine

import "sync"

type Queue interface {
	//返回队首元素，并从队列删除
	Poll() Task

	//添加元素到队尾
	Offer(task Task)

	//返回队列长度
	Size() int

	//清空队列
	Clear()
}

type blockingTaskQueue struct {
	queue    []Task
	size     int
	rwLock   sync.RWMutex
	waitChan chan bool
}

const initialCapacity = 16

func NewBlockingQueue() *blockingTaskQueue {
	return &blockingTaskQueue{
		queue:    make([]Task, 0, initialCapacity),
		size:     0,
		rwLock:   sync.RWMutex{},
		waitChan: make(chan bool),
	}
}

func (bq *blockingTaskQueue) Poll() Task {
	if len(bq.queue) == 0 {
		<-bq.waitChan
	}

	bq.rwLock.Lock()
	firstEle := bq.queue[0]
	bq.queue = bq.queue[1:]
	bq.size--
	bq.rwLock.Unlock()

	return firstEle
}

func (bq *blockingTaskQueue) Offer(ele Task) {
	bq.rwLock.Lock()
	bq.queue = append(bq.queue, ele)
	bq.size++
	bq.rwLock.Unlock()
	bq.waitChan <- true
}

func (bq *blockingTaskQueue) Size() int {
	return bq.size
}

func (bq *blockingTaskQueue) Clear() {
	bq.rwLock.Lock()
	bq.queue = make([]Task, 0, initialCapacity)
	bq.rwLock.Unlock()
}
