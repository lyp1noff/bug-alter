package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "client", "server or client")
	addr := flag.String("addr", "localhost:9000", "server address")
	flag.Usage = func() {
		fmt.Println("Usage: battleship -mode=server|client [-addr=host:port]")
	}
	flag.Parse()

	switch *mode {
	case "server":
		runServer(*addr)
	case "client":
		runClient(*addr)
	default:
		fmt.Println("Unknown mode:", *mode)
		os.Exit(1)
	}
}
