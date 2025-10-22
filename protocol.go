package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Player struct {
	Conn  net.Conn
	Board [][]byte
	Ships []Ship
	In    chan Message
	Out   chan Message
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
	Ships []Ship `json:"ships"`
}

type ShotData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ResultData struct {
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Hit       bool     `json:"hit"`
	Destroyed bool     `json:"destroyed"`
	SunkShip  [][2]int `json:"sunk_ship,omitempty"`
}

type GameEndData struct {
	Winner bool `json:"winner"`
}

func sendMessage(conn net.Conn, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	lenb := make([]byte, 4)
	binary.BigEndian.PutUint32(lenb, uint32(len(data)))
	if _, err := conn.Write(lenb); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func readMessage(conn net.Conn) (Message, error) {
	var lenb [4]byte
	if _, err := io.ReadFull(conn, lenb[:]); err != nil {
		return Message{}, err
	}
	length := binary.BigEndian.Uint32(lenb[:])
	if length > 1<<20 {
		return Message{}, fmt.Errorf("message too large: %d", length)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return Message{}, err
	}
	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return Message{}, fmt.Errorf("json unmarshal: %w", err)
	}
	return msg, nil
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
