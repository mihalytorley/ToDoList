// e2e_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

// helper to decode JSON response
func fetchTasks(ts *httptest.Server, t *testing.T) [][]string {
    resp, err := http.Get(ts.URL + "/todos.json")
    if err != nil {
        t.Fatalf("GET /todos.json failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("GET /todos.json returned %d: %s", resp.StatusCode, string(body))
    }
    var tasks [][]string
    if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
        t.Fatalf("Invalid JSON: %v", err)
    }
    return tasks
}

func TestE2E_ConcurrentHTTP(t *testing.T) {
    // 1) Prepare a temp CSV file
    tmpFile, err := os.CreateTemp("", "e2e_tasks_*.csv")
    if err != nil {
        t.Fatalf("Creating temp file: %v", err)
    }
    filename := tmpFile.Name()
    tmpFile.Close()
    defer os.Remove(filename)

    // 2) Wire up your server using httptest
    broker := NewBroker()
    tm := NewTaskManager(filename)
    mux := http.NewServeMux()
    mux.Handle("/", contextMiddleware(indexHandler(tm)))
    mux.Handle("/todos.json", contextMiddleware(jsonHandler(tm)))
    mux.Handle("/events", broker)
    mux.Handle("/todos", contextMiddleware(taskHandler(tm, broker)))

    ts := httptest.NewServer(mux)
    defer ts.Close()

    // 3) Concurrently POST new tasks
    writers := 5
    writesEach := 10
    var wg sync.WaitGroup

    for i := 0; i < writers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < writesEach; j++ {
                task := Task{
                    Name:   fmt.Sprintf("g%02d-%02d", id, j),
                    Status: "new",
                }
                body, _ := json.Marshal(task)
                resp, err := http.Post(ts.URL+"/todos", "application/json", bytes.NewReader(body))
                if err != nil {
                    t.Errorf("POST error: %v", err)
                    return
                }
                resp.Body.Close()
                if resp.StatusCode != http.StatusOK {
                    t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
                }
            }
        }(i)
    }
    wg.Wait()

    // 4) Fetch final list and assert count
    tasks := fetchTasks(ts, t)
    want := writers * writesEach
    if got := len(tasks); got != want {
        t.Fatalf("E2E: got %d tasks; want %d", got, want)
    }
}
