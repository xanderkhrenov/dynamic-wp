package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/xanderkhrenov/dynamic-wp/internal/config"
)

var configPath = flag.String("configPath", "./config/config.json", "Path to config file")

func readServer(conn net.Conn, wait chan struct{}) {
	defer close(wait)

	serverReader := bufio.NewReader(conn)
	for {
		var message string
		message, err := serverReader.ReadString('\n')
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}
		fmt.Print(message)
	}
}

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("error creating connection: %v", err)
		return
	}
	defer conn.Close()

	wait := make(chan struct{})
	go readServer(conn, wait)

	clientReader := bufio.NewReader(os.Stdin)
	for {
		select {
		case _, found := <-wait:
			if !found {
				return
			}
		default:
			command, _ := clientReader.ReadString('\n')
			fmt.Fprintf(conn, command+"\n")
		}
	}
}
