package main

import (
	"log"
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
		if err := LoadFont("assets/Wasted-Vindey.ttf"); err != nil {
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

				if dialogActive {
					drawDialogBox()
			}

        window.SwapBuffers()
        glfw.PollEvents()
    }
}
