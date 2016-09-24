package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
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
			name:              name,
			youtubebroadcast:  make(chan []byte),
			audiobroadcast:    make(chan []byte),
			youtubehoses:      make(map[*Hose]bool),
			youtuberegister:   make(chan *Hose),
			youtubeunregister: make(chan *Hose),
			audiohoses:        make(map[*Hose]bool),
			audioregister:     make(chan *Hose),
			audiounregister:   make(chan *Hose),
		}
		log.Println("Headquarters adding room: ", room)
		hq.rooms[name] = room
		go room.Run()
	}
	return room
}

var socket_path = "socket"
var audio_path = "audio"

func main() {
	http.HandleFunc("/"+socket_path+"/", socketHandlerFunc)
	http.HandleFunc("/"+audio_path+"/", audioHandlerFunc)
	http.Handle("/static/", http.FileServer(http.Dir("")))
	http.HandleFunc("/", roomHandler)
	log.Fatal(http.ListenAndServe(":4000", nil))
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

func audioHandlerFunc(w http.ResponseWriter, r *http.Request) {
	audio_room_name := r.URL.Path[len(audio_path)+2:]
	log.Println(audio_room_name)
	websocket.Handler(GetAudioRoomHandler(audio_room_name)).ServeHTTP(w, r)
}

var id = 0

func GetSocketRoomHandler(room_name string) func(c *websocket.Conn) {
	room := headquarters.GetRoom(room_name)
	return func(c *websocket.Conn) {
		hose := &Hose{
			name:          fmt.Sprintf("hose%d", id),
			client:        c,
			send:          make(chan interface{}, 256),
			roombroadcast: room.youtubebroadcast,
			closed:        false,
		}
		id++
		log.Println("About to register youtube", hose, " to ", room)
		room.youtuberegister <- hose
		defer func() { room.youtubeunregister <- hose }()
		go hose.PourDownStream()
		// go hose.testBroadcast()

		log.Println("youtube ", hose, " is drinking")
		hose.DrinkLoop()
	}
}

func GetAudioRoomHandler(room_name string) func(c *websocket.Conn) {
	room := headquarters.GetRoom(room_name)
	return func(c *websocket.Conn) {
		hose := &Hose{
			name:          fmt.Sprintf("hose%d", id),
			client:        c,
			send:          make(chan interface{}, 256),
			roombroadcast: room.audiobroadcast,
			closed:        false,
		}
		id++
		log.Println("About to register audio", hose, " to ", room)
		room.audioregister <- hose
		defer func() { room.audiounregister <- hose }()
		go hose.PourDownStream()
		// go hose.testBroadcast()

		log.Println("audio ", hose, " is drinking")
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
