package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/my0sot1s/header/wsh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var addr = flag.String("addr", ":"+os.Getenv("PORT"), "http service address")

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

func wsFirer(ws *Ws, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	msg := &wsh.Message{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ws.broadcast <- msg

	w.Write([]byte("done"))
}

func ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	w.Write([]byte("pong"))
}

func wsInspect(ws *Ws, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	detail := make(map[string]int)
	for k, val := range ws.mapTopics {
		detail[k] = len(val)
	}
	if b, err := json.Marshal(detail); err == nil {
		w.Write([]byte(b))
		return
	}
	w.Write([]byte("troube for inspect"))
}

func startGrpcServer(port string, handle *wsProvider) error {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	serve := grpc.NewServer()
	wsh.RegisterWsProviderServer(serve, handle)
	reflection.Register(serve)
	return serve.Serve(listen)
}

func main() {
	flag.Parse()
	ws := initWs()
	go ws.start()
	if os.Getenv("GRPC_PORT") != "" {
		wsHandle := &wsProvider{ws}
		go startGrpcServer(":"+os.Getenv("GRPC_PORT"), wsHandle)
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/demo", serveHome)
	http.HandleFunc("/wsClient.js", serveJS)
	http.HandleFunc("/ping", ping)

	http.HandleFunc("/ws-firer", func(w http.ResponseWriter, r *http.Request) {
		wsFirer(ws, w, r)
	})

	http.HandleFunc("/ws-inspect", func(w http.ResponseWriter, r *http.Request) {
		wsInspect(ws, w, r)
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
	log.Print("WS run away at ", os.Getenv("PORT"))
}
