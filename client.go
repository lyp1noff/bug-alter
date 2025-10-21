package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func runClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to", addr)

	scanner := bufio.NewScanner(os.Stdin)
	board := initBoard(5)
	for {
		msg, err := readMessage(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Error reading:", err)
			}
			break
		}

		if msg.Type == MessageResult {
			println(msg.Data)

			x, y := readCoordsAndValidate(scanner, board)

			msg := Message{MessageShot, ShotData{x, y}}
			err := sendMessage(conn, msg)
			if err != nil {
				return
			}
		}
	}
}
