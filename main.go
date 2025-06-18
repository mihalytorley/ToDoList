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
    "os/exec"
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

func (c *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return c.clone()
}

func (c *MyHandler) clone() *MyHandler {
    clone := *c
    return &clone
}

func main() {
    // Logger and context setup
    newUUID, err := exec.Command("uuidgen").Output()
    var handler slog.Handler
    handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
        AddSource: true,
    })
    handler = &MyHandler{handler}
    slog.SetDefault(slog.New(handler))
    ctx := context.Background()
    ctx = context.WithValue(ctx, traceCtxKey, newUUID)

    logger := slog.With("component", "test")

    //Reading CSV file starts here
    filename := "task_list.csv"
    flag.Parse()
    
    logger.Debug("Reading CSV file")
    data, err := readCSVFile(filename)
    if err!= nil {
        logger.ErrorContext(ctx, "Error reading file:", err)
        return
    }
    reader, err := parseCSV(data)
    if err!= nil {
        logger.ErrorContext(ctx, "Error creating CSV reader:", err)
        return
    }
    to_do_list, err := processCSV(reader)
    if err!= nil {
        logger.ErrorContext(ctx, "Error processing the CSV:", err)
        return
    }

    // Check if change is an addition, subtraction, or change in task status
    i := 0
    task_change_bool := false
    for _, task := range to_do_list {
        if slices.Contains(task, *name) {
            to_do_list[i][1] = *status
            task_change_bool = true
            if *status == "delete" {
                if len(to_do_list)-1 == i {
                    to_do_list = to_do_list[:i]
                } else {
                    to_do_list = slices.Delete(to_do_list, i, i+1)
                }
            }
        }
        i++
    }
    if task_change_bool == false {
        new_task := []string{*name, *status}
        to_do_list = append(to_do_list, new_task)
    }
    
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
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Exitted  before process finished")
        os.Exit(1)
    }()
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

func readCSVFile(filename string) ([]byte, error) {
    f, err := os.Open(filename)
    if err!= nil {
        return nil, err
    }
    defer f.Close()
    data, err := io.ReadAll(f)
    if err!= nil {
        return nil, err
    }
    return data, nil
}

func parseCSV(data []byte) (*csv.Reader, error) {
    reader := csv.NewReader(bytes.NewReader(data))
    return reader, nil
}

func processCSV(reader *csv.Reader) ([][]string, error){
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
