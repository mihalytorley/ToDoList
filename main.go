package main

import(
"fmt"
"flag"
)



func main() {
    fmt.Println("!... Hello World ...!")
    myslice := []string{}
    var input string = "start"
    for input != "exit" {
        fmt.Scan(&input)
        myslice = append(myslice, input)
    }
    fmt.Println("myslice has value ", myslice)
}

var flagvar int

func flag_test() {
    flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")
}
