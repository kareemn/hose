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

	// All connected youtube hoses for this room
	youtubehoses map[*Hose]bool

	// All connected audio hoses for this room
	audiohoses map[*Hose]bool

	// Send messages to this channel to braodcast to all youtube hoses.
	youtubebroadcast chan interface{}

	// Send messages to this channel to braodcast to all audio hoses.
	audiobroadcast chan interface{}

	// Add a hose to the hoses pool.
	youtuberegister chan *Hose

	// Remove a hose from the hoses pool
	youtubeunregister chan *Hose

	queue Queue

	// Add a hose to the hoses pool.
	audioregister chan *Hose

	// Remove a hose from the hoses pool
	audiounregister chan *Hose
}


func (room *Room) Run() {
	room.queue.AddItem(PlayableItem{"JXoAmDDPZz4", time.Now().Unix()})
	log.Println("Room is running")
	for {
		select {
		case hose := <-room.audioregister:
			log.Println(hose, " audio registering for ", room)
			room.audiohoses[hose] = true
		case hose := <-room.audiounregister:
			log.Println(hose, " audio unregistering for ", room)
			if room.audiohoses[hose] {
				delete(room.audiohoses, hose)
				hose.Close()
			}
		case pcm_broadcast := <-room.audiobroadcast:
			log.Println("received pcm_broadcast", len(pcm_broadcast.([]byte)) )
			for hose := range room.audiohoses {
			//	go func() {
				hose.send <- pcm_broadcast
			//	}
			}
		case hose := <-room.youtuberegister:
			log.Println(hose, " youtube registering for ", room)
			room.youtubehoses[hose] = true
			go func() {
				time.Sleep(1 * time.Second)
				b, _ := json.Marshal(room.queue.GetPlayingItem())
				if b != nil {
					log.Println("sent ", string(b), " to hose ", hose)
					hose.send <- string(b)
				}
			}()
		case hose := <-room.youtubeunregister:
			log.Println(hose, " youtube unregistering for ", room)
			if room.youtubehoses[hose] {
				delete(room.youtubehoses, hose)
				hose.Close()
			}
		case broadcast_message := <-room.youtubebroadcast:
			var p PlayableItem
			if err := json.Unmarshal([]byte(broadcast_message.(string)), &p); err == nil {
				p.Start = time.Now().Unix()
				room.queue.AddItem(p)
				log.Println(room.queue.String())
			} else {
				log.Println(err)
			}
			b, _ := json.Marshal(p)
			if b != nil {
				for hose := range room.youtubehoses {
						select {
						case hose.send <- string(b):
							// do something
							log.Println("Sent broadcast message", broadcast_message, " to hose: ", hose)
						default:
							log.Println("This hose hasn't picked up messages from it's buffer")
							delete(room.youtubehoses, hose)
							hose.Close()
						}
				}
			}
		}
	}
}

func (room *Room) HosesString() string {
	var buffer bytes.Buffer
	buffer.WriteString("Hoses<\n")
	for hose := range room.youtubehoses {
		buffer.WriteString("youtube " + hose.String())
		buffer.WriteString("\n")
	}
	for hose := range room.audiohoses {
		buffer.WriteString("audio " + hose.String())
		buffer.WriteString("\n")
	}
	buffer.WriteString(">")
	return buffer.String()
}

func (room *Room) String() string {
	return fmt.Sprintf("Room %s<%s>", room.name, room.HosesString())
}

func (room *Room) Close() {
	for hose := range room.youtubehoses {
		hose.Close()
	}
	for hose := range room.audiohoses {
		hose.Close()
	}
}