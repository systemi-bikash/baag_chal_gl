package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
    windowWidth  = 1000
    windowHeight = 1000
    tigerRadius  = 0.03
    goatRadius   = 0.02
)


var resetButtonRect = struct {
  minX, minY, maxX, maxY float32
}{
  0.75, 0.85, 0.95, 0.95,
}

func init() {
    // Lock the main thread for OpenGL
    runtime.LockOSThread()
}

func main() {
    if err := glfw.Init(); err != nil {
        log.Fatalln("failed to initialize glfw:", err)
    }
    defer glfw.Terminate()

    window, err := glfw.CreateWindow(windowWidth, windowHeight, "Baag-Chal Board", nil, nil)
    if err != nil {
        log.Fatalln("failed to create window:", err)
    }
    window.MakeContextCurrent()

    if err := gl.Init(); err != nil {
        log.Fatalln("failed to initialize OpenGL:", err)
    }

    gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
    gl.ClearColor(0.0, 0.0, 0.0, 1.0)

    // Initialize the board with tigers in the corners
    boardState[0][0] = 2
    boardState[0][4] = 2
    boardState[4][0] = 2
    boardState[4][4] = 2

    // Enable blending & texturing
    gl.Enable(gl.BLEND)
    gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
    gl.Enable(gl.TEXTURE_2D)

    // Load textures
    goatTex, err = LoadTexture("goat_1.png")
    if err != nil {
        log.Fatalf("Failed to load goat.png: %v", err)
    }
    tigerTex, err = LoadTexture("tiger.png")
    if err != nil {
        log.Fatalf("Failed to load tiger.png: %v", err)
    }

    // Set callbacks
    window.SetMouseButtonCallback(onMouseClick)
    window.SetCursorPosCallback(onMouseMove)
    window.SetKeyCallback(onKeyPress)

    // Main loop
    for !window.ShouldClose() {
        gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
        drawBoard()
        drawPieces()
        if draggingPiece {
            drawDraggedPiece()
        }

        window.SwapBuffers()
        glfw.PollEvents()
    }
}


//====================================================================
// LOAD TEXTURE
//====================================================================

// LoadTexture loads a texture from a PNG file.
func LoadTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	return texture, nil
}


//====================================================================
// MOVEMENT / CAPTURE
//====================================================================

func isValidMove(from, to [2]int) bool {
	// Must be in range and empty
	if to[0] < 0 || to[1] < 0 || to[0] >= 5 || to[1] >= 5 {
			return false
	}
	if boardState[to[0]][to[1]] != 0 {
			return false
	}

	dx := to[0] - from[0]
	dy := to[1] - from[1]

	// 1-step adjacency in any of 8 directions:
	// (|dx| == 1 && dy == 0)   => horizontal
	// (dx == 0 && |dy| == 1)   => vertical
	// (|dx| == 1 && |dy| == 1) => diagonal
	if (abs(dx) == 1 && dy == 0) ||
		(dx == 0 && abs(dy) == 1) ||
		(abs(dx) == 1 && abs(dy) == 1) {
			return true
	}

	// 2-step jump in any of 8 directions (possible capture):
	// (|dx| == 2 && dy == 0)   => horizontal jump
	// (dx == 0 && |dy| == 2)   => vertical jump
	// (|dx| == 2 && |dy| == 2) => diagonal jump
	if (abs(dx) == 2 && dy == 0) ||
		(dx == 0 && abs(dy) == 2) ||
		(abs(dx) == 2 && abs(dy) == 2) {
			return true
	}

	return false
}




//====================================================================
// RENDERING
//====================================================================


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
    gl.LineWidth(2.0)
    gl.Color3f(1.0, 1.0, 1.0) // White lines

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

//====================================================================
// COORDINATE TRANSFORMS / HELPER
//====================================================================

// screenToBoardCoords converts window coords to which board cell was clicked.
func screenToBoardCoords(x, y float64) (int, int) {
  ndcX, ndcY := screenToNDC(x, y)
  // each grid cell is 0.4 wide from -0.8..+0.8
  // check which cell center is within ~0.2 of the click
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

// screenToNDC converts window mouse coords to normalized device coords (-1..1).
func screenToNDC(x, y float64) (float32, float32) {
  ndcX := float32((x / windowWidth) * 2 - 1)
  ndcY := float32(1 - (y / windowHeight) * 2)
  return ndcX, ndcY
}

func boardPosX(i int) float32 {
  return -0.8 + 0.4*float32(i)
}

func boardPosY(j int) float32 {
  return -0.8 + 0.4*float32(j)
}
