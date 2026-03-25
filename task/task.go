package task

import (
	"context"
	"io"
	"math"
	"os"
	"time"

	container2 "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type Task struct {
	ID            uuid.UUID
	ContainerId   string
	Name          string
	State         State
	Image         string
	Memory        int64
	Disk          int64
	Cpu           float64
	ExposedPorts  nat.PortSet
	PortBind      map[string]string
	RestartPolicy container2.RestartPolicyMode
	StartTime     time.Time
	EndTime       time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	TimeStamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttaschStderr bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy container2.RestartPolicyMode
}

func NewConfig(t *Task) *Config {

	return &Config{
		Name:          t.Name,
		Image:         t.Image,
		Cpu:           t.Cpu,
		Disk:          t.Disk,
		Memory:        t.Memory,
		RestartPolicy: t.RestartPolicy,
		ExposedPorts:  t.ExposedPorts,
	}
}

type Docker struct {
	Client *client.Client
	Config Config
}

func NewDocker(c *Config) (*Docker, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &Docker{
		dc,
		*c,
	}, nil

}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	dockerReader, err := d.Client.ImagePull(
		ctx, d.Config.Image, image.PullOptions{})

	if err != nil {
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, dockerReader)

	rPol := container2.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}

	res := container2.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cConf := container2.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}
	hConf := container2.HostConfig{
		RestartPolicy:   rPol,
		Resources:       res,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(
		ctx, &cConf, &hConf, nil, nil, d.Config.Name)

	if err != nil {
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, container2.StartOptions{})
	if err != nil {
		return DockerResult{Error: err}
	}
	return DockerResult{ContainerId: resp.ID,
		Action: "start",
		Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	ctx := context.Background()

	err := d.Client.ContainerStop(ctx, id, container2.StopOptions{})
	if err != nil {
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, container2.RemoveOptions{
		Force:         false,
		RemoveLinks:   false,
		RemoveVolumes: true,
	})

	if err != nil {
		return DockerResult{Error: err}
	}

	return DockerResult{
		Action: "stop",
		Result: "success",
		Error:  nil,
	}
}
