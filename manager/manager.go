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
	TaskDManage  map[uuid.UUID]*task.Task
	EventDManage map[uuid.UUID]*task.TaskEvent
	Workers      []string

	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string

	LastWorker int

	WorkerNodes []*node.Node
	Scheduler   scheduler.Scheduler
}

func NewManager(workers []string) *Manager {
	taskDbManager := make(map[uuid.UUID]*task.Task)
	eventDbManager := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)
	for wk := range workers {
		workerTaskMap[workers[wk]] = []uuid.UUID{}
	}
	return &Manager{
		TaskDManage:   taskDbManager,
		EventDManage:  eventDbManager,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
	}
}

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) UpdateTasks() {
	for _, worker := range m.Workers {
		log.Printf("Checkimg worker %v for task updates", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error connecton to %v: %v", worker, err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request to %v: %v", worker, resp.Status)
		}

		d := json.NewDecoder(resp.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Printf("Error decoding response from %v: %v", worker, err)
		}

		for _, t := range tasks {
			log.Printf("Attemp to update task %v", t)

			_, ok := m.TaskDManage[t.ID]
			if !ok {
				log.Printf("Task with ID %s not fount", t.ID)
				return
			}

			if m.TaskDManage[t.ID].State != t.State {
				m.TaskDManage[t.ID].State = t.State
			}

			m.TaskDManage[t.ID].StartTime = t.StartTime
			m.TaskDManage[t.ID].EndTime = t.EndTime
			m.TaskDManage[t.ID].ContainerId = t.ContainerId
		}
	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		w := m.SelectWorker()

		e := m.Pending.Dequeue()
		te := e.(task.TaskEvent)
		t := te.Task
		log.Printf("Pulled %v off pending queue\n", t)

		m.EventDManage[te.ID] = &te
		m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], te.Task.ID)
		m.TaskWorkerMap[t.ID] = w

		t.State = task.Scheduled
		m.TaskDManage[t.ID] = &t

		data, err := json.Marshal(te)
		if err != nil {
			log.Printf("unable to marshal task event: %v", err)
		}

		url := fmt.Sprintf("htpp://%s/task", w)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", w, err)
			m.Pending.Enqueue(te)
			return
		}

		d := json.NewDecoder(resp.Body)
		if resp.StatusCode != http.StatusOK {
			e := worker.ErrResponse{}
			err := d.Decode(&e)
			if err != nil {
				fmt.Printf("Error decoding response: %v\n", err)
				return
			}
			log.Printf("Response error: %v\n", e)
			return
		}
		log.Printf("Response status: %v\n", resp.Status)

	} else {
		log.Printf("No work in the queue")
	}

}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
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
