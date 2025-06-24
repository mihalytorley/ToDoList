package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

type Task struct {
    Name string   `json:"name"`
    Status string `json:"status"`
}

var (
    name *string = flag.String("name", "", "a string describing the name of the task")
    status *string = flag.String("status", "", "a string describing if the task is either not started, started, or completed")
)

type contextKey int

const (
    traceCtxKey contextKey = iota + 1
)

type MyHandler struct {
    slog.Handler
}

func main() {
	c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Exitted  before process finished")
        os.Exit(1)
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
	
	if err := client(*name, *status); err != nil {
		logger.ErrorContext(ctx, "Error sending task to server")
	}
}

func client(name string, status string) error {
	task := &Task{
		Name: name,
		Status: status,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(task)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/", "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
	return nil
}