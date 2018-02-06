package main

import (
	"context"
	"log"

	"github.com/llitfkitfk/containerd"
)

func main() {
	client, err := containerd.New("./var/run/docker/containerd/docker-containerd.sock")
	if err != nil {
		log.Fatal("failed to new client: ", err)
	}
	defer client.Close()
	v, err := client.Version(context.Background())
	if err != nil {
		log.Fatal("failed to get version: ", err)
	}
	log.Println(v)
}
