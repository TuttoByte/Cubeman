package main

import (
	"cube/task"
	"fmt"

	"github.com/docker/docker/client"
)

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}


	dc, _ :=client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dc,
		Config: c,
	}


	result := d.Run()
	if result.Error != nil{
		fmt.Println(result.Error)
		return nil, nil
	}


	fmt.Printf(
		"Container %s is running\n", result.ContainerId,
	)


	return &d, &result
}
