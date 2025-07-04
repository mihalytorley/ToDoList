// main.go
package main

import (
	//"bytes"
	"context"
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"text/template"

	"github.com/google/uuid"
)

//go:embed templates/index.html
var templateFS embed.FS

type Task struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type contextKey int

const (
	traceCtxKey contextKey = iota + 1
    // ToDo: Don't use enum
    // ToDo: use magic string instead of iota
)

func changeCheck(list_of_lists [][]string, name, status string) [][]string {
	i := 0
    task_change_bool := false
    for _, task := range list_of_lists {
        if slices.Contains(task, name) {
            list_of_lists[i][1] = status
            task_change_bool = true
            if status == "delete" {
                if len(list_of_lists)-1 == i {
                    list_of_lists = list_of_lists[:i]
                } else {
                    list_of_lists = slices.Delete(list_of_lists, i, i+1)
                }
            }
        }
        i++
    }
    if !task_change_bool {
        new_task := []string{name, status}
        list_of_lists = append(list_of_lists, new_task)
    }
    return list_of_lists
}

//=====================================================================
// Template

func add(x, y int) int { return x + y }

var tmpl = template.Must(
    template.New("index.html").
        Funcs(template.FuncMap{"add": add}).
        ParseFS(templateFS, "templates/index.html"),
)

//=====================================================================
// Actor stuff

type TaskRequest struct {
	Response chan<- [][]string
}

type TaskUpdate struct {
	NewTasks [][]string
	Done     chan<- error
}

type TaskChange struct {
    Name   string
    Status string
    Done   chan error
}

type TaskManager struct {
	filename   string
	readCh     chan TaskRequest
	writeCh    chan TaskUpdate
	changeCh chan TaskChange
	shutdownCh chan struct{}
}


func NewTaskManager(filename string) *TaskManager {
	tm := &TaskManager{
		filename:   filename,
		readCh:     make(chan TaskRequest),
		writeCh:    make(chan TaskUpdate),
		changeCh:   make(chan TaskChange),
		shutdownCh: make(chan struct{}),
	}
	go tm.run()
	return tm
}

func (tm *TaskManager) run() {
	var currentTasks [][]string
	tasks, err := tm.loadTasks()
	if err != nil {
		log.Println("Error loading tasks:", err)
		currentTasks = [][]string{}
	} else {
		currentTasks = tasks
	}

	for {
		select {
		case req := <-tm.readCh:
			// Return a copy
			copyTasks := make([][]string, len(currentTasks))
			for i := range currentTasks {
				copyTasks[i] = append([]string(nil), currentTasks[i]...)
			}
			req.Response <- copyTasks

		case upd := <-tm.writeCh:
			currentTasks = upd.NewTasks
			err := tm.saveTasks(currentTasks)
			upd.Done <- err

		case ch := <-tm.changeCh:
    		// apply changeCheck inside the actor
			currentTasks = changeCheck(currentTasks, ch.Name, ch.Status)
			err := tm.saveTasks(currentTasks)
			ch.Done <- err

		case <-tm.shutdownCh:
			return
		}
	}
}

func (tm *TaskManager) GetTasks() ([][]string, error) {
	resp := make(chan [][]string)
	tm.readCh <- TaskRequest{Response: resp}
	return <-resp, nil
}

func (tm *TaskManager) UpdateTasks(newTasks [][]string) error {
	done := make(chan error)
	tm.writeCh <- TaskUpdate{NewTasks: newTasks, Done: done}
	return <-done
}

func (tm *TaskManager) ChangeTask(name, status string) error {
    done := make(chan error)
    tm.changeCh <- TaskChange{Name: name, Status: status, Done: done}
    return <-done
}

func (tm *TaskManager) loadTasks() ([][]string, error) {
	f, err := os.Open(tm.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	var tasks [][]string
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, rec)
	}
	return tasks, nil
}

