package workerpool

import (
	"context"
	"fmt"
	"sync"
)

type WorkerPool struct {
	mu      *sync.RWMutex
	nextID  int
	cancels map[int]context.CancelFunc
	wg      *sync.WaitGroup
	tasks   chan string
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		mu:      &sync.RWMutex{},
		nextID:  1,
		cancels: make(map[int]context.CancelFunc),
		wg:      &sync.WaitGroup{},
		tasks:   make(chan string),
	}
}

func (wp *WorkerPool) AddTask(task string) error {
	wp.mu.RLock()
	activeWorkers := len(wp.cancels)
	wp.mu.RUnlock()

	if activeWorkers == 0 {
		return fmt.Errorf("no workers to process task")
	}
	wp.tasks <- task
	return nil
}

func (wp *WorkerPool) AddWorker() (workerID int) {
	ctx, cancel := context.WithCancel(context.Background())
	wp.mu.Lock()
	defer wp.mu.Unlock()

	workerID = wp.nextID
	wp.cancels[wp.nextID] = cancel
	wp.nextID++

	wp.wg.Add(1)
	go newWorker(workerID).run(ctx, wp.tasks, wp.wg)

	return workerID
}

func (wp *WorkerPool) DeleteWorker(id int) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	cancel, found := wp.cancels[id]
	if !found {
		return fmt.Errorf("there is no worker %d", id)
	}

	cancel()
	delete(wp.cancels, id)
	return nil
}

func (wp *WorkerPool) DeleteAnyWorker() (id int, err error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.cancels) == 0 {
		return 0, fmt.Errorf("no worker to delete")
	}

	var cancel context.CancelFunc
	for id, cancel = range wp.cancels {
		break
	}

	cancel()
	delete(wp.cancels, id)
	return id, nil
}

func (wp *WorkerPool) Shutdown() {
	close(wp.tasks)

	wp.mu.Lock()
	defer wp.mu.Unlock()

	for id, cancel := range wp.cancels {
		cancel()
		delete(wp.cancels, id)
	}
	wp.wg.Wait()
}
