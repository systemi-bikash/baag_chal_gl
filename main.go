package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
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

			// loading font
		if err := LoadFont("assets/Copenhagen.ttf"); err != nil {
    		log.Fatalln("Failed to load font:", err)
		}
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
        drawUI()
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
	// Ensure destination is within range
	if to[0] < 0 || to[1] < 0 || to[0] >= 5 || to[1] >= 5 {
			return false
	}

	// Ensure the destination is connected to the source
	validMoves, exists := validConnections[from]
	if !exists {
			return false
	}
	isConnected := false
	for _, conn := range validMoves {
			if conn == to {
					isConnected = true
					break
			}
	}
	if !isConnected {
			return false
	}

	// Check adjacency or valid jump (already handled)
	dx := to[0] - from[0]
	dy := to[1] - from[1]

	// Single step (adjacency)
	if (abs(dx) == 1 && dy == 0) ||
			(dx == 0 && abs(dy) == 1) ||
			(abs(dx) == 1 && abs(dy) == 1) {
			return true
	}

	// Two-step jump with a goat in the middle
	if (abs(dx) == 2 && dy == 0) ||
			(dx == 0 && abs(dy) == 2) ||
			(abs(dx) == 2 && abs(dy) == 2) {
			midX := (from[0] + to[0]) / 2
			midY := (from[1] + to[1]) / 2
			if boardState[midX][midY] == 1 { // Must be a goat in the middle
					return true
			}
	}

	return false
}

