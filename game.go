package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

const (
	BoardSize          = 5
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

func redraw() {
	fmt.Print("\033[H\033[J")
}

//func printLine(size int) {
//	for range size {
//		fmt.Print("-")
//	}
//	fmt.Println()
//}
//
//func printBoard(board [][]byte) {
//	fmt.Print("  ")
//	for x := range board[0] {
//		fmt.Printf("%d ", x)
//	}
//	fmt.Println()
//	for i, row := range board {
//		fmt.Printf("%d ", i)
//		for _, cell := range row {
//			fmt.Printf("%c ", cell)
//		}
//		fmt.Println()
//	}
//}

func printBoards(own, enemy [][]byte) {
	size := len(own)

	//fmt.Println()
	fmt.Println("Your board       Enemy board")
	fmt.Print("   ")
	for x := 0; x < size; x++ {
		fmt.Printf("%c ", 'A'+x)
	}
	fmt.Print("       ")
	for x := 0; x < size; x++ {
		fmt.Printf("%c ", 'A'+x)
	}
	fmt.Println()

	for y := 0; y < size; y++ {
		fmt.Printf("%2d ", y+1)
		for _, cell := range own[y] {
			fmt.Printf("%c ", cell)
		}
		fmt.Print("    ")
		fmt.Printf("%2d ", y+1)
		for _, cell := range enemy[y] {
			fmt.Printf("%c ", cell)
		}
		fmt.Println()
	}
	fmt.Println()
}

func placeWarships(board [][]byte, enemyWarshipsCount int) [][]int {
	warshipsLocation := make([][]int, enemyWarshipsCount)

	for i := 0; i < enemyWarshipsCount; {
		y := rand.IntN(len(board))
		x := rand.IntN(len(board[0]))
		if board[y][x] != Ship {
			warshipsLocation[i] = []int{x, y}
			board[y][x] = Ship
			i++
		}
	}

	return warshipsLocation
}

func shoot(board [][]byte, x, y int) bool {
	if board[y][x] == Ship {
		board[y][x] = Hit
		return true
	}
	board[y][x] = Miss

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
	var x, y int
	for {
		fmt.Print("Enter coordinates (e.g. A1, 1A, A 1): ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Input error:", err)
			}
			continue
		}

		input := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		input = strings.ReplaceAll(input, " ", "")

		if len(input) < 2 {
			fmt.Println("Please enter a letter and a number.")
			continue
		}

		var letter rune
		var numberStr string

		for _, ch := range input {
			if ch >= 'A' && ch <= 'Z' {
				letter = ch
			} else if ch >= '0' && ch <= '9' {
				numberStr += string(ch)
			}
		}

		if letter == 0 || numberStr == "" {
			fmt.Println("Invalid format. Example: A1 or 1A")
			continue
		}

		yInt, err := strconv.Atoi(numberStr)
		if err != nil {
			fmt.Println("Invalid number.")
			continue
		}

		x = int(letter - 'A')
		y = yInt - 1 // т.к. отображаем от 1, а индексация с 0

		if x < 0 || x >= BoardSize || y < 0 || y >= BoardSize {
			fmt.Println("Coordinates out of range.")
			continue
		}

		if board[y][x] != Empty {
			fmt.Println("You have already shot this cell.")
			continue
		}

		return x, y
	}
}
