package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var MAX_SIZE = 1024
var PONG_WAIT = 2 * time.Minute
var WRITE_WAIT = 10 * time.Second
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	payload []byte
	created int64
	Type    string
	eventId string
}

type Command struct {
	client  *Client
	Type    string
	payload string
}

type Ws struct {
	clients   map[*Client]bool
	broadcast chan *Message
	register  chan *Command
	mapEvents map[string][]*Client
}

type Client struct {
	id     string
	ws     *Ws
	conn   *websocket.Conn
	sender chan []byte
}

func initWs() *Ws {
	return &Ws{
		broadcast: make(chan *Message),
		register:  make(chan *Command),
		clients:   make(map[*Client]bool),
		mapEvents: make(map[string][]*Client),
	}
}

func (ws *Ws) start() {
	for {
		select {
		case cmd := <-ws.register:
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
		case message := <-ws.broadcast:
			// Client subEvent
			for _, client := range ws.mapEvents[message.eventId] {
				client.sender <- []byte(message.payload)
			}
		}
	}
}

func (ws *Ws) closeConnection(client *Client) {
	delete(ws.clients, client)
	close(client.sender)
}

func initClient(ws *Ws, w http.ResponseWriter, r *http.Request) *Client {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	client := &Client{
		id:     makeID(17),
		conn:   conn,
		ws:     ws,
		sender: make(chan []byte, 1024),
	}
	ws.register <- &Command{
		Type:   "connected",
		client: client,
	}
	return client
}

// READ DATA FROM CLIENT SENT
func (client *Client) inPump() {
	defer func() {
		client.ws.register <- &Command{Type: "disconnected", client: client}
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
		log.Print("message received: ", string(message))
		cmd := &Command{}
		if err := json.Unmarshal(message, cmd); err != nil {
			log.Fatal(err)
		}
		// just using for command
		cmd.client = client
		client.ws.register <- cmd
	}
}

func (client *Client) outPump() {
	ticker := time.NewTicker(1 * time.Minute)
	defer func() {
		ticker.Stop()
		client.ws.register <- &Command{Type: "disconnected", client: client}
		client.conn.Close()
		close(client.sender)
	}()
	for {
		select {
		case message, ok := <-client.sender:
			client.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if !ok {
				// The hub closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			for i := 0; i < len(message); i++ {
				w.Write(message)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) subscribe(eventID string) {
	client.ws.register <- &Command{
		client:  client,
		Type:    "subscribed",
		payload: eventID,
	}
}

func (client *Client) unsubscribe(eventID string) {
	client.ws.register <- &Command{
		client:  client,
		Type:    "unsubscribed",
		payload: eventID,
	}
}
