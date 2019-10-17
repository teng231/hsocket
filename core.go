package main

import(
	"net/http"
	"time"
	"log"
	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


type Message struct {
	payload []byte
	created int64
	Type string
	eventId string
}

type Command struct {
	client *Client
	Type string
	payload string
}

type Ws struct {
	clients map[*Client] bool
	broadcast chan *Message
	register chan *Command
	mapEvents map[string][]*Client
}

type Client struct {
	id string
	ws *Ws
	conn *websocket.Conn
	sender chan []byte
}

func startServe() {
	ws := initWs()
}


func initWs() *Ws {
	return &Ws{
		broadcast:  make(chan *Message),
		register:   make(chan *Command),
		clients:    make(map[*Client]bool),
		mapEvents:  make(map[string][]*Client),
	}
}


func(ws *Ws) start() {
	for {
		select {
		case cmd := <- ws.register:
			switch cmd.Type {
			case "connected":
				ws.clients[cmd.client] = true
			case "disconnected":
				// REMOVE FROM HUB
				if _, ok := ws.clients[cmd.client]; ok {
					ws.closeConnection(cmd.client)
				}
				// REMOVE RROM EVENT LIST
				for evtID := range ws.mapEvents {
					removeItemFromSlice(ws.mapEvents[evtID], cmd.client)
				}
			case "subscribed":
				if _, ok := ws.clients[cmd.client]; !ok {
					return
				}
				ws.mapEvents[cmd.payload] = append(ws.mapEvents[cmd.payload], cmd.client)
			case "unsubscribed":
				if _, ok := ws.clients[cmd.client]; !ok {
					return
				}
				removeItemFromSlice(ws.mapEvents[cmd.payload], cmd.client)
			}
		case message := <- ws.broadcast:
			// Client subEvent
			for _, client := range ws.mapEvents[message.eventId] {
				client.sender <- []byte(message.payload)
			}
		}
	}
}

func(ws*Ws) closeConnection(client *Client) {
	delete(ws.clients, client)
	close(client.sender)
}

func(client *Client)initClient(ws *Ws, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	client = &Client{
		id: makeID(13),
		conn: conn,
		sender: make(chan []byte, 1024),
	}
	ws.register <- &Command {
		Type: "connected",
	}
}


// READ DATA FROM CLIENT SENT
func(client *Client) inPump() {
	MAX_SIZE := 1024
	PONG_WAIT := 2 * time.Minute
	defer func() {
		client.ws.register <- &Command { Type: "disconnected" }
		client.conn.Close()
		close(client.sender)
	}()
	client.conn.SetReadLimit(int64(MAX_SIZE))
	client.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	client.conn.SetPongHandler(func(string) error {
		 client.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// DO Somthing
		log.Print(message)
		// client.ws.broadcast <- &Message {
		// 	created: time.Now().UnixNano(),
		// 	payload: message,
		// 	eventId:
		// }
	}
}

func(client *Client) outPump() {

}

func(client *Client) subscribe(eventID string) {
	client.ws.register <- &Command{
		client: client,
		Type: "subscribed",
		payload: eventID,
	}
}

func(client *Client) unsubscribe(eventID string) {
	client.ws.register <- &Command{
		client: client,
		Type: "unsubscribed",
		payload: eventID,
	}
}