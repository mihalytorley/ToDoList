package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestRaceConditions(t *testing.T) {
	test_to_do_list_initial := []Task{{Name: "task_21", Status: "hello"},{Name: "task_22", Status: "hello"},{Name: "task_23", Status: "hello"},{Name: "task_24", Status: "hello"},{Name: "task_25", Status: "hello"},{Name: "task_26", Status: "hello"},{Name: "task_27", Status: "hello"},{Name: "task_28", Status: "hello"}}
	for _, task := range test_to_do_list_initial {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
	test_to_do_list_second := []Task{{Name: "task_211", Status: "hello"},{Name: "task_212", Status: "hello"},{Name: "task_213", Status: "hello"},{Name: "task_214", Status: "hello"},{Name: "task_215", Status: "hello"},{Name: "task_216", Status: "hello"},{Name: "task_217", Status: "hello"},{Name: "task_218", Status: "hello"}}
	for _, task := range test_to_do_list_second {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(task)
		fmt.Print(err)
		resp, err := http.Post("http://localhost:8080/todos", "application/json", b)
		fmt.Print(err)
		defer resp.Body.Close()
	}
}
