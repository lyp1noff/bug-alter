package main

import (
	"fmt"
	"net"
	"time"
)

func runServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on %s\n", addr)

	players := make([]*Player, 0, 2)

	for len(players) < 2 {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		players = append(players, &Player{conn, initBoard(Size), make(chan Message, 1), make(chan Message, 1)})
		fmt.Println("Player connected:", conn.RemoteAddr())
	}

	for _, player := range players {
		placeWarships(player.Board, EnemyWarshipsCount)
		err := sendMessage(player.Conn, Message{
			Type: MessageInit,
			Data: nil,
		})
		if err != nil {
			return
		}
	}

	for _, player := range players {
		go func(player *Player) {
			for {
				inMsg, err := readMessage(player.Conn)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				player.In <- inMsg
			}
		}(player)
	}

	for _, player := range players {
		go func(player *Player) {
			for {
				msg := <-player.Out
				err := sendMessage(player.Conn, msg)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
			}
		}(player)
	}

	time.Sleep(200 * time.Millisecond)
	players[0].Out <- Message{Type: MessageResult, Data: nil}
	for {
		select {
		case msg := <-players[0].In:
			fmt.Println("Player 1:", msg)
			gameProcessor(players[0], players[1], msg)

		case msg := <-players[1].In:
			fmt.Println("Player 2:", msg)
			gameProcessor(players[1], players[0], msg)
		}
	}
}

func gameProcessor(playerFrom, playerTo *Player, msg Message) {
	switch msg.Type {
	case MessageShot:
		data := msg.Data.(map[string]interface{})
		x := int(data["x"].(float64))
		y := int(data["y"].(float64))

		hit := shoot(playerTo.Board, x, y)

		playerFrom.Out <- Message{
			Type: MessageResult,
			Data: ResultData{Hit: hit, GameOver: false},
		}

		if hit {
			playerFrom.Out <- Message{Type: MessageTurn, Data: nil}
		} else {
			playerTo.Out <- Message{Type: MessageTurn, Data: nil}
		}
	}
}
