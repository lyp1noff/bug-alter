package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/inancgumus/screen"
)

var (
	DefaultBoardSize   = 10
	DefaultShipLengths = []int{5, 4, 3, 3, 2, 1}
)

const (
	CellEmpty = '.'
	CellShip  = 'S'
	CellHit   = 'X'
	CellMiss  = 'o'
)

type Ship struct {
	Coords [][2]int
	Sunk   bool
}

type ShotResult struct {
	Hit            bool
	Destroyed      bool
	SunkShipCoords [][2]int
}

func initBoard(size int) [][]byte {
	board := make([][]byte, size)
	for y := range board {
		board[y] = make([]byte, size)
		for x := range board[y] {
			board[y][x] = CellEmpty
		}
	}

	return board
}

func redraw() {
	screen.Clear()
	screen.MoveTopLeft()
}

func printBoards(own, enemy [][]byte) {
	size := len(own)

	boardWidth := 3 + size*2
	spaceBetween := strings.Repeat(" ", boardWidth-len("Your board")+4)
	fmt.Printf("   Your board%sEnemy board\n", spaceBetween)

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
			printCell(cell)
		}
		fmt.Print("    ")
		fmt.Printf("%2d ", y+1)
		for _, cell := range enemy[y] {
			printCell(cell)
		}
		fmt.Println()
	}
	fmt.Println()
}

func printCell(c byte) {
	switch c {
	case CellShip:
		_, _ = color.New(color.FgBlue).Printf("%c ", c)
	case CellHit:
		_, _ = color.New(color.FgRed, color.Bold).Printf("%c ", c)
	case CellMiss:
		_, _ = color.New(color.FgHiBlack).Printf("%c ", c)
	default:
		fmt.Printf("%c ", c)
	}
}

func placeShips(board [][]byte, shipLengths []int) []Ship {
	size := len(board)
	ships := make([]Ship, 0, len(shipLengths))

	for _, length := range shipLengths {
		for {
			x := rand.IntN(size)
			y := rand.IntN(size)
			horizontal := rand.IntN(2) == 0

			if canPlace(board, x, y, length, horizontal) {
				coords := make([][2]int, 0, length)
				for i := 0; i < length; i++ {
					nx, ny := x, y
					if horizontal {
						nx += i
					} else {
						ny += i
					}
					board[ny][nx] = CellShip
					coords = append(coords, [2]int{nx, ny})
				}
				ships = append(ships, Ship{Coords: coords})
				break
			}
		}
	}
	return ships
}

func canPlace(board [][]byte, x, y, length int, horizontal bool) bool {
	size := len(board)
	for i := 0; i < length; i++ {
		nx, ny := x, y
		if horizontal {
			nx += i
		} else {
			ny += i
		}
		if nx < 0 || ny < 0 || nx >= size || ny >= size {
			return false
		}
		if board[ny][nx] != CellEmpty {
			return false
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				xx, yy := nx+dx, ny+dy
				if xx >= 0 && yy >= 0 && xx < size && yy < size && board[yy][xx] == CellShip {
					return false
				}
			}
		}
	}
	return true
}

func shoot(board [][]byte, ships []Ship, x, y int) ShotResult {
	if board[y][x] == CellShip {
		board[y][x] = CellHit

		for i := range ships {
			for _, c := range ships[i].Coords {
				if c[0] == x && c[1] == y {
					if isSunk(board, ships[i].Coords) {
						ships[i].Sunk = true
						markAroundDestroyed(board, ships[i].Coords)
						return ShotResult{
							Hit:            true,
							Destroyed:      true,
							SunkShipCoords: ships[i].Coords,
						}
					}
					return ShotResult{Hit: true}
				}
			}
		}
	}
	if board[y][x] == CellEmpty {
		board[y][x] = CellMiss
	}
	return ShotResult{}
}

func isSunk(board [][]byte, coords [][2]int) bool {
	for _, c := range coords {
		if board[c[1]][c[0]] != CellHit {
			return false
		}
	}
	return true
}

func markAroundDestroyed(board [][]byte, coords [][2]int) {
	size := len(board)
	for _, c := range coords {
		x, y := c[0], c[1]
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				xx, yy := x+dx, y+dy
				if xx >= 0 && yy >= 0 && xx < size && yy < size {
					if board[yy][xx] == CellEmpty {
						board[yy][xx] = CellMiss
					}
				}
			}
		}
	}
}

func isGameOver(board [][]byte) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == CellShip {
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
		y = yInt - 1

		if x < 0 || x >= DefaultBoardSize || y < 0 || y >= DefaultBoardSize {
			fmt.Println("Coordinates out of range.")
			continue
		}

		if board[y][x] != CellEmpty {
			fmt.Println("You have already shot this cell.")
			continue
		}

		return x, y
	}
}
