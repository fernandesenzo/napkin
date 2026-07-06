package client

import (
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Hub interface {
	RegisterChan() chan<- *Client
	UnregisterChan() chan<- *Client
	BroadcastChan() chan<- string
	GetCode() string
}

type Client struct {
	hub     Hub
	conn    *websocket.Conn
	Send    chan string
	limiter *rate.Limiter
}

func NewClient(hub Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		Send:    make(chan string, 256),
		limiter: rate.NewLimiter(rate.Limit(10), 20), // only to prevent spam
	}
}
