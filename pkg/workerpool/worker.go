package workerpool

import (
	"context"
	"fmt"
	"sync"
)

type worker struct {
	id int
}

func newWorker(id int) *worker {
	return &worker{id}
}
func (w *worker) run(ctx context.Context, tasks chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d stopped\n", w.id)
			return
		case task, found := <-tasks:
			if !found {
				fmt.Printf("worker %d stopped\n", w.id)
				return
			}
			reply := w.processTask(task)
			fmt.Println(reply)
		}
	}
}

func (w *worker) processTask(task string) string {
	return fmt.Sprintf("worker %d processed task: %s", w.id, task)
}
