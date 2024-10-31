package workerpool

import (
	"context"
	"fmt"
	"sync"
)

type Worker struct {
	id int
}

func NewWorker(id int) *Worker {
	return &Worker{id}
}
func (w *Worker) Run(ctx context.Context, tasks chan string, wg *sync.WaitGroup) {
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
			reply := w.ProcessTask(task)
			fmt.Println(reply)
		}
	}
}

func (w *Worker) ProcessTask(task string) string {
	return fmt.Sprintf("worker %d processed task: %s", w.id, task)
}
