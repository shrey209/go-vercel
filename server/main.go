package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type App struct {
	dockerConn *client.Client
	Router     *gin.Engine
	s3         *S3Storage
}

type CodeUrl struct {
	URL      string `json:"url"`
	CodePath string `json:"codepath"`
}

func (app *App) initialize() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Connection with Docker failed:", err)
		os.Exit(1)
	}
	app.dockerConn = cli

	app.Router = gin.Default()

	app.Router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	app.Router.POST("/deploy", app.handleDeploy)
}

func (app *App) handleDeploy(c *gin.Context) {
	var out CodeUrl

	if err := c.ShouldBindJSON(&out); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	fmt.Println("Received deployment request:", out)

	log.Println("Git clone started")
	// Ensure CpyCode is implemented
	CpyCode(out.URL)

	hostPath, err := filepath.Abs("code-storage/t1/vite-project")
	if err != nil {
		log.Fatal("Error getting absolute path:", err)
	}

	containerPath := "/app"
	err = runContainer(app.dockerConn, hostPath, containerPath)
	if err != nil {
		log.Println("Failed to run container:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = app.s3.UploadDirectory("code-storage/t1/vite-project/dist")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deployment successful"})
}

func main() {
	app := &App{}
	storage, err := NewS3Storage()
	if err != nil {
		log.Fatal("Initialization failed:", err)
	}
	app.s3 = storage
	app.initialize()

	log.Println("Server started on port 8080")
	if err := app.Router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
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

	if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		log.Printf("Warning: could not remove container: %v\n", err)
	}

	return nil
}
