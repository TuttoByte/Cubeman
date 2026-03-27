package worker

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	tEvent := task.TaskEvent{}

	err := d.Decode(&tEvent)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshling body : %v\n", err)
		log.Printf(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Worker.AddTask(tEvent.Task)
	log.Printf("Added task %v\n", tEvent.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(tEvent.Task)

}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	if taskID == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)
	}

	tID, _ := uuid.Parse(taskID)

	_, ok := a.Worker.DWatch[tID]
	if !ok {
		log.Printf("No task with provided ID %v", tID)
	}
	taskToStop := a.Worker.DWatch[tID]
	tCopy := *taskToStop
	tCopy.State = task.Completed
	a.Worker.AddTask(tCopy)
	log.Printf("Added tasl %v to stop container %v", taskToStop.ID, taskToStop.ContainerId)
	w.WriteHeader(204)

}

func (a *Api) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.Stats)
}
