package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestRaceConditions(t *testing.T) {
	test_to_do_list_initial := []Task{{Name: "task_1", Status: "hello"},{Name: "task_2", Status: "hello"},{Name: "task_3", Status: "hello"},{Name: "task_4", Status: "hello"},{Name: "task_5", Status: "hello"},{Name: "task_6", Status: "hello"},{Name: "task_7", Status: "hello"},{Name: "task_8", Status: "hello"}}
	for _, task := range test_to_do_list_initial {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
	test_to_do_list_second := []Task{{Name: "task_11", Status: "hello"},{Name: "task_12", Status: "hello"},{Name: "task_13", Status: "hello"},{Name: "task_14", Status: "hello"},{Name: "task_15", Status: "hello"},{Name: "task_16", Status: "hello"},{Name: "task_17", Status: "hello"},{Name: "task_18", Status: "hello"}}
	for _, task := range test_to_do_list_second {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
}
