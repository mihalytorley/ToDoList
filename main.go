package main

import(
    "fmt"
    "flag"
    "bytes"
    "encoding/csv"
    "io"
    "os"
)


type task struct {
    name string
    status string
}

var (
    name *string
    status *string
)

func init() {
    name = flag.String("name", "", "a string describing the name of the task")
    status = flag.String("status", "", "a string describing if the task is either not started, started, or completed")
}

func main() {
    to_do_list := []task{}
    fmt.Println("!... Hello World ...!")
    flag.Parse()
    new_task := task{
        name: *name,
        status: *status,
    }
    to_do_list = append(to_do_list, new_task)
    fmt.Println(to_do_list)
    /*myslice := []string{}
    var input string = "start"
    for input != "exit" {
        fmt.Scan(&input)
        myslice = append(myslice, input)
    }
    fmt.Println("myslice has value ", myslice) */
}

