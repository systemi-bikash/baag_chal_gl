package main

import (
	"fmt"
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

var  bgTex uint32

var bgAverageBrightness float32


// vertex positions and texture coordinates for a full-screen quad.
var (
	bgShaderProgram uint32
	VAO            uint32
	VBO            uint32
	// 6 vertices * 4 floats each => x,y,u,v
	quadVertices = []float32{
    // x,    y,    u,   v
    -1.0,  1.0,   0.0, 0.0,
    -1.0, -1.0,   0.0, 1.0,
     1.0, -1.0,   1.0, 1.0,

    -1.0,  1.0,   0.0, 0.0,
     1.0, -1.0,   1.0, 1.0,
     1.0,  1.0,   1.0, 0.0,
}
)

var vertexShaderSource = `
#version 120
attribute vec2 aPos;
attribute vec2 aTexCoord;

varying vec2 vTexCoord;

void main() {
    gl_Position = vec4(aPos, 0.0, 1.0);
    vTexCoord = aTexCoord;
}
` + "\x00"


var fragmentShaderSource = `
#version 120
varying vec2 vTexCoord;
uniform sampler2D uTexture;

void main() {
    vec4 color = texture2D(uTexture, vTexCoord);

    // Per-pixel brightness
    float brightness = dot(color.rgb, vec3(0.2126, 0.7152, 0.0722));

    if (brightness > 0.5) {
        // lighten
        gl_FragColor = vec4(color.rgb + 0.2, color.a);
    } else {
        // darken
        gl_FragColor = vec4(color.rgb * 0.5, color.a);
    }
}
` + "\x00"


func init() {
    // Lock the main thread for OpenGL
    runtime.LockOSThread()
}


func main() {
	// 1) Init GLFW
	if err := glfw.Init(); err != nil {
			log.Fatalln("Failed to init GLFW:", err)
	}
	defer glfw.Terminate()

	// If you want a 2.1 context (for immediate mode compatibility on macOS):
	//   Remove the hints for 3.3:
	// glfw.WindowHint(glfw.ContextVersionMajor, 2)
	// glfw.WindowHint(glfw.ContextVersionMinor, 1)
	// OR if on Windows, you might do 3.3 + a compatibility profile if you want modern + old calls.

	glfw.WindowHint(glfw.Resizable, glfw.False) // or let it be resizable

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "GoGL Demo", nil, nil)
	if err != nil {
			log.Fatalln("Failed to create window:", err)
	}
	window.MakeContextCurrent()

	// 2) Init GL
	if err := gl.Init(); err != nil {
			log.Fatalln("Failed to init GL:", err)
	}

	versionStr := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version:", versionStr)

	gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	// 3) Build the background shader
	bgShaderProgram = createShaderProgram(vertexShaderSource, fragmentShaderSource)

	// 4) Create the VAO/VBO for the full-screen quad
	{
			var vbo uint32
			gl.GenBuffers(1, &vbo)
			VBO = vbo
			gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
			gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

			var vao uint32
			gl.GenVertexArrays(1, &vao)
			VAO = vao

			gl.BindVertexArray(VAO)

			// aPos => location=0 in the shader
			gl.EnableVertexAttribArray(0)
			gl.VertexAttribPointer(
					0, 2, gl.FLOAT, false,
					4*4, gl.PtrOffset(0),
			)
			// aTexCoord => location=1
			gl.EnableVertexAttribArray(1)
			gl.VertexAttribPointer(
					1, 2, gl.FLOAT, false,
					4*4, gl.PtrOffset(2*4),
			)

			gl.BindVertexArray(0)
	}

	// 5) Load background image
	var err2 error
	bgTex, err2 = loadTexture("assets/mountain.png")
	if err2 != nil {
			log.Fatalf("Failed to load background: %v", err2)
	}

	// **Compute average brightness** so we know if the wallpaper is bright or dark
	bgAverageBrightness = computeAverageBrightness("test.png")

	// 6) Initialize board
	boardState[0][0] = 2
	boardState[0][4] = 2
	boardState[4][0] = 2
	boardState[4][4] = 2

	// 7) Load goat/tiger
	goat, err := loadTexture("assets/goat_1.png")
	if err != nil {
    log.Fatalf("Failed to load goat texture: %v", err)
}
	tiger, err := loadTexture("assets/tiger.png")
	if err != nil {
    log.Fatalf("Failed to load tiger texture: %v", err)
}
	goatTex, tigerTex = goat, tiger

	// Setup callbacks (optional)
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if action == glfw.Press && key == glfw.KeyEscape {
					w.SetShouldClose(true)
			}
	})

	// main loop
	for !window.ShouldClose() {
			// Clear once
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			// ----- 1) Draw the background with our “detect color” fragment shader
			gl.UseProgram(bgShaderProgram)
			gl.BindVertexArray(VAO)

			// Bind the background texture
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, bgTex)

			// The uniform sampler2D location
			loc := gl.GetUniformLocation(bgShaderProgram, gl.Str("uTexture\x00"))
			gl.Uniform1i(loc, 0) // use TEXTURE0

			// Draw 6 vertices
			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			gl.BindVertexArray(0)
			gl.UseProgram(0)

			// ----- 2) Draw the board lines, pieces, UI, etc.
			drawBoard()
			drawPieces()
			drawUI()
			if draggingPiece {
					drawDraggedPiece()
			}
			if dialogActive {
					drawDialogBox()
			}

			// Swap buffers & poll
			window.SwapBuffers()
			glfw.PollEvents()
	}
	// Cleanup
	gl.DeleteProgram(bgShaderProgram)
	gl.DeleteBuffers(1, &VBO)
	gl.DeleteVertexArrays(1, &VAO)
}


