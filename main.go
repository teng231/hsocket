package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8000", "http service address")

func main() {
	flag.Parse()
	ws := initWs()
	go ws.start()
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
