package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/u007/dockmetric/configuration"
	"github.com/u007/dockmetric/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("dockmetric containers-id")
		fmt.Println("eg: dockmetric app_web_1,[...]")
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Println(".env ignored / error")
	}

	containerNames := os.Args[1]
	containers := strings.Split(containerNames, ",")
	log.Println("Containers: ", containers)
	// goPath := os.Getenv("GOPATH")
	// os.RemoveAll(goPath + "/src/" + repo)
	// fmt.Printf("simulating: go get -u %s@%s\n", repo, version)
	// cmd := exec.Command("go", "get", "-u", repo)
	// cmd.Dir = goPath
	// if err := cmd.Run(); err != nil {
	// 	panic(err)
	// }
	ctx := context.Background()
	logger := configuration.SetupLogging()
	ctx = context.WithValue(ctx, "log", logger)

	for _, container := range containers {
		go func(container string) error {
			return service.RunDockerStats(ctx, container)
		}(container)
	}

	// c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c
	for {
		time.Sleep(2 * time.Second)
	}

	// fmt.Printf("Installed %s@%s\n", repo+subPath, version)
}
