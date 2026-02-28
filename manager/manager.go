package manager

import (
	"cube/task"
	"errors"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending      queue.Queue
	TaskDManage  map[string][]*task.Task
	EventDManage map[string][]*task.TaskEvent
	Workers      []string

	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}

func (m *Manager) SelectWorker() error {
	return errors.New("not implement")
}

func (m *Manager) UpdateTasks() error {
	return errors.New("not implement")
}
func (m *Manager) SendWork() error {
	return errors.New("not implement")
}
