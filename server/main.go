package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

type App struct {
	dockerConn *client.Client
	Router     *gin.Engine
}

func (app *App) initialize() {

}

func main() {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Connection with Docker failed:", err)
		os.Exit(1)
	}
	fmt.Println("Connection to Docker completed.")

	test := "https://github.com/shreyash-209/t1.git"
	CpyCode(test)
	fmt.Println("Repo cloned:", test)

	hostPath, err := filepath.Abs("code-storage/t1/vite-project")
	if err != nil {
		log.Fatal("Error getting absolute path:", err)
	}
	containerPath := "/app"

	err = runContainer(cli, hostPath, containerPath)
	if err != nil {
		log.Fatal("Failed to run container:", err)
	}
}

func runContainer(cli *client.Client, hostPath, containerPath string) error {
	ctx := context.Background()

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:      "node:18",
			Cmd:        []string{"sh", "-c", "npm install && npm run build"},
			WorkingDir: containerPath,
		},
		&container.HostConfig{
			Binds: []string{hostPath + ":" + containerPath},
		},
		nil, nil, "",
	)
	if err != nil {
		return fmt.Errorf("error creating container: %v", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %v", err)
	}

	fmt.Println("Container started successfully:", resp.ID)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case <-statusCh:
		fmt.Println("Container finished execution.")
	case err := <-errCh:
		return fmt.Errorf("error waiting for container: %v", err)
	}

	err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	if err != nil {
		log.Printf("Warning: could not remove container: %v\n", err)
	}

	return nil
}
