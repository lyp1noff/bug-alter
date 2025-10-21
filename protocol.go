package main

import (
	"encoding/json"
	"net"
)

type Player struct {
	Conn     net.Conn
	Board    [][]byte
	Warships [][]int
	In       chan Message
	Out      chan Message
}

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MessageType string

const (
	MessageInit    MessageType = "init"
	MessageShot    MessageType = "shot"
	MessageResult  MessageType = "result"
	MessageTurn    MessageType = "turn"
	MessageGameEnd MessageType = "game_end"
)

type InitData struct {
	Warships [][]int `json:"warships"`
}

type ShotData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ResultData struct {
	X   int  `json:"x"`
	Y   int  `json:"y"`
	Hit bool `json:"hit"`
}

type GameEndData struct {
	Winner bool `json:"winner"`
}

func sendMessage(conn net.Conn, msg Message) error {
	return json.NewEncoder(conn).Encode(msg)
}

func readMessage(conn net.Conn) (Message, error) {
	var msg Message
	err := json.NewDecoder(conn).Decode(&msg)
	return msg, err
}

func (m *Message) DecodeData(v any) error {
	return json.Unmarshal(m.Data, v)
}

func (m *Message) EncodeData(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Data = data
	return nil
}

func (m *Message) HasData() bool {
	return len(m.Data) > 0 && string(m.Data) != "null"
}
