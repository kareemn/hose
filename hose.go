package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"log"
)

type Hose struct {
	name   string
	closed bool
	client *websocket.Conn
	// Send messages to this channel to send them along to the websocket.
	send chan string

	room *Room
}

func (hose *Hose) Close() {
	log.Println("Closing ", hose)
	close(hose.send)
	hose.client.Close()
	hose.closed = true
}

// Receive messages from websocket client.
func (hose *Hose) DrinkLoop() {
	for {
		if hose.closed {
			break
		}
		var message string
		err := websocket.Message.Receive(hose.client, &message)
		if err != nil {
			log.Println("Error ", err, "for ", hose, "Stop drinking.")
			break
		}
		log.Println("received message from ", hose, " ", message)
		hose.room.broadcast <- message
	}
}

// Send messages to websocket client.
func (hose *Hose) PourDownStream() {
	for message := range hose.send {
		if hose.closed {
			break
		}
		log.Println("Pouring in ", hose)
		err := websocket.Message.Send(hose.client, message)
		if err != nil {
			log.Println("Error %s for hose: %s. Stop pouring.", err, hose)
			break
		}
	}
}

func (hose *Hose) String() string {
	return fmt.Sprintf("%s", hose.name)
}