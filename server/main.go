package main

import (
	"fmt"
	"log"
	"os"

	"github.com/docker/docker/client"
)

var cli *client.Client

func main() {
	var err error

	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Connection with Docker failed:", err)
		os.Exit(1)
	}
	fmt.Println("connection completed")

}
