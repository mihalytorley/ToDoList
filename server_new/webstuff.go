package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

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

func add(x, y int) int {
    return x + y
}

var tmpl = template.Must(
    template.New("index.html").
        Funcs(template.FuncMap{"add": add}).
        ParseFiles("templates/index.html"),
)

func taskHandler(actor Actor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        start_handler := time.Now()
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

        //post Method
        if r.Method == http.MethodPost {
            logger.InfoContext(ctx, "Post Method")
            //Catching request
            task := &Task{}
            err := json.NewDecoder(r.Body).Decode(task)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            fmt.Println("got task:", task)

            actor.Send(*task)
            elapsed_handler := time.Since(start_handler)
            fmt.Println("\nTime elapsed handler: ", elapsed_handler)
        }

        w.WriteHeader(http.StatusCreated)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    filename := "task_list.csv"
    todos, err := readCSVFile(filename)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if err := tmpl.Execute(w, todos); err != nil {
        log.Printf("template execute error: %v\n", err)
        http.Error(w, "Internal Server Error", 500)
        return
    }
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    filename := "task_list.csv"
    todos, err := readCSVFile(filename)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}