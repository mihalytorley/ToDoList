package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"log/slog"
// 	"net/http"
// 	"os"

// 	"github.com/google/uuid"
// )

// func contextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
//         // Logger and context setup
//         id := uuid.New()
//         var handler slog.Handler
//         handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
//             AddSource: true,
//         })
//         handler = &MyHandler{handler}
//         slog.SetDefault(slog.New(handler))
//         ctx := context.Background()
//         ctx = context.WithValue(ctx, traceCtxKey, id.String())
//         logger := slog.With()

//         logger.InfoContext(ctx, "Logging request")

// 		next.ServeHTTP(w, r)
// 	})
// }

// func taskHandler(w http.ResponseWriter, r *http.Request) {
//     // Logger and context setup
//     id := uuid.New()
//     var handler slog.Handler
//     handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
//         AddSource: true,
//     })
//     handler = &MyHandler{handler}
//     slog.SetDefault(slog.New(handler))
//     ctx := context.Background()
//     ctx = context.WithValue(ctx, traceCtxKey, id.String())

//     logger := slog.With()

//     filename := "task_list.csv"

//     //post Method
//     if r.Method == http.MethodPost {
//         // Catching request
//         task := &Task{}
//         err := json.NewDecoder(r.Body).Decode(task)
//         if err != nil {
//             http.Error(w, err.Error(), http.StatusBadRequest)
//             return
//         }

//         fmt.Println("got task:", task)

//         // reading CSV file
//         logger.Debug("Reading CSV file")
//         to_do_list, err := readCSVFile(filename)
//         if err!= nil {
//             logger.ErrorContext(ctx, "Error reading file")
//             return
//         }

//         // Check if change is an addition, subtraction, or change in task status
//         name := task.Name
//         status := task.Status
//         fmt.Print(name, ",", status)
//         // task_split := strings.Split(task, ",")
//         // name := task_split[0]
//         // status := task_split[1]
//         to_do_list = changeCheck(to_do_list, name, status)
        
//         /*myslice := []string{}
//         var input string = "start"
//         for input != "exit" {
//             fmt.Scan(&input)
//             myslice = append(myslice, input)
//         }
//         fmt.Println("myslice has value ", myslice) */

//         // Writing CSV file starts here
//         logger.Debug("Writing to CSV file")
//         writer, file, err := createCSVWriter(filename)
//         if err != nil {
//             logger.ErrorContext(ctx, "Error creating CSV writer")
//             return
//         }
//         defer file.Close()
//         for _, record := range to_do_list {
//             err = writeCSVRecord(writer, record)
//             if err := writer.Error(); err != nil {
//                 logger.ErrorContext(ctx, "Error writing to CSV")
//             }
//         }
//         // Flush the writer and check for any errors
//         writer.Flush()
//         if err := writer.Error(); err != nil {
//             logger.ErrorContext(ctx, "Error flushing CSV writer")
//         }
//         logger.InfoContext(ctx, "Task change recorded")
//         fmt.Println("\n", to_do_list)
//     }

//     //Get method
//     if r.Method == http.MethodGet {
//         logger.Debug("Reading CSV file")
//         to_do_list, err := readCSVFile(filename)
//         if err!= nil {
//             logger.ErrorContext(ctx, "Error reading file")
//             return
//         }
//         logger.InfoContext(ctx, "Returning To Do List")
//         fmt.Println("\n", to_do_list)
//     }

// 	w.WriteHeader(http.StatusCreated)
// }

// func indexHandler(w http.ResponseWriter, r *http.Request) {
//     // we don’t need to load CSV here if JS will fetch JSON,
//     // but we can pass an initial render too:
//     filename := "task_list.csv"
//     todos, err := readCSVFile(filename)
//     if err != nil {
//         http.Error(w, err.Error(), 500)
//         return
//     }
//     if err := tmpl.Execute(w, todos); err != nil {
//         log.Printf("template execute error: %v\n", err)
//         http.Error(w, "Internal Server Error", 500)
//         return
//     }
// }

// func jsonHandler(w http.ResponseWriter, r *http.Request) {
//     filename := "task_list.csv"
//     todos, err := readCSVFile(filename)
//     if err != nil {
//         http.Error(w, err.Error(), 500)
//         return
//     }
//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(todos)
// }
