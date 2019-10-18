package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8000", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func main() {
	flag.Parse()
	ws := initWs()
	go ws.start()
	http.HandleFunc("/", serveHome)

	http.HandleFunc("/ws-firer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		msg := &Message{}
		if err := json.Unmarshal(b, msg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		ws.broadcast <- msg
		w.Write([]byte("done"))
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client := initClient(ws, w, r)
		go client.inPump()
		go client.outPump()
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
