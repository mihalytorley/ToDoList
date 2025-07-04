package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestRaceConditions(t *testing.T) {
	test_to_do_list_initial := []Task{{Name: "task_21", Status: "hello"},{Name: "task_22", Status: "hello"},{Name: "task_23", Status: "hello"},{Name: "task_24", Status: "hello"},{Name: "task_25", Status: "hello"},{Name: "task_26", Status: "hello"},{Name: "task_27", Status: "hello"},{Name: "task_28", Status: "hello"},{Name: "task_31", Status: "hello"},{Name: "task_32", Status: "hello"},{Name: "task_33", Status: "hello"},{Name: "task_34", Status: "hello"},{Name: "task_35", Status: "hello"},{Name: "task_36", Status: "hello"},{Name: "task_37", Status: "hello"},{Name: "task_38", Status: "hello"},{Name: "task_41", Status: "hello"},{Name: "task_42", Status: "hello"},{Name: "task_43", Status: "hello"},{Name: "task_44", Status: "hello"},{Name: "task_45", Status: "hello"},{Name: "task_46", Status: "hello"},{Name: "task_47", Status: "hello"},{Name: "task_48", Status: "hello"},{Name: "task_51", Status: "hello"},{Name: "task_52", Status: "hello"},{Name: "task_53", Status: "hello"},{Name: "task_54", Status: "hello"},{Name: "task_55", Status: "hello"},{Name: "task_56", Status: "hello"},{Name: "task_57", Status: "hello"},{Name: "task_58", Status: "hello"},{Name: "task_61", Status: "hello"},{Name: "task_62", Status: "hello"},{Name: "task_63", Status: "hello"},{Name: "task_64", Status: "hello"},{Name: "task_65", Status: "hello"},{Name: "task_66", Status: "hello"},{Name: "task_67", Status: "hello"},{Name: "task_68", Status: "hello"}}
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
