package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/golang/freetype"
)


var resetButtonRect = struct {
	minX, minY, maxX, maxY float32
}{
	0.75, 0.85, 0.95, 0.95,
}

func drawUI() {
	// 1) Draw the reset button (simple rectangle) near the top-right
	gl.Color3f(0.2, 0.2, 0.2) // dark gray
	gl.Begin(gl.QUADS)
	gl.Vertex2f(resetButtonRect.minX, resetButtonRect.minY)
	gl.Vertex2f(resetButtonRect.maxX, resetButtonRect.minY)
	gl.Vertex2f(resetButtonRect.maxX, resetButtonRect.maxY)
	gl.Vertex2f(resetButtonRect.minX, resetButtonRect.maxY)
	gl.End()

	// If you have a text rendering approach, you'd label the button:
	drawText2D(resetButtonRect.minX+0.01, resetButtonRect.minY+0.01, "RESET")

	// 2) Draw a banner for goat stats in the top-left corner
	goatsRemaining := maxGoats - placedGoats
	banner := fmt.Sprintf("Goats Placed: %d | Captured: %d | Remaining: %d",
			placedGoats, capturedGoats, goatsRemaining)

	drawText2D(-0.6, .90, banner)
}


func drawText2D(x, y float32, text string) {
	if text == "" || mainFont == nil {
			return
	}

	// Create a small RGBA image buffer

	imgWidth, imgHeight := 256, 64
	rgba := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// Fill the background with transparent
	draw.Draw(rgba, rgba.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Create a freetype context for drawing
	c := freetype.NewContext()
	c.SetDPI(75)
	c.SetFont(mainFont)
	c.SetFontSize(25)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.White)
	c.SetSrc(image.NewUniform(color.RGBA{R:255, G:255, B:255, A:255}))

	// Draw the text at (ptX, ptY) in image coords
	pt := freetype.Pt(10, 40)
	_, err := c.DrawString(text, pt)
	if err != nil {
			log.Printf("drawText2D error: %v", err)
			return
	}

	var textTex uint32
	gl.GenTextures(1, &textTex)
	gl.BindTexture(gl.TEXTURE_2D, textTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(
			gl.TEXTURE_2D, 0, gl.RGBA,
			int32(imgWidth), int32(imgHeight), 0,
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix),
	)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Render the texture as a quad at (x, y) in NDC

	quadW := 0.35
	quadH := 0.1

	gl.Color3f(1, 1, 1)

	gl.BindTexture(gl.TEXTURE_2D, textTex)
	gl.Begin(gl.QUADS)
			gl.TexCoord2f(0, 0)
			gl.Vertex2f(x,       y)
			gl.TexCoord2f(1, 0)
			gl.Vertex2f(x+float32(quadW), y)
			gl.TexCoord2f(1, 1)
			gl.Vertex2f(x+float32(quadW), y-float32(quadH))
			gl.TexCoord2f(0, 1)
			gl.Vertex2f(x,       y-float32(quadH))
	gl.End()

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DeleteTextures(1, &textTex)
}


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
