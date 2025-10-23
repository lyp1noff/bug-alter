package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func connectWithRetry(addr string) net.Conn {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Connecting to %s...\n", addr)
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err == nil {
			return conn
		}

		fmt.Printf("❌ Connection failed: %v\n", err)
		fmt.Print("Enter another address or press Enter to exit: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return nil
		}
		addr = input
	}
}

func runClient(addr string) {
	conn := connectWithRetry(addr)
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
	}(conn)

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			log.Printf("SetNoDelay error: %v", err)
		}
	}

	err := sendMessage(conn, Message{Type: MessageHello})
	if err != nil {
		log.Println("Error establishing connection:", err)
		return
	}
	fmt.Println("Connected")

	msgQueue := make(chan Message, 8)

	go func() {
		for {
			msg, err := readMessage(conn)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
					fmt.Println("Disconnected from server.")
				} else {
					fmt.Printf("Connection error: %v\n", err)
				}

				close(msgQueue)
				return
			}
			msgQueue <- msg
		}
	}()

	board := initBoard(DefaultBoardSize)
	enemyBoard := initBoard(DefaultBoardSize)
	var ships []Ship

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

			ships = data.Ships

			for _, ship := range data.Ships {
				for _, c := range ship.Coords {
					board[c[1]][c[0]] = CellShip
				}
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
				enemyBoard[data.Y][data.X] = CellHit
			} else {
				enemyBoard[data.Y][data.X] = CellMiss
			}

			if data.Destroyed {
				if len(data.SunkShip) > 0 {
					markAroundDestroyed(enemyBoard, data.SunkShip)
				}
			}

			redraw()
			printBoards(board, enemyBoard)

			result := "Miss"
			if data.Hit {
				result = "Hit"
			}
			if data.Destroyed {
				result = "Destroyed"
			}

			fmt.Printf("%s at %c%d\n", result, 'A'+data.X, data.Y+1)

		case MessageShot:
			var shot ShotData
			if msg.HasData() {
				_ = msg.DecodeData(&shot)
				shoot(board, ships, shot.X, shot.Y)

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
