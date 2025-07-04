package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestRaceConditions(t *testing.T) {
	test_to_do_list_initial := []Task{{Name: "task_1", Status: "not_started"},{Name: "task_2", Status: "not_started"},{Name: "task_3", Status: "not_started"},{Name: "task_4", Status: "not_started"},{Name: "task_5", Status: "not_started"},{Name: "task_6", Status: "not_started"},{Name: "task_7", Status: "not_started"},{Name: "task_8", Status: "not_started"}}
	for _, task := range test_to_do_list_initial {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
	test_to_do_list_second := []Task{{Name: "task_11", Status: "not_started"},{Name: "task_12", Status: "not_started"},{Name: "task_13", Status: "not_started"},{Name: "task_14", Status: "not_started"},{Name: "task_15", Status: "not_started"},{Name: "task_16", Status: "not_started"},{Name: "task_17", Status: "not_started"},{Name: "task_18", Status: "not_started"}}
	for _, task := range test_to_do_list_second {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
}
