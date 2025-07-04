package main

import (
	"fmt"
	"net/http"
)

type Broker struct {
    // New client connection requests
    newClients chan chan string
    // Closed client connection notifications
    defunct    chan chan string
    // Messages to broadcast
    messages   chan string
    // Active clients
    clients    map[chan string]bool
}

func NewBroker() *Broker {
    b := &Broker{
        newClients: make(chan chan string),
        defunct:    make(chan chan string),
        messages:   make(chan string),
        clients:    make(map[chan string]bool),
    }
    go b.listen()
    return b
}

func (b *Broker) listen() {
    for {
        select {
        case c := <-b.newClients:
            b.clients[c] = true
        case c := <-b.defunct:
            delete(b.clients, c)
            close(c)
        case msg := <-b.messages:
            for c := range b.clients {
                c <- msg
            }
        }
    }
}

// SSE handler: streams events to each connected client.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    // Headers for SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Create a message channel for this client
    msgChan := make(chan string)
    b.newClients <- msgChan

    // When handler exits, notify broker to remove this client
    defer func() {
        b.defunct <- msgChan
    }()

    // Listen and serve messages
    for {
        msg, open := <-msgChan
        if !open {
            return
        }
        fmt.Fprintf(w, "data: %s\n\n", msg)
        flusher.Flush()
    }
}
