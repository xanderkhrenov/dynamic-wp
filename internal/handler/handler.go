package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/xanderkhrenov/dynamic-wp/internal/workerpool"
)

const msgInstruct = ""

const (
	msgInvalid  = "invalid command"
	msgNoParams = "command has no parameters"
)

const (
	cmdQuit         = "quit"
	cmdAddTask      = "task"
	cmdAddWorker    = "add"
	cmdDeleteWorker = "del"
)

func AddHandler(listener net.Listener, wp *workerpool.WorkerPool, ch chan<- struct{}) {
	defer close(ch)
	defer close(wp.Tasks())

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		var conn net.Conn
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		} else if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}

		wg.Add(1)
		go handleConn(ctx, conn, wp, wg)
	}
}

func handleConn(ctx context.Context, conn net.Conn, wp *workerpool.WorkerPool, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	name := conn.RemoteAddr().String()
	greet(conn, name)
	defer quit(conn, name)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			command := strings.TrimSpace(scanner.Text())
			if command == "" {
				continue
			}
			fmt.Printf("%s enters: %s\n", name, command)

			if command == cmdQuit {
				return
			}

			var reply string
			switch {
			case strings.HasPrefix(command, cmdAddTask):
				reply = addTask(command, wp)
			case strings.HasPrefix(command, cmdAddWorker):
				reply = addWorker(command, wp)
			case strings.HasPrefix(command, cmdDeleteWorker):
				reply = deleteWorker(command, wp)
			default:
				reply = msgInvalid
			}

			conn.Write([]byte(reply + "\n"))
			fmt.Printf("reply to %s: %s\n", name, reply)
		}
	}
}

func greet(conn net.Conn, name string) {
	fmt.Printf("%+v connected\n", name)
	conn.Write([]byte("Hello, " + name + "\n"))
	conn.Write([]byte(msgInstruct))
}

func quit(conn net.Conn, name string) {
	conn.Write([]byte("connection stopped\n"))
	fmt.Printf("%s disconnected\n", name)
}

func addTask(command string, wp *workerpool.WorkerPool) (reply string) {
	if command == cmdAddTask {
		return "task can not be blank"
	}

	if !strings.HasPrefix(command, cmdAddTask+" ") {
		return msgInvalid
	}

	if wp.ActiveWorkers() == 0 {
		return "no workers to process task, add worker and retry"
	}
	wp.Tasks() <- strings.TrimPrefix(command, cmdAddTask+" ")
	return "successful adding task"
}

func addWorker(command string, wp *workerpool.WorkerPool) (reply string) {
	if command == cmdAddWorker {
		id := wp.AddWorker()
		return fmt.Sprintf("successul adding worker %d", id)
	}

	if !strings.HasPrefix(command, cmdAddWorker+" ") {
		return msgInvalid
	}
	return msgNoParams
}

func deleteWorker(command string, wp *workerpool.WorkerPool) (reply string) {
	if command == cmdDeleteWorker {
		id, err := wp.DeleteAnyWorker()
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("successful deleting worker %d", id)
	}

	if !strings.HasPrefix(command, cmdDeleteWorker+" ") {
		return msgInvalid
	}

	strID := strings.TrimPrefix(command, cmdDeleteWorker+" ")
	id, err := strconv.Atoi(strID)
	if err != nil {
		return fmt.Sprintf("not numeric id %s", strID)
	}

	err = wp.DeleteWorker(id)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("successful deleting worker %d", id)
}
