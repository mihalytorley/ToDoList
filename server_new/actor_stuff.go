package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

type Actor struct {
    inbox chan Task
}

func NewActor() *Actor {
    return &Actor{
        inbox: make(chan Task, 5),
    }
}


func (a *Actor) Receive(broker Broker) {
    start_outside := time.Now()
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
    for task := range a.inbox {
        start_inside := time.Now()
        // reading CSV file
        logger.Debug("Reading CSV file")
        to_do_list, err := readCSVFile(filename)
        if err!= nil {
            logger.ErrorContext(ctx, "Error reading file")
            return
        }

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
		broker.messages <- "updated"
        elapsed_inside := time.Since(start_inside)
        fmt.Println("\nTime elapsed inside: ", elapsed_inside)
    }
    elapsed_outside := time.Since(start_outside)
    fmt.Println("\nTime elapsed outside: ", elapsed_outside)
}

func (a *Actor) Send(task Task) {
    a.inbox <- task
}