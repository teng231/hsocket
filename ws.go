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
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var (
	T_connected    = "connected"
	T_disconnected = "disconnected"
	T_subscribe    = "subscribe"
	T_unsubscribe  = "unsubscribe"

	E_json = "application/json"
	E_text = "text/plain"
)

type Event struct {
	client *Client

	Id               string `json:"id"`
	ConnId           string `json:"conn_id"`
	Text             string `json:"text"`
	Encoding         string `json:"encoding"`
	SenderId         string `json:"sender_id"`
	ConversationId   string `json:"conversation_id"`
	NotificationType string `json:"notification_type"`
	State            string `json:"state"`
	Created          int64  `json:"created"`
}
type Topic struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	members map[string]*Client
}

type WsHub struct {
	clients   map[string]*Client
	broadcast chan *Event
	topics    map[string]*Topic
}

type Client struct {
	id     string
	hub    *WsHub
	conn   *websocket.Conn
	sender chan *Event
}

func (hub *WsHub) onStart() {
	for {
		select {
		case event := <-hub.broadcast:
			if event.ConversationId != "" {
				if hub.topics[event.ConversationId] == nil {
					topic := &Topic{
						members: make(map[string]*Client),
						Name:    event.ConversationId,
					}
					hub.topics[event.ConversationId] = topic
				}
			}
			// when connected
			if event.NotificationType == T_connected {
				hub.clients[event.client.id] = event.client
			}
			// when disconnected
			if event.NotificationType == T_disconnected {
				for _, topic := range hub.topics {
					if len(topic.members) == 0 {
						continue
					}
					delete(topic.members, event.client.id)
				}
			}
			// when subscribe topic
			if event.NotificationType == T_subscribe {
				if event.ConversationId != "" {
					hub.topics[event.ConversationId].members[event.client.id] = event.client
				}
			}
			// when unsubscribe topic
			if event.NotificationType == T_unsubscribe {
				if event.ConversationId != "" {
					if hub.topics[event.ConversationId] != nil {
						delete(hub.topics[event.ConversationId].members, event.client.id)
					}
				}
			}

			// log.Print(event.SendTo, hub.topics[event.SendTo])
			if hub.topics[event.ConversationId] == nil {
				continue
			}
			// Client subEvent
			for _, client := range hub.topics[event.ConversationId].members {
				client.sender <- event
			}
		}
	}
}

func (hub *WsHub) closeConnection(client *Client) {
	delete(hub.clients, client.id)
	close(client.sender)
}

func makeWsClient(hub *WsHub, w http.ResponseWriter, r *http.Request) *Client {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	client := &Client{
		id:     "client." + makeID(17),
		conn:   conn,
		hub:    hub,
		sender: make(chan *Event, 200),
	}
	client.sender <- &Event{
		ConnId:           client.id,
		Encoding:         E_text,
		NotificationType: T_connected,
	}
	return client
}

func (client *Client) disconnect() {
	client.hub.broadcast <- &Event{
		Id:               "evt" + makeID(10),
		NotificationType: T_disconnected,
		client:           client,
		Text:             client.id + " disconected",
		Created:          time.Now().Unix(),
		Encoding:         E_text,
	}
	client.conn.Close()
}

func (client *Client) subscribe(topic, sender string) {
	client.hub.broadcast <- &Event{
		Id:               "evt" + makeID(10),
		NotificationType: T_subscribe,
		client:           client,
		Text:             client.id + " subscribed " + topic,
		Created:          time.Now().Unix(),
		ConversationId:   topic,
		Encoding:         E_text,
		SenderId:         sender,
	}
}

func (client *Client) unsubscribe(topic, sender string) {
	client.hub.broadcast <- &Event{
		Id:               "evt" + makeID(10),
		NotificationType: T_unsubscribe,
		client:           client,
		Text:             client.id + " unsubscribed " + topic,
		Created:          time.Now().Unix(),
		ConversationId:   topic,
		Encoding:         E_text,
		SenderId:         sender,
	}
}

func (client *Client) sendMultipleUser(topic, message string) {
	client.hub.broadcast <- &Event{
		Text:           message,
		ConversationId: topic,
		client:         client,
		Encoding:       E_json,
		Created:        time.Now().Unix(),
	}
}

func (hub *WsHub) getClient(sendTo, connid string) *Client {
	if hub.topics[sendTo] == nil {
		hub.topics[sendTo] = &Topic{
			Name:    sendTo,
			members: make(map[string]*Client),
		}
	}
	client := hub.topics[sendTo].members[connid]
	log.Print(client)
	return client
}
func initWs() *WsHub {
	return &WsHub{
		broadcast: make(chan *Event),
		clients:   make(map[string]*Client),
		topics:    make(map[string]*Topic),
	}
}

// READ DATA FROM CLIENT SENT
func (client *Client) onWsListenMessage() {
	// defer func() {
	// 	log.Print("xxxx2")
	// 	client.disconnect()
	// }()
	client.conn.SetReadLimit(int64(MAX_SIZE))
	client.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})
	for {
		_, evt, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("%v disconnected", client.id)
				client.disconnect()
			}
			break
		}
		// DO Somthing
		log.Print("message received: ", string(evt))
		if evt == nil || string(evt) == "" {
			continue
		}
		event := &Event{}
		if err := json.Unmarshal(evt, event); err != nil {
			log.Printf("err %v message %s", err, evt)
			break
		}
		event.client = client
		event.SenderId = client.id
		// when disconnected
		if event.NotificationType == T_disconnected {
			client.disconnect()
		}
		// when subscribe topic
		if event.NotificationType == T_subscribe {
			client.subscribe(event.ConversationId, event.SenderId)
		}
		// when unsubscribe topic
		if event.NotificationType == T_unsubscribe {
			client.unsubscribe(event.ConversationId, event.SenderId)
		}
		// // Client subEvent
		// for _, client := range client.hub.topics[event.SendTo].members {
		// 	client.sender <- event
		// }
	}
}

func (client *Client) onWsPushMessage() {
	ticker := time.NewTicker(1 * time.Minute)
	defer func() {
		log.Print("clossssseddd")
		ticker.Stop()
		client.disconnect()
	}()
	for {
		select {
		case message := <-client.sender:
			client.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			writer, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("ERRR: %v", err)
				client.disconnect()
				return
				// continue
			}
			messageBuffer, _ := json.Marshal(message)
			writer.Write(messageBuffer)

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				continue
			}
		}
	}
}
