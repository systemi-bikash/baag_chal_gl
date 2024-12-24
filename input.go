package main

import "github.com/go-gl/glfw/v3.3/glfw"


func onMouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button != glfw.MouseButtonLeft {
			return
	}
	mouseX, mouseY := window.GetCursorPos()

	// Check if clicked on the Reset button
	if action == glfw.Press {
			if isOverResetButton(mouseX, mouseY) {
					resetGame()
					return
			}
	}

	// handle board clicks
	boardX, boardY := screenToBoardCoords(mouseX, mouseY)
	if boardX < 0 || boardY < 0 || boardX >= 5 || boardY >= 5 {
			return
	}

	switch action {
	case glfw.Press:
			if turn == 1 {
					// Goat's turn
					onGoatPress(boardX, boardY)
			} else {
					// Tiger's turn
					onTigerPress(boardX, boardY)
			}
	case glfw.Release:
			if draggingPiece {
					onPieceRelease(boardX, boardY)
			}
	}
}

//  updates the currentDragPos if dragging a piece
func onMouseMove(window *glfw.Window, xpos, ypos float64) {
	if draggingPiece {
			currentDragPos[0] = float32((xpos / windowWidth) * 2 - 1)
			currentDragPos[1] = float32(1 - (ypos / windowHeight) * 2)
	}
}

func onKeyPress(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
if action == glfw.Press {
		switch key {
		case glfw.KeyR: // Reset the game
				resetGame()
		}
}
}
