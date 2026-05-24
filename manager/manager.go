package manager

import (
	"bytes"
	"cube/node"
	"cube/scheduler"
	"cube/task"
	"cube/worker"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
	"time"
)

type Manager struct {
	Pending      queue.Queue
	TaskDManage  map[string][]*task.Task
	EventDManage map[string][]*task.TaskEvent
	Workers      []string

	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string

	WorkerNodes []*node.Node
	Scheduler   scheduler.Scheduler
}

func NewManager(workers []string, schedulerType string) *Manager {

	var nodes []*node.Node
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}
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

func (m *Manager) checkTaskHealth(t task.Task) error {
	log.Printf("Calling health check for task %s :%s\n", t.ID, t.HostPorts)

	w := m.TaskWorkerMap[t.ID]
	hostPort := getHostPort(t.HostPorts)

	worker := strings.Split(w, ":")
	url := fmt.Sprintf("http://%s:%s", worker[0], hostPort)

	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Task %s is not healthy", t.ID)
		log.Println(msg)
		return errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Task %s is not healthy", t.ID)
		log.Println(msg)
		return errors.New(msg)
	}

	log.Printf("Task %s health check response: %v\n", t.ID, resp.StatusCode)

	return nil
}

func (m *Manager) GetTasks() []*task.Task {
	tasks := make([]*task.Task, 0)

	for _, t := range m.TaskDManage {
		tasks = append(tasks, t...)
	}
	return tasks
}

func (m *Manager) doHealthCheck() {
	for _, t := range m.GetTasks() {
		if t.State == task.Running && t.RestartCount < 3 {
			err := m.checkTaskHealth(*t)
			if err != nil {
				if t.RestartCount < 3 {
					m.restartTask(t)
				}
			}
		}
	}
}

func (m *Manager) restartTask(t *task.Task) {
	w := m.TaskWorkerMap[t.ID]
	t.State = task.Scheduled
	t.RestartCount++
	m.TaskDManage[t.ID] = t

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Running,
		TimeStamp: time.Now(),
		Task:      *t,
	}

	data, err := json.Marshal(te)
	if err != nil {
		log.Printf("Error marshalling task event: %v", err)
		return
	}

	url := fmt.Sprintf("http://%s/tasks", w)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error restarting task: %v", err)
		m.Pending.Enqueue(t)
		return
	}

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		e := worker.ErrResponse{}
		err := d.Decode(&e)
		if err != nil {
			log.Printf("Error restarting task: %v", err)
			return
		}
		log.Printf("Error restarting task: %v", resp.Status)
		return
	}

	newTask := task.Task{}
	err = d.Decode(&newTask)
	if err != nil {
		log.Printf("Error restarting task: %v", err)
		return
	}
	log.Printf("Restarted task: %v", newTask.ID)
}

func getHostPort(ports nat.PortMap) *string {
	for k, _ := range ports {
		return &ports[k][0].HostPort
	}
	return nil
}
