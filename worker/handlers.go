package worker

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {

}
