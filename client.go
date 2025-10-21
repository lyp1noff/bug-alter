package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func runClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
	}(conn)

	fmt.Println("Connected to", addr)

	msgQueue := make(chan Message, 8)
	go func() {
		for {
			msg, err := readMessage(conn)
			if err != nil {
				fmt.Println("Disconnected:", err)
				close(msgQueue)
				return
			}
			msgQueue <- msg
		}
	}()

	board := initBoard(BoardSize)
	enemyBoard := initBoard(BoardSize)

	scanner := bufio.NewScanner(os.Stdin)
	for msg := range msgQueue {
		switch msg.Type {
		case MessageInit:
			var data InitData
			err := msg.DecodeData(&data)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			for _, warship := range data.Warships {
				board[warship[1]][warship[0]] = Ship
			}

			redraw()
			printBoards(board, enemyBoard)

		case MessageResult:
			var data ResultData
			err := msg.DecodeData(&data)
			if err != nil {
				return
			}

			if data.Hit {
				enemyBoard[data.Y][data.X] = Hit
			} else {
				enemyBoard[data.Y][data.X] = Miss
			}

			redraw()
			printBoards(board, enemyBoard)

			result := "Miss"
			if data.Hit {
				result = "Hit"
			}
			fmt.Printf("%s at %c%d\n", result, 'A'+data.X, data.Y+1)

		case MessageShot:
			var shot ShotData
			if msg.HasData() {
				_ = msg.DecodeData(&shot)
				shoot(board, shot.X, shot.Y)

				redraw()
				printBoards(board, enemyBoard)
				fmt.Printf("Opponent shot at %c,%d\n\n", 'A'+shot.X, shot.Y+1)
			}

		case MessageTurn:
			x, y := readCoordsAndValidate(scanner, enemyBoard)

			shotMsg := Message{Type: MessageShot}
			err = shotMsg.EncodeData(ShotData{x, y})
			if err != nil {
				return
			}
			err = sendMessage(conn, shotMsg)
			if err != nil {
				return
			}

		case MessageGameEnd:
			var data GameEndData
			err := msg.DecodeData(&data)
			if err != nil {
				return
			}

			if data.Winner {
				fmt.Println("Winner Winner Chicken Dinner!")
			} else {
				fmt.Println("You lost!")
			}

			time.Sleep(5 * time.Second)
			return
		}
	}
}
