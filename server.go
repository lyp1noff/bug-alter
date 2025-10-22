package main

import (
	"errors"
	"io"
	"log"
	"net"
	"time"
)

func runServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}()

	log.Printf("Server is listening on %s\n", addr)

	for {
		log.Println("Waiting for players...")
		players := make([]*Player, 0, 2)

		for len(players) < 2 {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				continue
			}

			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.SetNoDelay(true)
			}

			player := &Player{
				Conn:  conn,
				Board: initBoard(DefaultBoardSize),
				Ships: make([]Ship, 0, len(DefaultShipLengths)),
				In:    make(chan Message, 8),
				Out:   make(chan Message, 8),
			}
			player.Ships = placeShips(player.Board, DefaultShipLengths)
			players = append(players, player)
			log.Printf("Player connected: %s", conn.RemoteAddr())
		}

		log.Println("Starting new game...")
		go runGame(players)
	}
}

func runGame(players []*Player) {
	disconnect := make(chan *Player, 1)

	for _, player := range players {
		go func(p *Player) {
			for {
				msg, err := readMessage(p.Conn)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
						log.Printf("Player %s disconnected gracefully", p.Conn.RemoteAddr())
					} else {
						log.Printf("Read error from %s: %v", p.Conn.RemoteAddr(), err)
					}
					disconnect <- p
					return
				}
				p.In <- msg
			}
		}(player)

		go func(p *Player) {
			for msg := range p.Out {
				if err := sendMessage(p.Conn, msg); err != nil {
					log.Printf("Write error to %s: %v", p.Conn.RemoteAddr(), err)
					return
				}
			}
		}(player)

		initMsg := Message{Type: MessageInit}
		if err := initMsg.EncodeData(InitData{Ships: player.Ships}); err == nil {
			player.Out <- initMsg
		}
	}

	time.Sleep(200 * time.Millisecond)
	players[0].Out <- Message{Type: MessageTurn}

	for {
		select {
		case msg := <-players[0].In:
			gameProcessor(players[0], players[1], msg)
		case msg := <-players[1].In:
			gameProcessor(players[1], players[0], msg)
		case p := <-disconnect:
			log.Printf("Player disconnected: %s", p.Conn.RemoteAddr())
			cleanupGame(players)
			log.Println("Game ended — returning to lobby.")
			return
		}
	}
}

func cleanupGame(players []*Player) {
	for _, p := range players {
		_ = p.Conn.Close()
	}
}

func gameProcessor(playerFrom, playerTo *Player, msg Message) {
	switch msg.Type {
	case MessageShot:
		var shot ShotData
		if err := msg.DecodeData(&shot); err != nil {
			return
		}

		shotResult := shoot(playerTo.Board, playerTo.Ships, shot.X, shot.Y)

		resultMsg := Message{Type: MessageResult}
		_ = resultMsg.EncodeData(ResultData{
			X:         shot.X,
			Y:         shot.Y,
			Hit:       shotResult.Hit,
			Destroyed: shotResult.Destroyed,
			SunkShip:  shotResult.SunkShipCoords,
		})
		playerFrom.Out <- resultMsg

		shotMsg := Message{Type: MessageShot}
		_ = shotMsg.EncodeData(ShotData{shot.X, shot.Y})
		playerTo.Out <- shotMsg

		if isGameOver(playerTo.Board) {
			for _, p := range []*Player{playerFrom, playerTo} {
				endMsg := Message{Type: MessageGameEnd}
				_ = endMsg.EncodeData(GameEndData{Winner: p == playerFrom})
				p.Out <- endMsg
			}
			time.Sleep(5 * time.Second)
			cleanupGame([]*Player{playerFrom, playerTo})
			log.Println("Game completed successfully — back to lobby.")
			return
		}

		if shotResult.Hit {
			playerFrom.Out <- Message{Type: MessageTurn}
		} else {
			playerTo.Out <- Message{Type: MessageTurn}
		}
	}
}