func (tm *TaskManager) saveTasks(tasks [][]string) error {
	// 1) Open a temp file in the same directory.
    dir := filepath.Dir(tm.filename)
    tmp, err := os.CreateTemp(dir, "tasks-*.tmp")
    if err != nil {
        return err
    }
    // Ensure we clean up on error
    tmpName := tmp.Name()
    defer func() {
        tmp.Close()
        os.Remove(tmpName)
    }()

    // 2) Write CSV
    writer := csv.NewWriter(tmp)
    for _, rec := range tasks {
        if err := writer.Write(rec); err != nil {
            return err
        }
    }
    writer.Flush()
    if err := writer.Error(); err != nil {
        return err
    }
    // 3) Sync to disk
    if err := tmp.Sync(); err != nil {
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    // 4) Atomically replace the old file
    return os.Rename(tmpName, tm.filename)
}

func (tm *TaskManager) Shutdown() {
	close(tm.shutdownCh)
}

//=====================================================================
// Middleware

type MyHandler struct {
	slog.Handler
}

func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceCtxKey).(string); ok {
		r.Add("trace_id", slog.StringValue(traceID))
	}
	return h.Handler.Handle(ctx, r)
}

func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trace ID & Logger
		id := uuid.New()
		var handler slog.Handler
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
            AddSource: true,
        })
        handler = &MyHandler{handler}
		slog.SetDefault(slog.New(handler))

		ctx := context.WithValue(r.Context(), traceCtxKey, id.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

//=====================================================================
// Handlers

func indexHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := tm.GetTasks()
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to load tasks", ErrAttr(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if err := tmpl.Execute(w, tasks); err != nil {
        log.Printf("template execute error: %v\n", err)
        http.Error(w, "Internal Server Error", 500)
        return
        }
	}
}

func jsonHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, _ := tm.GetTasks()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

func taskHandler(tm *TaskManager, broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With()
		ctx := r.Context()

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode request
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			logger.ErrorContext(ctx, "Invalid request body", ErrAttr(err))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		logger.InfoContext(ctx, "Processing task update", slog.String("name", t.Name), slog.String("status", t.Status))

		// Get and update tasks
		if err := tm.ChangeTask(t.Name, t.Status); err != nil {
			logger.ErrorContext(ctx, "Failed to change task", ErrAttr(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}

		// Notify clients
		broker.messages <- "updated"
		w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        updatedList, _ := tm.GetTasks()
        json.NewEncoder(w).Encode(updatedList)
	}
}

//=====================================================================
// SSE Broker

type Broker struct {
    // New client connection requests
    newClients chan chan string
    // Closed client connection notifications
    defunct    chan chan string
    // Messages to broadcast
    messages   chan string
    // Active clients
    clients    map[chan string]bool
}

func NewBroker() *Broker {
    b := &Broker{
        newClients: make(chan chan string),
        defunct:    make(chan chan string),
        messages:   make(chan string),
        clients:    make(map[chan string]bool),
    }
    go b.listen()
    return b
}

func (b *Broker) listen() {
    for {
        select {
        case c := <-b.newClients:
            b.clients[c] = true
        case c := <-b.defunct:
            delete(b.clients, c)
            close(c)
        case msg := <-b.messages:
            for c := range b.clients {
                c <- msg
            }
        }
    }
}

// SSE handler: streams events to each connected client.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    // Headers for SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Create a message channel for this client
    msgChan := make(chan string)
    b.newClients <- msgChan

    // When handler exits, notify broker to remove this client
    defer func() {
        b.defunct <- msgChan
    }()

    // Listen and serve messages
    for {
        msg, open := <-msgChan
        if !open {
            return
        }
        fmt.Fprintf(w, "data: %s\n\n", msg)
        flusher.Flush()
    }
}


//=====================================================================
// Main

func main() {
    broker := NewBroker()
    tm := NewTaskManager("task_list.csv")

    mux := http.NewServeMux()
    mux.Handle("/", contextMiddleware(indexHandler(tm)))
    mux.Handle("/todos.json", contextMiddleware(jsonHandler(tm)))
    mux.Handle("/events", broker)
    // ToDo: Fix broker being OO'ish
    mux.Handle("/todos", contextMiddleware(taskHandler(tm, broker)))

    server := &http.Server{Addr: ":8080", Handler: mux}

    // Start server in goroutine
    go func() {
        log.Print("Listening on :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Graceful shutdown goroutine
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        log.Print("Exited process gracefully")
		os.Exit(0)
    }()

    // Block main goroutine
    select {}
}
