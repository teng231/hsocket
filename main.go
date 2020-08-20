package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

var port = ""

func init() {
	flag.StringVar(&port, "port", ":8000", "http port")

}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("using /demo to run demo"))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func serveJS(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "wsClient.js")
}

func wsFirer(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	event := &Event{}
	if err := json.NewDecoder(r.Body).Decode(event); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// client := hub.getClient(event.ConnId, event.Sender)
	// // when disconnected
	// if event.NotificationType == T_disconnected {
	// 	client.disconnect()
	// 	return
	// }
	// // when subscribe topic
	// if event.NotificationType == T_subscribe {
	// 	client.subscribe(event.SendTo, event.Sender)
	// 	return
	// }
	// // when unsubscribe topic
	// if event.NotificationType == T_unsubscribe {
	// 	client.unsubscribe(event.SendTo, event.Sender)
	// 	return
	// }
	hub.broadcast <- event
	w.Write([]byte("done"))
}

func ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	w.Write([]byte("pong"))
}

func handleListTopics(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	bin, _ := json.Marshal(hub.topics)
	w.Write(bin)
}

func main() {
	flag.Parse()
	hub := initWs()
	go hub.onStart()

	http.HandleFunc("/", index)
	http.HandleFunc("/demo", serveHome)
	http.HandleFunc("/wsClient.js", serveJS)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/topics", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		handleListTopics(hub, w, r)
	})
	http.HandleFunc("/ws-firer", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		wsFirer(hub, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client := makeWsClient(hub, w, r)
		go client.onWsListenMessage()
		go client.onWsPushMessage()
	})

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Print("WS run away at ", port)
}
