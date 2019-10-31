package main

import (
	"context"
	"time"

	"github.com/my0sot1s/header/wsh"
)

type wsProvider struct {
	ws *Ws
}

// Broadcast trigger broadcast inside system
func (wp *wsProvider) Broadcast(ctx context.Context, in *wsh.Message) (*wsh.Resp, error) {
	wp.ws.broadcast <- in
	return &wsh.Resp{Created: time.Now().UnixNano()}, nil
}
