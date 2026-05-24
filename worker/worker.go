package worker

import (
	"context"
	"cube/task"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/moby/term"
	"go/types"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	DWatch    map[uuid.UUID]*task.Task
	Stats     *Stats
	TaskCount int
}

func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Printf("No tasks in queue")
		return task.DockerResult{
			Error: nil,
		}
	}

	taskInQueue := t.(task.Task)

	taskIn := w.DWatch[taskInQueue.ID]
	if taskIn == nil {
		taskIn = &taskInQueue
		w.DWatch[taskInQueue.ID] = &taskInQueue
	}

	var result task.DockerResult
	if task.ValidStateTransition(taskIn.State, taskInQueue.State) {
		switch taskInQueue.State {
		case task.Scheduled:
			result = w.StartTask(taskInQueue)
		case task.Completed:
			result = w.StopTask(taskInQueue)
		default:
			result.Error = errors.New("Error usage of run")
		}
	} else {
		err := fmt.Errorf("Invalid transition from %v to %v",
			taskIn.State, taskInQueue.State)
		result.Error = err
	}
	return result

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

func (w *Worker) GetTasks() []*task.Task {
	var tasks []*task.Task

	for _, v := range w.DWatch {
		tasks = append(tasks, v)
	}

	return tasks
}

func (w *Worker) CollectStats(ctx context.Context) {
	for {
		select {
		case <-time.After(15 * time.Second):
			log.Printf("Colelcting Stats")
			w.Stats = GetStats()
			w.Stats.TotalTaskCount = w.TaskCount
		case <-ctx.Done():
			log.Fatal("Context Error")
		}
	}
}

func (w *Worker) InspectTask(t task.Task) task.DockerInspectResponse {
	config := task.NewConfig(&t)
	d, err := task.NewDocker(config)
	if err != nil {
		return task.DockerInspectResponse{
			Error: err,
		}
	}
	return d.Inspect(t.ContainerId)
}

func (w *Worker) UpdateTask(t task.Task) {
	for {
		log.Printf("Updating task %v", t.ID)
		w.
	}
}


func(w *Worker) updateTask(){
	for id ,t := range w.DWatch{
		if t.State == task.Running{
			resp := w.InspectTask(*t)
			if resp.Error != nil {
				fmt.Printf("Error inspecting task %v: %v\n", t.ID, resp.Error)
			}


			if resp.Container == nil{
				log.Printf("No container found for task %v\n", t.ID)
				w.DWatch[id].State = task.Failed
			}

			if resp.Container.State.Status == "exited"{
				log.Printf("Container for task %s in non-running state %s", id, resp.Container.State.Status)
				w.DWatch[id].State = task.Failed
			}

			w.DWatch[id].HostPorts =  resp.Container.NetworkSettings.NetworkSettingsBase.Ports
		}
	}
}
