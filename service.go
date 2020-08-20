package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	log.SetFlags(log.Lshortfile)
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
		event.Sender = client.id
		// when disconnected
		if event.NotificationType == T_disconnected {
			client.disconnect()
		}
		// when subscribe topic
		if event.NotificationType == T_subscribe {
			client.subscribe(event.SendTo, event.Sender)
		}
		// when unsubscribe topic
		if event.NotificationType == T_unsubscribe {
			client.unsubscribe(event.SendTo, event.Sender)
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
