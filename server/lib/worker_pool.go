package lib

import (
	"github.com/google/uuid"
	"sync"
	"sync/atomic"
)

type Task[D any] struct {
	ID   uuid.UUID
	Data D
}

type WorkerPool interface {
	Start()
	EnqueueTask(task Task[map[string]any]) error
}

type WorkerWg struct {
	wg     *sync.WaitGroup
	status string
}

type WorkerPoolConfig struct {
	NumWorkers int
	WorkerFn   func(task Task[map[string]any])
}

type WorkerPoolImpl struct {
	workerFn  func(task Task[map[string]any])
	jobQueue  chan Task[map[string]any]
	wg        sync.WaitGroup
	active    int64
	processed int64
	pending   int64

	quitChan   chan struct{}
	workerLock sync.Mutex
	workers    map[uuid.UUID]chan struct{}
}

func NewWorkerPool(config WorkerPoolConfig) *WorkerPoolImpl {
	return &WorkerPoolImpl{
		workerFn: config.WorkerFn,
		jobQueue: make(chan Task[map[string]any], config.NumWorkers),
		workers:  make(map[uuid.UUID]chan struct{}),
		quitChan: make(chan struct{}),
	}
}

func (wp *WorkerPoolImpl) worker(id uuid.UUID, quit chan struct{}) {
	atomic.AddInt64(&wp.active, 1)
	defer func() {
		atomic.AddInt64(&wp.active, -1)
		wp.wg.Done()
	}()

	for {
		select {
		case task := <-wp.jobQueue:
			atomic.AddInt64(&wp.pending, -1)
			atomic.AddInt64(&wp.processed, 1)
			wp.workerFn(task)
			atomic.AddInt64(&wp.processed, -1)
		case <-quit:
			return
		case <-wp.quitChan:
			return
		}
	}
}

func (wp *WorkerPoolImpl) EnqueueTask(task Task[map[string]any]) {
	wp.jobQueue <- task
	atomic.AddInt64(&wp.pending, 1) // Task added to queue
}

func (wp *WorkerPoolImpl) ScaleUp(num int) []uuid.UUID {
	wp.workerLock.Lock()
	defer wp.workerLock.Unlock()
	var ids []uuid.UUID
	for i := 0; i < num; i++ {
		id := uuid.New()
		quit := make(chan struct{})
		wp.workers[id] = quit
		wp.wg.Add(1)
		go wp.worker(id, quit)
		ids = append(ids, id)
	}

	return ids
}

func (wp *WorkerPoolImpl) ScaleDown(workerId uuid.UUID) {
	wp.workerLock.Lock()
	defer wp.workerLock.Unlock()

	quit := wp.workers[workerId]

	atomic.AddInt64(&wp.active, -1)
	wp.wg.Add(-1)
	close(quit)
	delete(wp.workers, workerId)
}

func (wp *WorkerPoolImpl) Shutdown() {
	close(wp.quitChan)
	wp.wg.Wait()
	close(wp.jobQueue)
}

func (wp *WorkerPoolImpl) WorkerCount() int {
	wp.workerLock.Lock()
	defer wp.workerLock.Unlock()
	return len(wp.workers)
}

func (wp *WorkerPoolImpl) GetMetrics() (active, pending, processed int64) {
	return atomic.LoadInt64(&wp.active),
		atomic.LoadInt64(&wp.pending),
		atomic.LoadInt64(&wp.processed)
}
