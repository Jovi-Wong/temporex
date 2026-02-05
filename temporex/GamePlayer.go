package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type PlayerConn struct {
	createdAt time.Time
	conn      *websocket.Conn
}

type GamePlayer struct {
	playerID string
	groups   []string
}

func NewPlayerConn(conn *websocket.Conn) *PlayerConn {
	return &PlayerConn{
		createdAt: time.Now(),
		conn:      conn,
	}
}

func NewGamePlayer(playerID string) *GamePlayer {
	return &GamePlayer{
		playerID: playerID,
		groups:   []string{},
	}
}
