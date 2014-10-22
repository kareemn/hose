package main

import (
	"bytes"
	"container/list"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"encoding/json"
)

type RoomView struct {
	Title   string
	Playing string
}

type HeadQuarters struct {
	rooms map[string]*Room
}

var headquarters = HeadQuarters{
	rooms: make(map[string]*Room),
}

func (hq *HeadQuarters) GetRoom(name string) *Room {
	room := hq.rooms[name]
	if room == nil {
		room = &Room{
			name:       name,
			broadcast:  make(chan string),
			register:   make(chan *Hose),
			unregister: make(chan *Hose),
			hoses:      make(map[*Hose]bool),
		}
		log.Println("Headquarters adding room: ", room)
		hq.rooms[name] = room
		go room.Run()
	}
	return room
}

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
	room.queue.AddItem(PlayableItem{"JXoAmDDPZz4"})
	log.Println("Room is running")
	for {
		select {
		case hose := <-room.register:
			log.Println(hose, " registering for ", room)
			room.hoses[hose] = true
			go func() {
				time.Sleep(5 * time.Second)
				b, _ := json.Marshal(room.queue.GetPlayingItem())
				log.Println("sent ", string(b), " to hose ", hose)
				hose.send <- string(b)
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

type PlayableItem struct {
	Id string
}

type Queue struct {
	videos list.List
}

func (q *Queue) GetPlayingItem() PlayableItem {
	return q.videos.Back().Value.(PlayableItem)
}

func (q *Queue) AddItem(item PlayableItem) {
	q.videos.PushBack(item)
}

func (q *Queue) String() string {
	s := ""
	for e := q.videos.Front() ; e != nil ; e = e.Next() {
		var p PlayableItem = e.Value.(PlayableItem)
		s += ";" + p.Id
	}
	return s
}

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

var socket_path = "socket"

func main() {
	http.HandleFunc("/"+socket_path+"/", socketHandlerFunc)
	http.Handle("/static/", http.FileServer(http.Dir("")))
	http.HandleFunc("/", roomHandler)
	http.ListenAndServe("localhost:4000", nil)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Yoooo")
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	room := headquarters.GetRoom("kimo")
	

	roomView := &RoomView{Title: path, Playing: room.queue.String()}
	t, _ := template.ParseFiles("static/room.html")
	t.Execute(w, roomView)
}

func socketHandlerFunc(w http.ResponseWriter, r *http.Request) {
	socket_room_name := r.URL.Path[len(socket_path)+2:]
	log.Println(socket_room_name)
	websocket.Handler(GetSocketRoomHandler(socket_room_name)).ServeHTTP(w, r)
}

var id = 0

func GetSocketRoomHandler(room_name string) func(c *websocket.Conn) {
	room := headquarters.GetRoom(room_name)
	return func(c *websocket.Conn) {
		hose := &Hose{
			name:   fmt.Sprintf("hose%d", id),
			client: c,
			send:   make(chan string, 256),
			room:   room,
			closed: false,
		}
		id++
		log.Println("About to register ", hose, " to ", room)
		room.register <- hose
		defer func() { room.unregister <- hose }()
		log.Println("About to start pouring")
		go hose.PourDownStream()
		// go hose.testBroadcast()

		log.Println(hose, " is drinking")
		hose.DrinkLoop()
	}
}

func (hose *Hose) testBroadcast() {
	time.Sleep(5 * time.Second)
	if !hose.closed {
		hose.send <- "ahwSmcZxBAU"
	}
}

func socketHandler(c *websocket.Conn) {
	var s string
	fmt.Fscan(c, &s)
	fmt.Println("Received:", s)
	fmt.Fprint(c, "How do you do?")
}
