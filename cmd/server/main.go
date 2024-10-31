package main

import (
	"flag"
	"fmt"
	"github.com/xanderkhrenov/dynamic-wp/pkg/workerpool"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/xanderkhrenov/dynamic-wp/internal/config"
	"github.com/xanderkhrenov/dynamic-wp/internal/handler"
)

var configPath = flag.String("configPath", "./config/config.json", "Path to config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error creating listener: %v", err)
	}

	fmt.Printf("server started on %s\n", addr)
	defer fmt.Println("server stopped")

	wp := workerpool.NewWorkerPool()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, os.Interrupt)

	addHandlerChan := make(chan struct{})
	go handler.AddHandler(listener, wp, addHandlerChan)

	<-wait
	listener.Close()
	<-addHandlerChan

	wp.Shutdown()
}
