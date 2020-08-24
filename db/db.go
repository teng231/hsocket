package db

import (
	"context"
	"log"

	pb "github.com/my0sot1s/hsocket/header"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client      *mongo.Client
	conn        *mongo.Database
	cUser       *mongo.Collection
	cConversion *mongo.Collection
	cMessage    *mongo.Collection
}

func ConnectDb(uri, dbname string) (*DB, error) {
	retryWrite := false
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.RetryWrites = &retryWrite
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	db := &DB{client: client, conn: client.Database(dbname)}
	db.cUser = db.conn.Collection("user")
	db.cConversion = db.conn.Collection("conversation")
	db.cMessage = db.conn.Collection("message")
	return db, nil
}

func (d *DB) ListUsers(req *pb.UserRequest) ([]*pb.User, error) {
	options := options.Find()
	// Sort by `_id` field descending
	options.SetSort(bson.D{{"_id", -1}})
	options.SetLimit(req.GetLimit())
	options.SetSkip((req.GetPage() - 1) * req.GetLimit())
	m := bson.M{}
	if req.GetUsername() != "" {
		m["username"] = req.GetUsername()
	}
	if req.GetFullname() != "" {
		m["fullname"] = req.GetFullname()
	}

	cursor, err := d.cUser.Find(context.TODO(), m, options)
	if err != nil {
		return nil, err
	}
	users := []*pb.User{}
	for cursor.Next(context.TODO()) {
		user := &pb.User{}
		if err := cursor.Decode(user); err != nil {
			log.Print(err)
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (d *DB) GetUser(cond *pb.UserRequest) (*pb.User, error) {
	user := &pb.User{}
	m := bson.M{}
	if cond.GetId() != "" {
		m["_id"] = cond.GetId()
	}
	if cond.GetUsername() != "" {
		m["username"] = cond.GetUsername()
	}
	err := d.cUser.FindOne(context.TODO(), m).Decode(user)
	if err != nil {
		log.Print(cond.GetId())
		return nil, err
	}
	return user, nil
}

func (d *DB) InsertUsers(users ...*pb.User) error {
	if len(users) == 1 {
		_, err := d.cUser.InsertOne(context.TODO(), users[0])
		if err != nil {
			return err
		}
		return nil
	}
	payloads := []interface{}{}
	for _, user := range users {
		payloads = append(payloads, user)
	}
	_, err := d.cUser.InsertMany(context.TODO(), payloads)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) UpdateUser(updator, selector *pb.User) error {
	settor := pb.StructToMap(updator)
	delete(settor, "id")
	resp, err := d.cUser.UpdateOne(context.TODO(), bson.M{
		"_id": selector.GetId(),
	}, bson.M{"$set": settor})
	log.Print(resp)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) ListConversations(req *pb.ConversationRequest) ([]*pb.Conversation, error) {
	options := options.Find()
	// Sort by `_id` field descending
	options.SetSort(bson.D{{"_id", -1}})
	options.SetLimit(req.GetLimit())
	options.SetSkip((req.GetPage() - 1) * req.GetLimit())
	m := bson.D{}
	if req.GetUserId() != "" {
		m = bson.D{{"members." + req.GetUserId(), bson.M{"$exists": true}}}
	}
	cursor, err := d.cConversion.Find(context.TODO(), m, options)
	if err != nil {
		return nil, err
	}
	convos := []*pb.Conversation{}
	for cursor.Next(context.TODO()) {
		convo := &pb.Conversation{}
		if err := cursor.Decode(convo); err != nil {
			continue
		}
		convos = append(convos, convo)
	}
	return convos, nil
}

func (d *DB) GetConversation(cond *pb.ConversationRequest) (*pb.Conversation, error) {
	convo := &pb.Conversation{}
	err := d.cConversion.FindOne(context.TODO(),
		bson.M{"id": cond.GetId()}).Decode(convo)
	if err != nil {
		return nil, err
	}
	return convo, nil
}

func (d *DB) InsertConversations(convos ...*pb.Conversation) error {
	if len(convos) == 1 {
		_, err := d.cConversion.InsertOne(context.TODO(), convos[0])
		if err != nil {
			return err
		}
		return nil
	}
	payloads := []interface{}{}
	for _, convo := range convos {
		payloads = append(payloads, convo)
	}
	_, err := d.cConversion.InsertMany(context.TODO(), payloads)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) UpdateConversation(updator, selector *pb.Conversation) error {
	settor := pb.StructToMap(updator)
	delete(settor, "id")
	log.Print(settor)
	_, err := d.cConversion.UpdateOne(context.TODO(), bson.M{
		"id": selector.GetId(),
	}, bson.M{
		"$set": settor,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) ListMessages(req *pb.MessageRequest) ([]*pb.Message, error) {
	options := options.Find()
	// Sort by `_id` field descending
	options.SetSort(bson.D{{"_id", -1}})
	options.SetLimit(req.GetLimit())
	options.SetSkip((req.GetPage() - 1) * req.GetLimit())
	reqConvo := bson.M{}
	if req.GetConversationId() != "" {
		reqConvo["conversationid"] = req.GetConversationId()
	}

	cursor, err := d.cMessage.Find(context.TODO(), reqConvo, options)
	if err != nil {
		return nil, err
	}
	messages := []*pb.Message{}
	for cursor.Next(context.TODO()) {
		message := &pb.Message{}
		if err := cursor.Decode(message); err != nil {
			continue
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (d *DB) GetMessage(cond *pb.MessageRequest) (*pb.Message, error) {
	message := &pb.Message{}
	err := d.cMessage.FindOne(context.TODO(), bson.M{
		"_id": cond.GetId(),
	}).Decode(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (d *DB) InsertMessages(msg ...*pb.Message) error {
	if len(msg) == 1 {
		_, err := d.cMessage.InsertOne(context.TODO(), msg[0])
		if err != nil {
			return err
		}
		return nil
	}
	payloads := []interface{}{}
	for _, m := range msg {
		payloads = append(payloads, m)
	}
	_, err := d.cMessage.InsertMany(context.TODO(), payloads)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) UpdateMessage(updator, selector *pb.Message) error {
	settor := pb.StructToMap(updator)
	delete(settor, "id")
	_, err := d.cMessage.UpdateOne(context.TODO(), bson.M{
		"_id": selector.GetId(),
	}, bson.M{
		"$set": settor,
	})
	if err != nil {
		return err
	}
	return nil
}
