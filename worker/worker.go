package worker

import (
	"cube/task"
	"errors"

	"github.com/golang-collections/collections/queue"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	DWatch    map[string]*task.Task
	TaskCount int
}

func (w *Worker) RunTask() error {
	return errors.New("not implemet")
}
func (w *Worker) StartTask() error {
	return errors.New("not implement")
}
func (w *Worker) StopTask() error {
	return errors.New("not implement")
}
func (w *Worker) CollectStats() error {
	return errors.New("not implement")
}
