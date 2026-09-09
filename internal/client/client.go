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
	sendBufferSize = 256
)

type Config struct {
	MaxContentLength int
}

const readLimitHeadroom = 1024

type Hub interface {
	RegisterChan() chan<- *Client
	UnregisterChan() chan<- *Client
	BroadcastChan() chan<- string
	GetCode() string
}

type Client struct {
	hub              Hub
	conn             *websocket.Conn
	Send             chan string
	limiter          *rate.Limiter
	maxContentLength int
}

func NewClient(hub Hub, conn *websocket.Conn, config Config) *Client {
	return &Client{
		hub:              hub,
		conn:             conn,
		Send:             make(chan string, sendBufferSize),
		limiter:          rate.NewLimiter(rate.Limit(10), 20), // only to prevent spam
		maxContentLength: config.MaxContentLength,
	}
}

func (c *Client) maxMessageSize() int {
	if c.maxContentLength <= 0 {
		return readLimitHeadroom
	}
	return c.maxContentLength*4 + readLimitHeadroom
}
