package main

import "log"

var (
	// 0 = empty, 1 = goat, 2 = tiger
	boardState [5][5]int

	// turn: 1 = goat, 2 = tiger
	turn = 1

	placedGoats   = 0
	capturedGoats = 0

	// Dragging state
	draggingPiece  bool
	selectedPiece  = [2]int{-1, -1}
	currentDragPos = [2]float32{0.0, 0.0}
)

const maxGoats = 20

var (
  tigerTex uint32
  goatTex  uint32
)


var validConnections = map[[2]int][][2]int{
	{0, 0}: {{0, 1}, {1, 1}, {1, 0}},
	{0, 1}: {{0, 0}, {0, 2}, {1, 1}},
	{0, 2}: {{0, 1}, {0, 3}, {1, 1}, {1, 2}},
	{0, 3}: {{0, 2}, {0, 4}, {1, 3}},
	{0, 4}: {{0, 3}, {1, 3}, {1, 4}},

	{1, 0}: {{0, 0}, {1, 1}, {2, 0}},
	{1, 1}: {{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 2}, {2, 0}, {2, 2}, {2, 1}},
	{1, 2}: {{0, 2}, {1, 1}, {1, 3}, {2, 2}},
	{1, 3}: {{0, 3}, {0, 4}, {1, 2}, {1, 4}, {2, 4}, {2, 2}, {2, 3}},
	{1, 4}: {{0, 4}, {1, 3}, {2, 4}},

	{2, 0}: {{1, 0}, {2, 1}, {3, 0}},
	{2, 1}: {{1, 1}, {2, 0}, {2, 2}, {3, 1}},
	{2, 2}: {{1, 1}, {1, 2}, {1, 3}, {2, 1}, {2, 3}, {3, 1}, {3, 3}, {3, 2}},
	{2, 3}: {{1, 3}, {2, 2}, {2, 4}, {3, 3}},
	{2, 4}: {{1, 4}, {2, 3}, {3, 4}},

	{3, 0}: {{2, 0}, {3, 1}, {4, 0}},
	{3, 1}: {{2, 0}, {2, 1}, {2, 2}, {3, 0}, {3, 2}, {4, 0}, {4, 2}, {4, 1}},
	{3, 2}: {{2, 2}, {3, 1}, {3, 3}, {4, 2}},
	{3, 3}: {{2, 3}, {2, 4}, {3, 2}, {3, 4}, {4, 4}, {4, 2}, {4, 3}},
	{3, 4}: {{2, 4}, {3, 3}, {4, 4}},

	{4, 0}: {{3, 0}, {4, 1}, {4, 2}},
	{4, 1}: {{3, 1}, {4, 0}, {4, 2}},
	{4, 2}: {{3, 1}, {3, 2}, {3, 3}, {4, 1}, {4, 3}},
	{4, 3}: {{3, 3}, {3, 4}, {4, 2}, {4, 4}},
	{4, 4}: {{3, 4}, {4, 3}},
}

func switchTurn() {
	turn = 3 - turn
	log.Printf("Turn switched to %d", turn)
}

// placeGoat puts a goat on the board
func placeGoat(x, y int) {
	boardState[x][y] = 1
	placedGoats++
	log.Printf("Goat placed at (%d, %d). Total placed: %d", x, y, placedGoats)
}


//  handles finalizing a move for the piece being dragged.
func onPieceRelease(boardX, boardY int) {
	draggingPiece = false
	from := selectedPiece
	to := [2]int{boardX, boardY}

	if isValidMove(from, to) {
			if canCapture(from, to) {
					captureGoat(from, to)
			} else {
					// Normal move
					boardState[from[0]][from[1]] = 0
					boardState[to[0]][to[1]] = turn
			}
			switchTurn()
	} else {
			log.Printf("Invalid move from (%d, %d) to (%d, %d)", from[0], from[1], to[0], to[1])
	}

	selectedPiece = [2]int{-1, -1}
}


// canCapture checks if there's a goat in the midpoint when jumping 2 steps
func canCapture(from, to [2]int) bool {
	dx := to[0] - from[0]
	dy := to[1] - from[1]

	// Only a 2-step jump can capture
	if !((abs(dx) == 2 && dy == 0) ||
			(dx == 0 && abs(dy) == 2) ||
			(abs(dx) == 2 && abs(dy) == 2)) {
			return false
	}
	midX := (from[0] + to[0]) / 2
	midY := (from[1] + to[1]) / 2

	// Check if the midpoint is a goat
	return boardState[midX][midY] == 1
}


// captureGoat removes a goat from the midpoint and moves the tiger
func captureGoat(from, to [2]int) {
	midX := (from[0] + to[0]) / 2
	midY := (from[1] + to[1]) / 2
	boardState[midX][midY] = 0     // Remove the goat
	boardState[from[0]][from[1]] = 0
	boardState[to[0]][to[1]] = 2   // Move the tiger
	capturedGoats++
	log.Printf("Goat captured! Total captured: %d", capturedGoats)
}


// Utility for integer absolute value
func abs(n int) int {
	if n < 0 {
			return -n
	}
	return n
}

func onGoatPress(boardX, boardY int) {
	// 1) If fewer than 20 goats have been placed, place a new goat
	if placedGoats < 20 {
			if boardState[boardX][boardY] == 0 {
					placeGoat(boardX, boardY)

					// If the goats and tigers to alternate immediately:
					switchTurn()
			}
			return
	}
}

//  handles initiating a drag on a tiger piece
func onTigerPress(boardX, boardY int) {
	if boardState[boardX][boardY] == 2 {
			draggingPiece = true
			selectedPiece = [2]int{boardX, boardY}
			log.Printf("Tiger selected at (%d, %d)", boardX, boardY)
	}
}

