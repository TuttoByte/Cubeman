package worker

import (
	"cube/task"
	"errors"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	DWatch    map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) RunTask() error {
	return errors.New("not implemet")
}
func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	dTask, err := task.NewDocker(config)
	if err != nil {
		return task.DockerResult{
			Error: err,
		}
	}

	result := dTask.Stop(t.ContainerId)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerId, result.Error)
	}

	t.EndTime = time.Now().UTC()
	t.State = task.Completed
	w.DWatch[t.ID] = &t
	log.Printf("Stopped and removed container %v", t.ContainerId)
	return result

}
func (w *Worker) StartTask(t task.Task) task.DockerResult {
	cfg := task.NewConfig(&t)
	dTask, err := task.NewDocker(cfg)
	if err != nil {
		return task.DockerResult{
			Error: err,
		}
	}

	result := dTask.Run()
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.DWatch[t.ID] = &t
		return result
	}

	t.ContainerId = result.ContainerId
	t.State = task.Running
	w.DWatch[t.ID] = &t
	return result

}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}
func (w *Worker) CollectStats() error {
	return errors.New("not implement")
}
