package main

import "github.com/go-gl/glfw/v3.3/glfw"


func onMouseClick(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
			mx, my := w.GetCursorPos()

			// If dialog is active, intercept clicks for the dialog
			if dialogActive {
					ndcX, ndcY := screenToNDC(mx, my)
					// Check "Yes" button
					if pointInRect(ndcX, ndcY, dialogButtonYesRect) {
							if dialogOnNewGame != nil {
									dialogOnNewGame()
							}
							dialogActive = false
							return
					}
					// Check "No" button
					if pointInRect(ndcX, ndcY, dialogButtonNoRect) {
							if dailogOnCancel != nil {
								dailogOnCancel()
							}
							dialogActive = false
							return
					}
					// If clicked outside buttons, do nothing (stay in dialog).
					return
			}

			// If no dialog: check if we clicked Reset
			if isOverResetButton(mx, my) {
					resetGame()
					return
			}

			// Normal gameplay logic: placing goats, dragging tigers, etc.
			boardX, boardY := screenToBoardCoords(mx, my)
			if boardX == -1 || boardY == -1 {
					return
			}

			if turn == 1 {
					// Goat's turn
					onGoatPress(boardX, boardY)
			} else {
					// Tiger's turn
					onTigerPress(boardX, boardY)
			}
	}

	if button == glfw.MouseButtonLeft && action == glfw.Release {
			if draggingPiece {
					mx, my := w.GetCursorPos()
					boardX, boardY := screenToBoardCoords(mx, my)
					onPieceRelease(boardX, boardY)
			}
	}
}


//  updates the currentDragPos if dragging a piece
func onMouseMove(w *glfw.Window, xpos float64, ypos float64) {
	if draggingPiece {
			ndcX, ndcY := screenToNDC(xpos, ypos)
			currentDragPos[0] = ndcX
			currentDragPos[1] = ndcY
	}
}

// onKeyPress can handle ESC to close or other shortcuts
func onKeyPress(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
			if key == glfw.KeyEscape {
					w.SetShouldClose(true)
			}
	}
}

// Helper: check if a point is inside a rect in NDC
func pointInRect(x, y float32, r struct{x1,y1,x2,y2 float32}) bool {
	return x >= r.x1 && x <= r.x2 && y <= r.y1 && y >= r.y2
}

