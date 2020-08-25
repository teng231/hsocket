package db

import (
	"log"
	"testing"
	"time"

	pb "github.com/my0sot1s/hsocket/header"
)

func Test_connect(t *testing.T) {
	_, err := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	log.Print(err)
}

func Test_createUser(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	err := d.InsertUsers(&pb.User{
		Id:       pb.MakeId(),
		Username: "hoahong",
		Fullname: "Hoa hong",
		Created:  time.Now().Unix(),
		Avatar:   "https://vcdn-vnexpress.vnecdn.net/2020/07/23/ngoc-miu-3-1349-1595472279.jpg",
	}, &pb.User{
		Id:       pb.MakeId(),
		Username: "hiepto",
		Fullname: "Hiep To",
		Created:  time.Now().Unix(),
		Avatar:   "https://znews-photo.zadn.vn/w660/Uploaded/qfssu/2019_12_24/7_zing.jpg",
	}, &pb.User{
		Id:       pb.MakeId(),
		Username: "teng2",
		Fullname: "Te nguyen 2",
		Created:  time.Now().Unix(),
		Avatar:   "https://i.ytimg.com/vi/bWh83yaA0k0/maxresdefault.jpg",
	},
	)
	if err != nil {
		t.Fatal(err)
	}
}
func Test_listUsers(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	users, err := d.ListUsers(&pb.UserRequest{Limit: 2, Page: 1, Username: "teng"})
	if err != nil {
		t.Fatal(err)
	}
	log.Print(users)

	// users, err = d.ListUsers(&pb.UserRequest{Limit: 2, Page: 1, Fullname: "nguyen", Username: "teng"})
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// log.Print(users)

	users, err = d.ListUsers(&pb.UserRequest{Limit: 2, Page: 1, Fullname: "Te"})
	if err != nil {
		t.Fatal(err)
	}

	log.Print(users)
}
func Test_getUser(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	user, err := d.GetUser(&pb.UserRequest{Id: "5f4297a88aa6e74f1b2edcce"})
	if err != nil {
		t.Fatal(err)
	}
	log.Print(user)
}

func Test_updateUser(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	err := d.UpdateUser(&pb.User{
		Phone:   "0373140511",
		Created: time.Now().Unix(),
		// Username: "thudiu",
		// Fullname: "thu Diu",
		// Id: "5f41fec73af329b9755256ce",
	}, &pb.User{Id: "5f41fec73af329b9755256ce"})

	if err != nil {
		t.Fatal(err)
	}
}

var textDemo = []string{
	"making it over 2000 years old. Richard McClintock, a Latin professor at",
	"Hampden-Sydney College in Virginia, looked up one of the more obscure Latin",
	"words, consectetur, from a Lorem Ipsum passage, and going through the cites of",
	"the word in classical literature, discovered the undoubtable source.",
	"Lorem Ipsum",
	"comes from sections 1.10.32 and 1.10.33 of",
	"de Finibus Bonorum et Malorum",
	"(The Extremes of Good and Evil) by Cicero, written in 45 BC.",
	"This book is a treatise",
	"on the theory of ethics, very popular during the Renaissance. The first line of",
	"Lorem Ipsum, Lorem ipsum dolor sit amet..,",
	"comes from a line in section 1.10.32.",
}

func Test_insertMessage(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	msgs := make([]*pb.Message, 0)
	for i := 0; i < 11; i++ {
		msgs = append(msgs, &pb.Message{
			Id:             pb.MakeId(),
			Type:           pb.Message_raw.String(),
			ConversationId: "topic.general",
			SenderId:       "5f4297a88aa6e74f1b2edcce",
			Text:           textDemo[i],
			Created:        time.Now().Unix(),
			State:          pb.Message_sent.String(),
		})
	}
	err := d.InsertMessages(msgs...)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_listMessages(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	msgs, err := d.ListMessages(&pb.MessageRequest{Limit: 2, Page: 1, ConversationId: "topic.x"})
	if err != nil {
		t.Fatal(err)
	}
	log.Print(msgs)
}

func Test_insertConvo(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	// err := d.InsertConversations(&pb.Conversation{
	// 	Id:      "topic.general",
	// 	Type:    pb.Conversation_room.String(),
	// 	Created: time.Now().Unix(),
	// 	State:   pb.Conversation_active.String(),
	// }, &pb.Conversation{
	// 	Id:      "topic.room",
	// 	Type:    pb.Conversation_room.String(),
	// 	Created: time.Now().Unix(),
	// 	State:   pb.Conversation_active.String(),
	// })
	user1, err := d.GetUser(&pb.UserRequest{Id: "5f4297a88aa6e74f1b2edcd0"})
	user2, err := d.GetUser(&pb.UserRequest{Id: "5f4297a88aa6e74f1b2edcce"})
	err = d.InsertConversations(&pb.Conversation{
		Id:      pb.MakeId(),
		Type:    pb.Conversation_chat.String(),
		Created: time.Now().Unix(),
		Members: map[string]*pb.User{
			"5f4297a88aa6e74f1b2edcd0": user1,
			"5f4297a88aa6e74f1b2edcce": user2,
		},
		State: pb.Conversation_active.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_updateConvo(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	user, err := d.GetUser(&pb.UserRequest{Id: "5f4297a88aa6e74f1b2edcd0"})
	if err != nil {
		t.Fatal(err)
	}
	err = d.UpdateConversation(&pb.Conversation{
		Members: map[string]*pb.User{
			"5f4297a88aa6e74f1b2edcd0": user,
		},
	}, &pb.Conversation{
		Id: "topic.general",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func Test_listConvo(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	msgs, err := d.ListConversations(&pb.ConversationRequest{Limit: 2, Page: 1, UserId: "5f4297a88aa6e74f1b2edcd0"})
	if err != nil {
		t.Fatal(err)
	}
	log.Print(msgs)
}
func Test_getConvo(t *testing.T) {
	d, _ := ConnectDb("mongodb://admin:1qazxcvbnm@ds255924.mlab.com:55924/conversation", "conversation")
	convo, err := d.GetConversation(&pb.ConversationRequest{Id: "topic.general"})
	if err != nil {
		t.Fatal(err)
	}
	log.Print(convo)
}
