package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
)

// helper to read the file
func readDisk(name string) ([][]string, error) {
    f, err := os.Open(name)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    return csv.NewReader(f).ReadAll()
}

func TestChangeTask_Concurrent(t *testing.T) {
    const fname = "test_tasks.csv"
    os.Remove(fname)
    tm := NewTaskManager(fname)

    var wg sync.WaitGroup
    writers := 5
    writesEach := 10

    // fire off concurrent atomic ChangeTask calls
    for i := 0; i < writers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < writesEach; j++ {
                name := fmt.Sprintf("g%02d-%02d", id, j)
                if err := tm.ChangeTask(name, "new"); err != nil {
                    t.Errorf("ChangeTask failed: %v", err)
                }
            }
        }(i)
    }

    wg.Wait()

    // verify on‐disk count
    recs, err := readDisk(fname)
    if err != nil {
        t.Fatalf("reading disk: %v", err)
    }
    want := writers * writesEach
    if got := len(recs); got != want {
        t.Fatalf("record count = %d; want %d", got, want)
    }
}

func TestGetTasks_Concurrent(t *testing.T) {
    const fname = "test_tasks2.csv"
    // prepare a small file
    f, _ := os.Create(fname)
    csv.NewWriter(f).WriteAll([][]string{
        {"foo", "1"}, {"bar", "2"}, {"baz", "3"},
    })
    f.Close()

    tm := NewTaskManager(fname)
    var wg sync.WaitGroup

    readers := 10
    for i := 0; i < readers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            tasks, err := tm.GetTasks()
            if err != nil {
                t.Errorf("GetTasks error: %v", err)
                return
            }
            if len(tasks) != 3 {
                t.Errorf("GetTasks saw %d; want 3", len(tasks))
            }
        }()
    }
    wg.Wait()
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
