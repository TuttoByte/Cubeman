package manager

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Api struct {
	Address string
	Port    int
	Manager *Manager
	Router  *chi.Mux
}

func (a *Api) initRoute() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", func(router chi.Router) {
		router.Post("/", a.StartTaskHandler)
		router.Get("/", a.GetTasksHandler)
		router.Route("/{taskId}", func(router chi.Router) {
			router.Delete("/", a.StopTaskHandler)
		})
	})
}

func (a *Api) Start() {
	a.initRoute()
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}
