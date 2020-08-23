package main

import pb "github.com/my0sot1s/hsocket/header"

func (c *Convo) ListMessages(req *pb.MessageRequest) (*pb.Messages, error) {
	if req.GetLimit() == 0 {
		req.Limit = 15
	}
	if req.GetPage() == 0 {
		req.Page = 1
	}
	messages, err := c.db.ListMessages(req)
	if err != nil {
		return nil, err
	}
	return &pb.Messages{
		Messages: messages,
		NextPage: req.GetLimit() + 1,
	}, nil
}

func (c *Convo) CreateMessage(msg *pb.Message) error {
	err := c.db.InsertMessages(msg)
	if err != nil {
		return err
	}
	return nil
}
