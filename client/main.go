package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

type Task struct {
    Name string   `json:"name"`
    Status string `json:"status"`
}

type contextKey int

const (
    traceCtxKey contextKey = iota + 1
)

type MyHandler struct {
    slog.Handler
}

func main() {
	c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Exitted  before process finished")
        os.Exit(0)
    }()
    // Logger and context setup
    id := uuid.New()
    var handler slog.Handler
    handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
        AddSource: true,
    })
    handler = &MyHandler{handler}
    slog.SetDefault(slog.New(handler))
    ctx := context.Background()
    ctx = context.WithValue(ctx, traceCtxKey, id.String())

    logger := slog.With()

	flag.Parse()
	
	if err := client(); err != nil {
		logger.ErrorContext(ctx, "Error sending task to server")
	}
}

func client() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Prompt user for command
        fmt.Print("Input name of task and task status like so: name, status: \n")
        
        // Read a line from input
        input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		name := strings.ReplaceAll(strings.TrimSpace(strings.Split(input, ",")[0]), " ", "_")
		status := strings.ReplaceAll(strings.TrimSpace(strings.Split(input, ",")[1]), " ", "_")

		if len(strings.Split(input, ",")) > 2 {
			fmt.Print("In the right format plz \n")
			continue
		}

		if name != "" && status != "" {
			task := &Task{
				Name: name,
				Status: status,
			}

			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(task)
			if err != nil {
				return err
			}

			resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			fmt.Println(resp.Status)
		} else {
			resp, err := http.Get("http://localhost:8080/todos")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			fmt.Println(resp.Status)
		}
	}
}