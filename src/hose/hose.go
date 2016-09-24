package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
)

type Hose struct {
	name   string
	closed bool
	client *websocket.Conn
	// Send messages to this channel to send them along to the websocket.
	send chan interface{}

	// Send messages to this channel to broadcast to entire room.
	roombroadcast chan []byte
}

func (hose *Hose) Close() {
	log.Println("Closing ", hose)
	close(hose.send)
	hose.client.Close()
	hose.closed = true
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

// Receive messages from websocket client.
func (hose *Hose) DrinkLoop() {
	if hose.roombroadcast == nil {
		log.Fatal("No room broadcast set for drink loop")
		return
	}

	for {
		if hose.closed {
			break
		}
		var message []byte
		err := websocket.Message.Receive(hose.client, &message)
		if err != nil {
			log.Println("Error ", err, "for ", hose, "Stop drinking.")
			break
		}
		log.Println(message)
		hose.roombroadcast <- message
	}
}
