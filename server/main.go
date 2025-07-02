package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"text/template"

	"github.com/google/uuid"
)

type Task struct {
    Name string   `json:"name"`
    Status string `json:"status"`
}

type contextKey int

const (
    traceCtxKey contextKey = iota + 1
    // ToDo: Don't use enum
    // ToDo: use magic string instead of iota
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

func main() {
    // ctrl + C graceful shutdown setup
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Exitted process gracefully")
        os.Exit(1)
    }()

    // Server stuff
    broker := NewBroker()

    mux := http.NewServeMux()

    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/todos.json", jsonHandler)

    mux.Handle("/events", broker)

    th := http.HandlerFunc(taskHandlerWithBroker(broker))
    mux.Handle("/todos", contextMiddleware(th, *broker))

	log.Print("Listening...")
	http.ListenAndServe(":8080", mux)
}

func createCSVWriter(filename string) (*csv.Writer, *os.File, error) {
    f, err := os.Create(filename)
    if err != nil {
        return nil, nil, err
    }
    writer := csv.NewWriter(f)
    return writer, f, nil
}

func writeCSVRecord(writer *csv.Writer, record []string) (error){
    err := writer.Write(record)
    if err != nil {
        return err
    }
    return nil
}

func readCSVFile(filename string) ([][]string, error) {
    f, err := os.Open(filename)
    if err!= nil {
        return nil, err
    }
    defer f.Close()
    data, err := io.ReadAll(f)
    if err!= nil {
        return nil, err
    }
    reader := csv.NewReader(bytes.NewReader(data))
    task_list := [][]string{}
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        } else if err!= nil {
            return nil, err
        }
        task_list = append(task_list, record)
    }
    return task_list, nil
}

func changeCheck(list_of_lists [][]string, name string, status string) ([][]string) {
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
    if task_change_bool == false {
        new_task := []string{name, status}
        list_of_lists = append(list_of_lists, new_task)
    }
    return list_of_lists
}

func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
    if traceID, ok := ctx.Value(traceCtxKey).(string); ok {
        r.Add("trace_id", slog.StringValue(traceID))
    }
    return h.Handler.Handle(ctx, r)
}

func ErrAttr(err error) slog.Attr {
    return slog.Any("error", err)
}

//=====================================================================
// Webstuff

func contextMiddleware(next http.Handler, broker Broker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
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

        logger.InfoContext(ctx, "Logging request")

		next.ServeHTTP(w, r)
	})
}

func taskHandlerWithBroker(broker *Broker) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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

        filename := "task_list.csv"
        to_do_list, err := readCSVFile(filename)
        if err!= nil {
            logger.ErrorContext(ctx, "Error reading file")
            return
        }


        //post Method
        if r.Method == http.MethodPost {
            // Catching request
            task := &Task{}
            err := json.NewDecoder(r.Body).Decode(task)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            fmt.Println("got task:", task)

            // Check if change is an addition, subtraction, or change in task status
            name := task.Name
            status := task.Status
            fmt.Print(name, ",", status)
            to_do_list = changeCheck(to_do_list, name, status)
            

            // Writing CSV file starts here
            logger.Debug("Writing to CSV file")
            writer, file, err := createCSVWriter(filename)
            if err != nil {
                logger.ErrorContext(ctx, "Error creating CSV writer")
                return
            }
            defer file.Close()
            for _, record := range to_do_list {
                err = writeCSVRecord(writer, record)
                if err := writer.Error(); err != nil {
                    logger.ErrorContext(ctx, "Error writing to CSV")
                }
            }
            // Flush the writer and check for any errors
            writer.Flush()
            if err := writer.Error(); err != nil {
                logger.ErrorContext(ctx, "Error flushing CSV writer")
            }
            logger.InfoContext(ctx, "Task change recorded")
            fmt.Println("\n", to_do_list)
        }

        //Get method
        if r.Method == http.MethodGet {
            logger.InfoContext(ctx, "Returning To Do List")
            fmt.Println("\n", to_do_list)
        }


        broker.messages <- "updated"
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

/*func (c *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return c.clone()
}

func (c *MyHandler) clone() *MyHandler {
    clone := *c
    return &clone
}*/

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

