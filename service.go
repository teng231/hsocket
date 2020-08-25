package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	GetUser(req *pb.UserRequest) (*pb.User, error)
	InsertUsers(user ...*pb.User) error
	UpdateUser(updator, selector *pb.User) error
	ListConversations(req *pb.ConversationRequest) ([]*pb.Conversation, error)
	GetConversation(req *pb.ConversationRequest) (*pb.Conversation, error)
	InsertConversations(convos ...*pb.Conversation) error
	UpdateConversation(updator, selector *pb.Conversation) error
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
	r.POST("/send", func(c *gin.Context) {
		w, r := c.Writer, c.Request
		event := wsFirer(hub, w, r)
		err := conv.db.InsertMessages(&pb.Message{
			Id:             pb.MakeId(),
			ConversationId: event.ConversationId,
			SenderId:       event.SenderId,
			Text:           event.Text,
			Created:        time.Now().Unix(),
			Type:           pb.Message_raw.String(),
		})
		log.Print("save message error", err)
	})
	r.GET("/users", func(c *gin.Context) {
		rq := &pb.UserRequest{}
		BindQuery(rq, c)
		if rq.GetLimit() == 0 {
			rq.Limit = 10
		}
		if rq.GetPage() == 0 {
			rq.Page = 1
		}
		log.Print(rq)
		users, err := conv.db.ListUsers(rq)
		if err != nil {
			c.JSON(500, err)
			return
		}
		c.JSON(200, users)
	})
	r.GET("/users/:username", func(c *gin.Context) {
		rq := &pb.UserRequest{}
		rq.Username = c.Param("username")
		user, err := conv.db.GetUser(rq)
		if err != nil {
			c.JSON(500, err)
			return
		}
		c.JSON(200, user)
	})
	r.GET("/conversations", func(c *gin.Context) {
		rq := &pb.ConversationRequest{}
		BindQuery(rq, c)
		if rq.GetLimit() == 0 {
			rq.Limit = 10
		}
		if rq.GetPage() == 0 {
			rq.Page = 1
		}
		log.Print(rq)
		convo, err := conv.db.ListConversations(rq)
		if err != nil {
			c.JSON(500, err)
			return
		}
		c.JSON(200, convo)
	})
	r.POST("/conversations", func(c *gin.Context) {
		convo := &pb.Conversation{}
		c.ShouldBindJSON(convo)
		convo.Type = pb.Conversation_chat.String() // chat 2 nguoi
		convo.Id = pb.MakeId()
		if convo.Members == nil {
			convo.Members = map[string]*pb.User{
				convo.GetCreatorId(): convo.GetCreator(),
			}
		} else {
			convo.Members[convo.GetCreatorId()] = convo.GetCreator()
		}

		err := conv.db.InsertConversations(convo)
		if err != nil {
			c.JSON(500, err)
			return
		}
		newConvo, err := conv.db.GetConversation(&pb.ConversationRequest{Id: convo.GetId()})
		if err != nil {
			c.JSON(500, err)
			return
		}
		c.JSON(200, newConvo)
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

func wsFirer(hub *WsHub, w http.ResponseWriter, r *http.Request) *Event {
	if r.Method != "POST" {
		return nil
	}
	event := &Event{}
	if err := json.NewDecoder(r.Body).Decode(event); err != nil {
		http.Error(w, err.Error(), 400)
		return nil
	}
	hub.broadcast <- event
	w.Write([]byte("done"))
	return event
}

func handleListTopics(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	bin, _ := json.Marshal(hub.topics)
	w.Write(bin)
}
