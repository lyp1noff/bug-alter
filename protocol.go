package main

import (
	"encoding/json"
	"net"
)

type Player struct {
	Conn  net.Conn
	Board [][]byte
	In    chan Message
	Out   chan Message
}

type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

type MessageType string

const (
	MessageInit   MessageType = "init"
	MessageShot   MessageType = "shot"
	MessageResult MessageType = "result"
	MessageTurn   MessageType = "turn"
)

type ShotData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ResultData struct {
	Hit      bool `json:"hit"`
	GameOver bool `json:"game_over"`
}

func sendMessage(conn net.Conn, msg Message) error {
	return json.NewEncoder(conn).Encode(msg)
}

func readMessage(conn net.Conn) (Message, error) {
	var msg Message
	err := json.NewDecoder(conn).Decode(&msg)
	return msg, err
}
