package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func runServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener:", err)
			return
		}
	}(listener)

	log.Printf("Server is listening on %s\n", addr)

	disconnect := make(chan *Player, 1)
	players := make([]*Player, 0, 2)

	for len(players) < 2 {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				log.Printf("SetNoDelay error: %v", err)
			}
		}

		players = append(players, &Player{
			Conn:     conn,
			Board:    initBoard(BoardSize),
			Warships: make([][]int, 0, EnemyWarshipsCount),
			In:       make(chan Message, 8),
			Out:      make(chan Message, 8),
		})
		log.Println("Player connected:", conn.RemoteAddr())
	}

	for _, player := range players {
		go func(player *Player) {
			for {
				inMsg, err := readMessage(player.Conn)
				if err != nil {
					log.Println("Error:", err)
					disconnect <- player
					_ = player.Conn.Close()
					return
				}
				player.In <- inMsg
			}
		}(player)

		go func(player *Player) {
			for {
				msg := <-player.Out
				err := sendMessage(player.Conn, msg)
				log.Printf("[Writer %p] Sending %s to %s\n", player, msg.Type, player.Conn.RemoteAddr())
				if err != nil {
					log.Println("Error:", err)
					return
				}
			}
		}(player)

		player.Warships = placeWarships(player.Board, EnemyWarshipsCount)

		msg := Message{Type: MessageInit}
		err := msg.EncodeData(InitData{player.Warships})
		if err != nil {
			return
		}
		player.Out <- msg
	}

	time.Sleep(200 * time.Millisecond)
	players[0].Out <- Message{Type: MessageTurn, Data: nil}

	for {
		select {
		case msg := <-players[0].In:
			log.Println("Player 0:", msg.Type)
			gameProcessor(players[0], players[1], msg)

		case msg := <-players[1].In:
			log.Println("Player 1:", msg.Type)
			gameProcessor(players[1], players[0], msg)

		case p := <-disconnect:
			log.Println("Disconnect:", p.Conn.RemoteAddr())
			for _, pl := range players {
				close(pl.Out)
				_ = pl.Conn.Close()
			}
			return
		}
	}
}

func gameProcessor(playerFrom, playerTo *Player, msg Message) {
	switch msg.Type {
	case MessageShot:
		var shot ShotData
		err := msg.DecodeData(&shot)
		if err != nil {
			return
		}

		hit := shoot(playerTo.Board, shot.X, shot.Y)

		resultMsg := Message{Type: MessageResult}
		err = resultMsg.EncodeData(ResultData{X: shot.X, Y: shot.Y, Hit: hit})
		if err != nil {
			return
		}
		playerFrom.Out <- resultMsg

		shotMsg := Message{Type: MessageShot}
		err = shotMsg.EncodeData(ShotData{shot.X, shot.Y})
		if err != nil {
			return
		}
		playerTo.Out <- shotMsg

		loserList := []bool{isGameOver(playerFrom.Board), isGameOver(playerTo.Board)}
		if loserList[0] || loserList[1] {
			for i, player := range []*Player{playerFrom, playerTo} {
				endMsg := Message{Type: MessageGameEnd}
				err := endMsg.EncodeData(GameEndData{Winner: !loserList[i]})
				if err != nil {
					return
				}
				player.Out <- endMsg
			}
			time.Sleep(5 * time.Second)
			return
		}

		if hit {
			playerFrom.Out <- Message{Type: MessageTurn, Data: nil}
		} else {
			playerTo.Out <- Message{Type: MessageTurn, Data: nil}
		}
	}
}
