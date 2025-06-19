package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"
    "github.com/google/uuid"
)


type task struct {
    name string
    status string
}

var (
    name *string
    status *string
)

func init() {
    name = flag.String("name", "", "a string describing the name of the task")
    status = flag.String("status", "", "a string describing if the task is either not started, started, or completed")
}

type contextKey int

const (
    traceCtxKey contextKey = iota + 1
)

type MyHandler struct {
    slog.Handler
}

func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
    if traceID, ok := ctx.Value(traceCtxKey).(string); ok {
        r.Add("trace_id", slog.StringValue(traceID))
    }
    return h.Handler.Handle(ctx, r)
}

/*func (c *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return c.clone()
}

func (c *MyHandler) clone() *MyHandler {
    clone := *c
    return &clone
}*/

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
    //slog.InfoContext(ctx, "message")

    logger := slog.With()

    //Reading CSV file starts here
    filename := "task_list.csv"
    flag.Parse()
    
    logger.Debug("Reading CSV file")
    to_do_list, err := readCSVFile(filename)
    if err!= nil {
        logger.ErrorContext(ctx, "Error reading file:", err)
        return
    }

    // Check if change is an addition, subtraction, or change in task status
    to_do_list = changeCheck(to_do_list)
    
    /*myslice := []string{}
    var input string = "start"
    for input != "exit" {
        fmt.Scan(&input)
        myslice = append(myslice, input)
    }
    fmt.Println("myslice has value ", myslice) */

    // Writing CSV file starts here
    logger.Debug("Writing to CSV file")
    writer, file, err := createCSVWriter(filename)
    if err != nil {
        logger.ErrorContext(ctx, "Error creating CSV writer:", err)
        return
    }
    defer file.Close()
    for _, record := range to_do_list {
        err = writeCSVRecord(writer, record)
        if err := writer.Error(); err != nil {
            logger.ErrorContext(ctx, "Error writing to CSV:", err)
        }
    }
    // Flush the writer and check for any errors
    writer.Flush()
    if err := writer.Error(); err != nil {
        logger.ErrorContext(ctx, "Error flushing CSV writer:", err)
    }
    logger.InfoContext(ctx, "Task change recorded")
    fmt.Println(to_do_list)
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

func changeCheck(list_of_lists [][]string) ([][]string) {
    i := 0
    task_change_bool := false
    for _, task := range list_of_lists {
        if slices.Contains(task, *name) {
            list_of_lists[i][1] = *status
            task_change_bool = true
            if *status == "delete" {
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
        new_task := []string{*name, *status}
        list_of_lists = append(list_of_lists, new_task)
    }
    return list_of_lists
}

