package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
	"github.com/my0sot1s/hsocket/db"
	pb "github.com/my0sot1s/hsocket/header"
)

type Convo struct {
	db IDB
}

func init() {
	log.SetFlags(log.Lshortfile)
	decoder.SetAliasTag("json")
}

type IDB interface {
	ListUsers(req *pb.UserRequest) ([]*pb.User, error)
	GetUser(req *pb.User) (*pb.User, error)
	InsertUsers(user ...*pb.User) error
	UpdateUser(updator, selector *pb.User) error
	// ListConversations(req map[string]interface{}) ([]*pb.Conversation, error)
	// GetConversation(req map[string]interface{}) (*pb.Conversation, error)
	// InsertConversation(convo *pb.Conversation) error
	// UpdateConversation(updator, selector *pb.Conversation) error
	ListMessages(req *pb.MessageRequest) ([]*pb.Message, error)
	GetMessage(req *pb.MessageRequest) (*pb.Message, error)
	InsertMessages(msg ...*pb.Message) error
	// UpdateMessage(updator, selector *pb.Message) error
}

var decoder = schema.NewDecoder()

func BindQuery(in interface{}, ctx *gin.Context) error {
	err := decoder.Decode(in, ctx.Request.URL.Query())
	return err
}

func start() {
	hub := initWs()
	go hub.onStart()
	r := gin.Default()
	r.Use(cors.Default())
	dbConnect, err := db.ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	if err != nil {
		panic(err)
	}
	conv := &Convo{
		db: dbConnect,
	}
	r.GET("/", func(c *gin.Context) {
		c.String(200, "We got Gin")
	})
	r.GET("/topics", func(c *gin.Context) {
		w, r := c.Writer, c.Request
		handleListTopics(hub, w, r)
	})
	r.POST("/ws-firer", func(c *gin.Context) {
		w, r := c.Writer, c.Request
		wsFirer(hub, w, r)
	})
	r.GET("/users", func(c *gin.Context) {
		// w, r := c.Writer, c.Request
		// enableCors(&w)
	})
	r.GET("/conversations", func(c *gin.Context) {
		// w, r := c.Writer, c.Request
		// enableCors(&w)
	})
	r.GET("/messages/:convoid", func(c *gin.Context) {
		rq := &pb.MessageRequest{}
		BindQuery(rq, c)
		rq.ConversationId = c.Param("convoid")
		messages, err := conv.ListMessages(rq)
		if err != nil {
			c.JSON(500, err)
			return
		}
		c.JSON(200, messages)
	})
	r.GET("/ws", func(c *gin.Context) {
		w, r := c.Writer, c.Request
		client := makeWsClient(hub, w, r)
		go client.onWsListenMessage()
		go client.onWsPushMessage()
	})
	r.Run(port)
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
	hub.broadcast <- event
	w.Write([]byte("done"))
}

func handleListTopics(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	bin, _ := json.Marshal(hub.topics)
	w.Write(bin)
}
