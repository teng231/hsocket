package main

import (
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
	Raw              string `json:"raw"`
	Encoding         string `json:"encoding"`
	Sender           string `json:"sender"`
	ToGroup          bool   `json:"to_group"`
	SendTo           string `json:"send_to"` // topic_id || user
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
			if event.ToGroup {
				if hub.topics[event.SendTo] == nil {
					topic := &Topic{
						members: make(map[string]*Client),
						Name:    event.SendTo,
					}
					hub.topics[event.SendTo] = topic
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
				if event.ToGroup {
					hub.topics[event.SendTo].members[event.client.id] = event.client
				}
			}
			// when unsubscribe topic
			if event.NotificationType == T_unsubscribe {
				if event.ToGroup {
					if hub.topics[event.SendTo] != nil {
						delete(hub.topics[event.SendTo].members, event.client.id)
					}
				}
			}

			// log.Print(event.SendTo, hub.topics[event.SendTo])
			if hub.topics[event.SendTo] == nil {
				continue
			}
			// Client subEvent
			for _, client := range hub.topics[event.SendTo].members {
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
		Raw:              client.id + " disconected",
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
		Raw:              client.id + " subscribed " + topic,
		Created:          time.Now().Unix(),
		SendTo:           topic,
		ToGroup:          true,
		Encoding:         E_text,
		Sender:           sender,
	}
}

func (client *Client) unsubscribe(topic, sender string) {
	client.hub.broadcast <- &Event{
		Id:               "evt" + makeID(10),
		NotificationType: T_unsubscribe,
		client:           client,
		ToGroup:          true,
		Raw:              client.id + " unsubscribed " + topic,
		Created:          time.Now().Unix(),
		SendTo:           topic,
		Encoding:         E_text,
		Sender:           sender,
	}
}

func (client *Client) sendMultipleUser(topic, message string) {
	client.hub.broadcast <- &Event{
		Raw:      message,
		SendTo:   topic,
		ToGroup:  true,
		client:   client,
		Encoding: E_json,
		Created:  time.Now().Unix(),
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
