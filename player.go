package main

import "github.com/gorilla/websocket"

type Transfer struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

type Point struct {
	X float64
	Y float64
}

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type PlayerType struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type Player struct {
	Conn   *websocket.Conn `json:"-"`
	Name   string          `json:"name"`
	Coords Point           `json:"coords"`
	Radius float64         `json:"radius"`
	Color  string          `json:"color"`
	IsDead bool            `json:"isDead"`
}

func (p *Point) IsMove() bool {
	return p.X != 0 || p.Y != 0
}
