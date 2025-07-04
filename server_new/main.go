// main.go
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
)

//=====================================================================
// Logic

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

//=====================================================================
// Main

func main() {
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("Gracefully exitted")
        os.Exit(0)
    }()
    
    broker := NewBroker()
    actor := NewActor()
    

    mux := http.NewServeMux()

    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/todos.json", jsonHandler)
    mux.Handle("/events", broker)
    
    th := http.HandlerFunc(taskHandler(*actor))
    mux.Handle("/todos", th)
    go actor.Receive(*broker)
    

	log.Print("Listening...")
	http.ListenAndServe(":8080", mux)
    // Block main goroutine
    select {}
}
