package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"
)

func TestRaceConditions(t *testing.T) {
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

func TestChangeCheckAddition(t *testing.T) {
    test_to_do_list_initial := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    test_to_do_list_actual := changeCheck(test_to_do_list_initial, "task_5", "not_started")
    test_to_do_list_expected := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"},{"task_5", "not_started"}}
    if reflect.DeepEqual(test_to_do_list_actual, test_to_do_list_expected) != true {
        t.Errorf("Addition to list is wrong")
    }
}

func TestChangeCheckDeletionMiddle(t *testing.T) {
    test_to_do_list_initial := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    test_to_do_list_actual := changeCheck(test_to_do_list_initial, "task_2", "delete")
    test_to_do_list_expected := [][]string{{"task_1", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    if reflect.DeepEqual(test_to_do_list_actual, test_to_do_list_expected) != true {
        t.Errorf("Deletion from middle of list is wrong")
    }
}

func TestChangeCheckDeletionEnd(t *testing.T) {
    test_to_do_list_initial := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    test_to_do_list_actual := changeCheck(test_to_do_list_initial, "task_4", "delete")
    test_to_do_list_expected := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"}}
    if reflect.DeepEqual(test_to_do_list_actual, test_to_do_list_expected) != true {
        t.Errorf("Deletion from end of list is wrong")
    }
}

func TestChangeCheckStatusChange(t *testing.T) {
    test_to_do_list_initial := [][]string{{"task_1", "not_started"},{"task_2", "not_started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    test_to_do_list_actual := changeCheck(test_to_do_list_initial, "task_2", "started")
    test_to_do_list_expected := [][]string{{"task_1", "not_started"},{"task_2", "started"},{"task_3", "not_started"},{"task_4", "not_started"}}
    if reflect.DeepEqual(test_to_do_list_actual, test_to_do_list_expected) != true {
        t.Errorf("Deletion to list is wrong")
    }
}
