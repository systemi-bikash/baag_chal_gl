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

// for dialog
var showDialogbox bool
var TigerWinsTitle = "Game Over-Tiger Wins"
var GoatWinsTitle = "Game Over-Goat Wins"
var TigerWinsDialogMessage = "Tiger has captured 5 goat.\n Start a new game ?"
var GoatWinsDialogMessage = "Tiger doesn't have any moves.\n Start a new game ?"
var ScreenWidth = 600
var ScreenHeight = 600

var newGameButton,cancelButton image.Rectangle


var (
	dialogActive bool
	dialogTitle string
	dialogMessage string
	dialogIcon string
	dialogOnNewGame func()
	dailogOnCancel func()
	dialogBoxRect = struct{
		x1,x2,y1,y2 float32
	}{
		x1: -0.5,
		y1: 0.3,
		x2: 0.5,
		y2: -0.2,
	}

	dialogButtonYesRect = struct {
		x1, y1, x2, y2 float32
}{
		x1: -0.3,
		y1: -0.15,
		x2: -0.1, y2:
		-0.25,
}
dialogButtonNoRect = struct {
		x1, y1, x2, y2 float32
}{
		x1: 0.1,
		y1: -0.15,
		x2: 0.3,
		y2: -0.25,
}
)


func showDialog(
	title string,
	message string,
	iconPath string,
	onNewGame func(),
	onCancel func(),
) {
	// Store the data so we can draw the dialog in drawDialogBox()
	dialogActive = true
	dialogTitle = title
	dialogMessage = message
	dialogIcon = iconPath
	dialogOnNewGame = onNewGame
	dailogOnCancel = onCancel

	log.Printf("[DIALOG] Title: %s\nMessage: %s\nIcon: %s", title, message, iconPath)
}

func drawUI() {
	// 1) Draw the reset button (simple rectangle) near the top-right
  gl.Color3f(0.2, 0.2, 0.2)
  gl.Begin(gl.QUADS)
  gl.Vertex2f(resetButtonRect.minX, resetButtonRect.minY)
  gl.Vertex2f(resetButtonRect.maxX, resetButtonRect.minY)
  gl.Vertex2f(resetButtonRect.maxX, resetButtonRect.maxY)
  gl.Vertex2f(resetButtonRect.minX, resetButtonRect.maxY)
  gl.End()

	// If you have a text rendering approach, you'd label the button:
  resetLabelX := (resetButtonRect.minX+resetButtonRect.maxX)/2 - 0.03
  resetLabelY := (resetButtonRect.minY+resetButtonRect.maxY)/2 + 0.02
  drawText2D(resetLabelX, resetLabelY, "RESET")

	// 2) Draw a banner for goat stats in the top-left corner
	goatsRemaining := maxGoats - placedGoats
    banner := fmt.Sprintf("Goats Placed: %d | Captured: %d | Remaining: %d",
        placedGoats, capturedGoats, goatsRemaining)

        drawText2D(-0.95, 0.92, banner)
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
  c.SetDPI(72)
  c.SetFont(mainFont)
  c.SetFontSize(28)
  c.SetClip(rgba.Bounds())
  c.SetDst(rgba)
  c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255}))

	// Draw the text at (ptX, ptY) in image coords
  pt := freetype.Pt(10, 44)
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

  quadW := float32(0.0065) * float32(len(text))
  quadH := 0.08

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
  gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
      int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()), 0,
      gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
  gl.BindTexture(gl.TEXTURE_2D, 0)

  return texture, nil
}
func drawDialogBox() {
  // 1) Dark overlay
  gl.Color4f(0.0, 0.0, 0.0, 0.5)
  gl.Begin(gl.QUADS)
  gl.Vertex2f(-1.0, 1.0)
  gl.Vertex2f(1.0, 1.0)
  gl.Vertex2f(1.0, -1.0)
  gl.Vertex2f(-1.0, -1.0)
  gl.End()

  // 2) Dialog box rectangle
  gl.Color3f(0.8, 0.8, 0.8)
  gl.Begin(gl.QUADS)
  gl.Vertex2f(dialogBoxRect.x1, dialogBoxRect.y1)
  gl.Vertex2f(dialogBoxRect.x2, dialogBoxRect.y1)
  gl.Vertex2f(dialogBoxRect.x2, dialogBoxRect.y2)
  gl.Vertex2f(dialogBoxRect.x1, dialogBoxRect.y2)
  gl.End()

  // 3) "Yes" button
  gl.Color3f(0.4, 0.8, 0.4)
  gl.Begin(gl.QUADS)
  gl.Vertex2f(dialogButtonYesRect.x1, dialogButtonYesRect.y1)
  gl.Vertex2f(dialogButtonYesRect.x2, dialogButtonYesRect.y1)
  gl.Vertex2f(dialogButtonYesRect.x2, dialogButtonYesRect.y2)
  gl.Vertex2f(dialogButtonYesRect.x1, dialogButtonYesRect.y2)
  gl.End()

  // 4) "No" button
  gl.Color3f(0.8, 0.4, 0.4)
  gl.Begin(gl.QUADS)
  gl.Vertex2f(dialogButtonNoRect.x1, dialogButtonNoRect.y1)
  gl.Vertex2f(dialogButtonNoRect.x2, dialogButtonNoRect.y1)
  gl.Vertex2f(dialogButtonNoRect.x2, dialogButtonNoRect.y2)
  gl.Vertex2f(dialogButtonNoRect.x1, dialogButtonNoRect.y2)
  gl.End()

  // 5) Title/message and button labels
  drawText2D(dialogBoxRect.x1+0.05, dialogBoxRect.y1-0.05, dialogTitle)
  drawText2D(dialogBoxRect.x1+0.05, dialogBoxRect.y1-0.15, dialogMessage)
  drawText2D(dialogButtonYesRect.x1+0.03, dialogButtonYesRect.y1-0.03, "Yes")
  drawText2D(dialogButtonNoRect.x1+0.03, dialogButtonNoRect.y1-0.03, "No")
}

