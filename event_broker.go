package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

// EventBroker keeps track of debugging clients
type EventBroker struct {
	Notifier       chan []byte
	newClients     chan chan []byte
	closingClients chan chan []byte
	clients        map[chan []byte]bool
}

// Initialize all external dependencies for EventBroker
func (e *EventBroker) Initialize() {
	e.Notifier = make(chan []byte, 1)
	e.newClients = make(chan chan []byte)
	e.closingClients = make(chan chan []byte)
	e.clients = make(map[chan []byte]bool)
}

func (e *EventBroker) brokerage() {
	for {
		select {
		case s := <-e.newClients:
			e.clients[s] = true
			log.Printf("Client added. %d registered clients", len(e.clients))
		case s := <-e.closingClients:
			delete(e.clients, s)
			log.Printf("Removed client. %d registered clients", len(e.clients))
		case event := <-e.Notifier:
			for clientMessageChan := range e.clients {
				clientMessageChan <- event
			}
		}
	}
}

// Run starts the HTTP SSE server
func (e *EventBroker) Run(addr string) {

	go e.brokerage()

	loggedRouter := handlers.LoggingHandler(os.Stdout, e)
	log.Fatal(http.ListenAndServe(addr, loggedRouter))
}

// ServeHTTP implements the http.Handler
func (e *EventBroker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte)

	e.newClients <- messageChan
	defer func() {
		e.closingClients <- messageChan
	}()
	notify := rw.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		e.closingClients <- messageChan
	}()

	for {
		fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)
		flusher.Flush()
	}

}
