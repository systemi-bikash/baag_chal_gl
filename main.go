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
    windowWidth  = 800
    windowHeight = 800
    tigerRadius  = 0.03
    goatRadius   = 0.02
)

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


var (
  tigerTex uint32
  goatTex  uint32
)


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
    goatTex, err = LoadTexture("goat.png")
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
// CALLBACKS
//====================================================================

//  handles placing goats, starting drags, and finalizing moves.
func onMouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
    if button != glfw.MouseButtonLeft {
        return
    }
    mouseX, mouseY := window.GetCursorPos()
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
// TURN LOGIC / HELPER FUNCTIONS
//====================================================================

// onGoatPress handles placing a goat (if <20 placed) or
// potentially dragging goats if you want goats to move after 20 are placed.
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

// switchTurn toggles the turn: 1 -> 2, 2 -> 1
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
// canCapture checks if there is a goat in the midpoint between 'from' and 'to'
// when jumping 2 steps in any of the 8 directions.
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

// Utility function for absolute value of an int
func abs(n int) int {
	if n < 0 {
			return -n
	}
	return n
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

//====================================================================
// RENDERING
//====================================================================

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
// COORDINATE TRANSFORMS
//====================================================================

// screenToBoardCoords converts window coordinates to a rough board index.
func screenToBoardCoords(x, y float64) (int, int) {
    // Convert [0..width] to [-1..+1], [0..height] to [+1..-1]
    nx := float32((x / windowWidth)*2 - 1)
    ny := float32(1 - (y / windowHeight)*2)

    // each grid cell is 0.4 units in each direction from -0.8..+0.8
    // simple bounding check for which cell the user clicked
    for i := 0; i < 5; i++ {
        for j := 0; j < 5; j++ {
            cx := boardPosX(i)
            cy := boardPosY(j)
            if math.Abs(float64(cx)-float64(nx)) < 0.2 &&
              math.Abs(float64(cy)-float64(ny)) < 0.2 {
                return i, j
            }
        }
    }
    return -1, -1
}

// boardPosX, boardPosY map board indices [0..4] to normalized device coords [-0.8..+0.8].
func boardPosX(i int) float32 {
    return -0.8 + 0.4*float32(i)
}

func boardPosY(j int) float32 {
    return -0.8 + 0.4*float32(j)
}
