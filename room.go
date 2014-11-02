package main

import (
	"bytes"
	"fmt"
	"log"
	"time"
	"encoding/json"
)

type Room struct {
	name string

	// All connected hoses for this room
	hoses map[*Hose]bool

	// Send messages to this channel to braodcast to all hoses.
	broadcast chan string

	// Add a hose to the hoses pool.
	register chan *Hose

	// Remove a hose from the hoses pool
	unregister chan *Hose

	queue Queue
}

func (room *Room) Run() {
	room.queue.AddItem(PlayableItem{"JXoAmDDPZz4", time.Now().Unix()})
	log.Println("Room is running")
	for {
		select {
		case hose := <-room.register:
			log.Println(hose, " registering for ", room)
			room.hoses[hose] = true
			go func() {
				time.Sleep(5 * time.Second)
				b, _ := json.Marshal(room.queue.GetPlayingItem())
				if b != nil {
					log.Println("sent ", string(b), " to hose ", hose)
					hose.send <- string(b)
				}
			}()
		case hose := <-room.unregister:
			log.Println(hose, " unregistering for ", room)
			if room.hoses[hose] {
				delete(room.hoses, hose)
				hose.Close()
			}
		case broadcast_message := <-room.broadcast:
			var p PlayableItem
			if err := json.Unmarshal([]byte(broadcast_message), &p); err == nil {
				p.Start = time.Now().Unix()
				room.queue.AddItem(p)
				log.Println(room.queue.String())
			} else {
				log.Println(err)
			}
			for hose := range room.hoses {
				select {
				case hose.send <- broadcast_message:
					// do something
					log.Println("Sent broadcast message", broadcast_message, " to hose: ", hose)
				default:
					log.Println("This hose hasn't picked up messages from it's buffer")
					delete(room.hoses, hose)
					hose.Close()
				}
			}
		}
	}
}

func (room *Room) HosesString() string {
	var buffer bytes.Buffer
	buffer.WriteString("Hoses<\n")
	for hose := range room.hoses {
		buffer.WriteString(hose.String())
		buffer.WriteString("\n")
	}
	buffer.WriteString(">")
	return buffer.String()
}

func (room *Room) String() string {
	return fmt.Sprintf("Room %s<%s>", room.name, room.HosesString())
}

func (room *Room) Close() {
	for hose := range room.hoses {
		hose.Close()
	}
}