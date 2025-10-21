// package main
package legacy

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

const (
	Size               = 5
	EnemyWarshipsCount = 3
)

const (
	Empty = '.'
	Ship  = 'S'
	Hit   = 'X'
	Miss  = 'o'
)

func initBoard(size int) [][]byte {
	board := make([][]byte, size)
	for y := range board {
		board[y] = make([]byte, size)
		for x := range board[y] {
			board[y][x] = Empty
		}
	}

	return board
}

func printLine(size int) {
	for range size {
		fmt.Print("-")
	}
	fmt.Println()
}

func printBoard(board [][]byte) {
	fmt.Print("  ")
	for x := range board[0] {
		fmt.Printf("%d ", x)
	}
	fmt.Println()
	for i, row := range board {
		fmt.Printf("%d ", i)
		for _, cell := range row {
			fmt.Printf("%c ", cell)
		}
		fmt.Println()
	}
}

func placeWarships(board [][]byte, enemyWarshipsCount int) {
	for i := 0; i < enemyWarshipsCount; {
		y := rand.IntN(len(board))
		x := rand.IntN(len(board[0]))
		if board[y][x] != Ship {
			board[y][x] = Ship
			i++
		}
	}
}

func shoot(board, visibleBoard [][]byte, x, y int) bool {
	if board[y][x] == Ship {
		board[y][x] = Hit
		visibleBoard[y][x] = Hit
		return true
	}
	board[y][x] = Miss
	visibleBoard[y][x] = Miss

	return false
}

func isGameOver(board [][]byte) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == Ship {
				return false
			}
		}
	}

	return true
}

func readCoordsAndValidate(scanner *bufio.Scanner, board [][]byte) (int, int) {
	var inputX, inputY int
	for {
		fmt.Print("Enter coordinates (X Y): ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Input error:", err)
			}
			continue
		}

		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			fmt.Println("Please enter two numbers separated by a space.")
			continue
		}

		x, err1 := strconv.Atoi(fields[0])
		y, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			fmt.Println("Coordinates must be integers.")
			continue
		}

		if x < 0 || y < 0 || x >= Size || y >= Size {
			fmt.Println("Coordinates out of range.")
			continue
		}

		if board[y][x] != Empty && board[y][x] != Ship {
			fmt.Println("You have already shot this cell.")
			continue
		}

		inputX, inputY = x, y
		break
	}

	return inputX, inputY
}

func legacy() {
	board := initBoard(Size)
	visibleBoard := initBoard(Size)
	scanner := bufio.NewScanner(os.Stdin)

	placeWarships(board, EnemyWarshipsCount)

	//printBoard(board)

	fmt.Println("Game!")

	for {
		printLine(Size*2 + 2)
		printBoard(visibleBoard)

		x, y := readCoordsAndValidate(scanner, board)
		if shoot(board, visibleBoard, x, y) {
			fmt.Println("HIT!")
		} else {
			fmt.Println("MISS!")
		}

		if isGameOver(board) {
			printBoard(visibleBoard)
			fmt.Println("You WON!")
			break
		}
	}
}
