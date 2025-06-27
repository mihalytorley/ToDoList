package main

import (
	"bytes"
	"context"
	"encoding/csv"
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

func main() {
    // ctrl + C graceful shutdown setup
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Exitted process gracefully")
        os.Exit(1)
    }()

    // Server stuff
    mux := http.NewServeMux()
    
    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/todos.json", jsonHandler)

    th := http.HandlerFunc(taskHandler)
    mux.Handle("/todos", contextMiddleware(th))

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

/*func (c *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return c.clone()
}

func (c *MyHandler) clone() *MyHandler {
    clone := *c
    return &clone
}*/

