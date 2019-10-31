package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/my0sot1s/header/wsh"
)

var MAX_SIZE = 1024
var PONG_WAIT = 2 * time.Minute
var WRITE_WAIT = 10 * time.Second
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Command struct {
	client *Client
	Type   string
	Topic  string
	MsgID  string
}

type Ws struct {
	clients   map[*Client]bool
	broadcast chan *wsh.Message
	register  chan *Command
	mapTopics map[string][]*Client
}

type Client struct {
	id     string
	ws     *Ws
	conn   *websocket.Conn
	sender chan *wsh.Message
}

func initWs() *Ws {
	return &Ws{
		broadcast: make(chan *wsh.Message),
		register:  make(chan *Command),
		clients:   make(map[*Client]bool),
		mapTopics: make(map[string][]*Client),
	}
}

func (ws *Ws) start() {
	for {
		select {
		case cmd := <-ws.register:
			switch cmd.Type {
			case "connected":
				log.Print("-- connected  ", cmd.client.id)
				ws.clients[cmd.client] = true
			case "disconnected":
				log.Print(cmd.client.id, " disconnected")
				// REMOVE FROM HUB
				if _, ok := ws.clients[cmd.client]; ok {
					ws.closeConnection(cmd.client)
				}
				// REMOVE RROM EVENT LIST
				for topic, clients := range ws.mapTopics {
					if len(clients) == 0 {
						break
					}
					ws.mapTopics[topic] = removeItemFromSlice(clients, cmd.client)
				}
			case "subscribe":
				if _, ok := ws.clients[cmd.client]; !ok {
					return
				}
				ws.mapTopics[cmd.Topic] = append(ws.mapTopics[cmd.Topic], cmd.client)
			case "unsubscribe":
				if _, ok := ws.clients[cmd.client]; !ok {
					return
				}
				ws.mapTopics[cmd.Topic] = removeItemFromSlice(ws.mapTopics[cmd.Topic], cmd.client)
			}
		case message := <-ws.broadcast:
			// Client subEvent
			for _, client := range ws.mapTopics[message.Topic] {
				select {
				case client.sender <- message:
				}
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
		sender: make(chan *wsh.Message),
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
				log.Printf(client.id, "disconnected")
				client.conn.Close()
			}
			break
		}
		// DO Somthing
		log.Print("message received: ", string(message))
		cmd := &Command{}
		if err := json.Unmarshal(message, cmd); err != nil {
			log.Fatal(err)
		}

		if cmd.Type == "subscribe" {
			client.subscribe(cmd.Topic)
			client.ws.broadcast <- &wsh.Message{
				Body:  fmt.Sprintf("%s joined at %d", client.id, time.Now().UnixNano()),
				Topic: cmd.Topic,
			}
		} else if cmd.Type == "unsubscribe" {
			client.unsubscribe(cmd.Topic)
		} else {
			// just using for command
			cmd.client = client
			client.ws.register <- cmd
		}
	}
}

func (client *Client) outPump() {
	ticker := time.NewTicker(1 * time.Minute)
	defer func() {
		ticker.Stop()
		client.ws.register <- &Command{Type: "disconnected", client: client}
		client.conn.Close()
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
				log.Print("ERRR")
				return
			}
			var b []byte
			if b, err = json.Marshal(message); err != nil {
				log.Print("ERRR", err)
				return
			}
			w.Write(b)

			if err := w.Close(); err != nil {
				log.Print("ERRR", err)
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

func (client *Client) closeConnection() {
	client.ws.register <- &Command{
		client: client,
		Type:   "disconected",
	}
	client.conn.Close()
}

func (client *Client) subscribe(topic string) {
	client.ws.register <- &Command{
		client: client,
		Type:   "subscribe",
		Topic:  topic,
	}
}

func (client *Client) unsubscribe(topic string) {
	client.ws.register <- &Command{
		client: client,
		Type:   "unsubscribe",
		Topic:  topic,
	}
}