// createShaderProgram compiles the vertex & fragment shaders, then links them into a program.
func createShaderProgram(vsSource, fsSource string) uint32 {
	vertexShader := compileShader(vsSource, gl.VERTEX_SHADER)
	fragmentShader := compileShader(fsSource, gl.FRAGMENT_SHADER)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// check for link errors
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
			var logLength int32
			gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

			infoLog := make([]byte, logLength+1)
			gl.GetProgramInfoLog(program, logLength, nil, &infoLog[0])
			log.Fatalf("Failed to link program: %s\n", string(infoLog))
	}

	// shaders can be deleted after linking
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}


func compileShader(src string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	cSources, freeFn := gl.Strs(src)
	gl.ShaderSource(shader, 1, cSources, nil)
	freeFn()
	gl.CompileShader(shader)

	// check compile errors
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
			var logLength int32
			gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

			infoLog := make([]byte, logLength+1)
			gl.GetShaderInfoLog(shader, logLength, nil, &infoLog[0])
			shaderTypeStr := "UNKNOWN"
			if shaderType == gl.VERTEX_SHADER {
					shaderTypeStr = "VERTEX"
			} else if shaderType == gl.FRAGMENT_SHADER {
					shaderTypeStr = "FRAGMENT"
			}
			log.Fatalf("Failed to compile %s shader:\n%s\n", shaderTypeStr, string(infoLog))
	}
	return shader
}

// loadTexture loads a PNG file into an OpenGL texture.
func loadTexture(filepath string) (uint32, error) {
	f, err := os.Open(filepath)
	if err != nil {
			return 0, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
			return 0, err
	}

	// convert to RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	// set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// load data
	width := int32(rgba.Bounds().Dx())
	height := int32(rgba.Bounds().Dy())
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0,
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.BindTexture(gl.TEXTURE_2D, 0)
	return tex, nil
}

func processInput(window *glfw.Window) {
	if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
	}
}

// computeAverageBrightness loads the same PNG, loops its pixels, returns average brightness in [0..1]
func computeAverageBrightness(path string) float32 {
	f, err := os.Open(path)
	if err != nil {
			log.Printf("Cannot open for brightness calc: %v", err)
			return 0.5
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
			log.Printf("Cannot decode for brightness calc: %v", err)
			return 0.5
	}

	bounds := img.Bounds()
	var sum float64
	var count float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					// RGBA from Go is 16-bit per channel (0..65535)
					rr := float64(r) / 65535.0
					gg := float64(g) / 65535.0
					bb := float64(b) / 65535.0
					// Luma:
					lum := rr*0.2126 + gg*0.7152 + bb*0.0722
					sum += lum
					count++
			}
	}
	avg := float32(sum / count) // ~ 0..1
	return avg
}
