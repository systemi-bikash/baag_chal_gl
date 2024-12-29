package main

import (
	"log"
	"math"

	"github.com/go-gl/gl/v2.1/gl"
)

// resetGame re-initializes the entire board, placing tigers in the corners, etc.
func resetGame() {
  // Clear board
  for i := 0; i < 5; i++ {
      for j := 0; j < 5; j++ {
          boardState[i][j] = 0
      }
  }
  // Place tigers in corners
  boardState[0][0] = 2
  boardState[0][4] = 2
  boardState[4][0] = 2
  boardState[4][4] = 2

  turn = 1
  placedGoats = 0
  capturedGoats = 0
  draggingPiece = false
  selectedPiece = [2]int{-1, -1}
  currentDragPos = [2]float32{0.0, 0.0}
	gameOver = false

  log.Println("Game reset.")
}

// isOverResetButton checks if the mouse click in window coordinates is
// inside the bounding box for the Reset button (in normalized coords).
func isOverResetButton(mouseX, mouseY float64) bool {
  ndcX, ndcY := screenToNDC(mouseX, mouseY)
  return ndcX >= resetButtonRect.minX && ndcX <= resetButtonRect.maxX &&
        ndcY >= resetButtonRect.minY && ndcY <= resetButtonRect.maxY
}

// drawBoard renders the 5x5 grid plus diagonal lines.
func drawBoard() {
	if bgAverageBrightness > 0.5 {
		// background is “light,” so use black lines
		gl.Color3f(0, 0, 0)
} else {
		// background is “dark,” so use white lines
		gl.Color3f(1, 1, 1)
}

    // 5 vertical + 5 horizontal lines
    for i := 0; i < 5; i++ {
        // Horizontal
        startX := float32(-0.8)
        startY := float32(-0.8 + 0.4*float32(i))
        endX := float32(0.8)
        endY := startY
        gl.Begin(gl.LINES)
        gl.Vertex2f(startX, startY)
        gl.Vertex2f(endX, endY)
        gl.End()

        // Vertical
        startX = float32(-0.8 + 0.4*float32(i))
        startY = float32(-0.8)
        endX = startX
        endY = float32(0.8)
        gl.Begin(gl.LINES)
        gl.Vertex2f(startX, startY)
        gl.Vertex2f(endX, endY)
        gl.End()
    }

    // Some diagonals (traditional Baag-Chal has specific diagonals):
    gl.Begin(gl.LINES)
    // Main diagonal
    gl.Vertex2f(-0.8, 0.8)
    gl.Vertex2f(0.8, -0.8)
    // Opposite diagonal
    gl.Vertex2f(-0.8, -0.8)
    gl.Vertex2f(0.8, 0.8)
    // Extra diagonals to center points
    gl.Vertex2f(0.0, 0.8)
    gl.Vertex2f(-0.8, 0.0)
    gl.Vertex2f(0.0, 0.8)
    gl.Vertex2f(0.8, 0.0)
    gl.Vertex2f(0.0, -0.8)
    gl.Vertex2f(-0.8, 0.0)
    gl.Vertex2f(0.0, -0.8)
    gl.Vertex2f(0.8, 0.0)
    gl.End()
}

func drawPieceTex(x, y float32, tex uint32, size float32) {
  gl.BindTexture(gl.TEXTURE_2D, tex)

  half := size * 0.5

  gl.Begin(gl.QUADS)

  // Top-left
  gl.TexCoord2f(0, 0) // Flip vertically
  gl.Vertex2f(x-half, y+half)

  // Top-right
  gl.TexCoord2f(1, 0) // Flip vertically
  gl.Vertex2f(x+half, y+half)

  // Bottom-right
  gl.TexCoord2f(1, 1) // Flip vertically
  gl.Vertex2f(x+half, y-half)

  // Bottom-left
  gl.TexCoord2f(0, 1) // Flip vertically
  gl.Vertex2f(x-half, y-half)
  gl.End()

  gl.BindTexture(gl.TEXTURE_2D, 0)
}

// drawPieces renders goats and tigers on the board.
func drawPieces() {
    for i := 0; i < 5; i++ {
        for j := 0; j < 5; j++ {
            piece := boardState[i][j]
            if piece == 0 {
                continue
            }
            if draggingPiece && selectedPiece == [2]int{i, j} {
                // Skip the piece that is currently being dragged
                continue
            }
            x := boardPosX(i)
            y := boardPosY(j)

            switch piece {
            case 1: // goat
                drawPieceTex(x, y, goatTex, 0.12)
            case 2: // tiger
                drawPieceTex(x, y, tigerTex, 0.15)
            }
        }
    }
}

// drawDraggedPiece draws the piece under the mouse cursor
func drawDraggedPiece() {
    if turn == 1 {
        // gl.Color3f(0.0, 1.0, 0.0) // Goat color
        // drawCircle(currentDragPos[0], currentDragPos[1], goatRadius, 20)
        drawPieceTex(currentDragPos[0], currentDragPos[1], goatTex, 0.12)
    } else {
        // gl.Color3f(1.0, 0.0, 0.0) // Tiger color
        // drawCircle(currentDragPos[0], currentDragPos[1], tigerRadius, 20)
        drawPieceTex(currentDragPos[0], currentDragPos[1], tigerTex, 0.15)
    }
}

// drawCircle draws a filled circle at (x, y).
func drawCircle(x, y, radius float32, segments int) {
    angleStep := float32(2.0*math.Pi) / float32(segments)
    gl.Begin(gl.POLYGON)
    for i := 0; i < segments; i++ {
        theta := angleStep * float32(i)
        px := x + radius*float32(math.Cos(float64(theta)))
        py := y + radius*float32(math.Sin(float64(theta)))
        gl.Vertex2f(px, py)
    }
    gl.End()
}


// screenToBoardCoords converts window coords to a board cell
func screenToBoardCoords(x, y float64) (int, int) {
	ndcX, ndcY := screenToNDC(x, y)
	// Each grid cell is 0.4 wide from -0.8..+0.8
	// Check which cell center is within ~0.2 of the click
	for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
					cx := boardPosX(i)
					cy := boardPosY(j)
					if math.Abs(float64(cx)-float64(ndcX)) < 0.2 &&
							math.Abs(float64(cy)-float64(ndcY)) < 0.2 {
							return i, j
					}
			}
	}
	return -1, -1
}

// screenToNDC converts window mouse coords to normalized device coords (-1..1)
func screenToNDC(x, y float64) (float32, float32) {
	ndcX := float32((x / windowWidth) * 2 - 1)
	ndcY := float32(1 - (y / windowHeight) * 2)
	return ndcX, ndcY
}

// boardPosX, boardPosY map board indices [0..4] to NDC [-0.8..+0.8]
func boardPosX(i int) float32 {
	return -0.8 + 0.4*float32(i)
}

func boardPosY(j int) float32 {
	return -0.8 + 0.4*float32(j)
}
