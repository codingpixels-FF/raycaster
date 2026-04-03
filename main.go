package main

import (
	"coding-pixels/raycasting/codingpixels"
	"fmt"
	"image"
	"math"
	"unsafe"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth      = 1980 * 2
	screenHeight     = 1080
	screenPixelSize  = 4
	screenHeightHalf = screenHeight / 2
	statusBarHeight  = screenHeight / 2
	renderWidth      = screenWidth / 1
	renderHeight     = screenHeight / 1
	renderHeightHalf = renderHeight / 2
)

func main() {
	// Set the maximum number of CPU cores to use
	// Set to debug or trace level for detailed logs
	raylib.SetTraceLogLevel(raylib.LogError)
	raylib.InitWindow(screenWidth, screenHeight+statusBarHeight, "Symmetric Dynamic Texture in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("question_32.png") // Loaded in GPU memory (VRAM)
	defer raylib.UnloadTexture(wallTexture)

	// Create a 2D array for the pixel data of one column
	raylib.SetTargetFPS(15)
	pixelBuffer1 := make([]uint32, renderWidth*renderHeight)
	pixelBuffer2 := make([]uint32, renderWidth*renderHeight)
	currentPixelBuffer := &pixelBuffer1

	// render
	dataPtr := unsafe.Pointer(&(*currentPixelBuffer)[0])

	img := raylib.Image{
		Data:    dataPtr,
		Width:   renderWidth,
		Height:  renderHeight,
		Mipmaps: 1,
		Format:  raylib.UncompressedR8g8b8a8,
	}

	finalRenderRectagle := raylib.Rectangle{
		X:      0,
		Y:      0,
		Width:  screenWidth,
		Height: screenHeight,
	}

	textureRectangle := raylib.Rectangle{
		X:      0,
		Y:      0,
		Width:  renderWidth,
		Height: renderHeight,
	}

	offsetX := 20
	offsetY := 20
	lastMouseAbsX := -1
	lastMouseAbsY := -1
	//scaleOnX := 0.01
	textureScaleZoom := 8
	wallImage := raylib.LoadImageFromTexture(wallTexture)    // Loaded in CPU memory (RAM)
	raylib.ImageFormat(wallImage, raylib.UncompressedR8g8b8) // Format image to RGB 24bit (no alpha channel)
	// color picker
	colorPicker := codingpixels.NewColorPicker(254, 4, 20, 1, 2, int32(textureScaleZoom)*wallImage.Width+int32(offsetX*2), 0, 2)
	//colorPickerSingleColorRectSizeX := float64(colorPicker.SingleColorRectSizeX)

	// Define the camera to look into our 3d world
	camera := raylib.Camera{
		Position: raylib.Vector3{
			Y: 10.0,
			Z: 10.0,
		},
		Target:     raylib.Vector3{},
		Up:         raylib.Vector3{Y: 1.0},
		Fovy:       45.0,
		Projection: raylib.CameraPerspective,
	}
	angle := 35.0
	editTextureYsizeScaled := textureScaleZoom * int(wallImage.Height)

	for !raylib.WindowShouldClose() {

		defer raylib.UnloadImage(wallImage)

		// Check if we point to pixelBuffer2
		// Empty the other pixel buffer safely
		if &(*currentPixelBuffer)[0] == &pixelBuffer2[0] {
			currentPixelBuffer = &pixelBuffer1
		} else {
			currentPixelBuffer = &pixelBuffer2
		}

		wallImageImage := wallImage.ToImage()
		prefillBufferWithImage(offsetX, offsetY, textureScaleZoom, wallImage, wallImageImage, currentPixelBuffer)

		// Color picker position
		colorPicker.Render()

		// Mouse
		mouseAbsX := int(raylib.GetMouseX())
		mouseAbsY := int(raylib.GetMouseY())
		//mouseButton1Pressed := raylib.IsMouseButtonPressed(raylib.MouseButtonLeft)
		mouseButton1Down := raylib.IsMouseButtonDown(raylib.MouseButtonLeft)
		mouseButton1Up := raylib.IsMouseButtonUp(raylib.MouseButtonLeft)
		if mouseButton1Up {
			lastMouseAbsX = -1
			lastMouseAbsY = -1
		}
		mouseButton2Pressed := raylib.IsMouseButtonDown(raylib.MouseButtonRight)
		if mouseButton1Down {
			if mouseAbsX > 0 && mouseAbsX < editTextureYsizeScaled+offsetX*2 && mouseAbsY > 0 && mouseAbsY < editTextureYsizeScaled+offsetY {
				if lastMouseAbsX != -1 || lastMouseAbsY != -1 {
					// reduce for scale
					imageLastX := (lastMouseAbsX - offsetX) / textureScaleZoom
					imageLastY := (lastMouseAbsY - offsetY) / textureScaleZoom
					imageX := (mouseAbsX - offsetX) / textureScaleZoom
					imageY := (mouseAbsY - offsetY) / textureScaleZoom
					setLineColor(wallImage, imageLastX, imageLastY, imageX, imageY, colorPicker.GetPixelColor())
				}
				lastMouseAbsX = mouseAbsX
				lastMouseAbsY = mouseAbsY
			}

			if int32(mouseAbsX) > colorPicker.StartPositionX && int32(mouseAbsX) < colorPicker.ColorPickerEndPositionX {
				colorPicker.UpdateColors(int32(mouseAbsX), int32(mouseAbsY))
				// reset lines
				lastMouseAbsX = -1
				lastMouseAbsY = -1
			}
		}
		if mouseButton2Pressed {
			for yy := 0; yy < int(wallImage.Height)-2; yy++ {
				setLineColor(wallImage, 0, yy, int(wallImage.Width)-2, yy, colorPicker.GetPixelColor())
			}
		}

		// Contrast background for texture
		raylib.DrawRectangle(
			0,
			0,
			int32(editTextureYsizeScaled+2*offsetX),
			int32(editTextureYsizeScaled+3*offsetX),
			raylib.NewColor(60, 60, 60, 255),
		)
		// Currect color indicator
		pixelColor := colorPicker.GetPixelColor()
		raylib.DrawRectangle(
			int32(offsetX),
			int32(editTextureYsizeScaled+2*offsetY),
			wallImage.Width,
			wallImage.Height,
			raylib.NewColor(pixelColor.R, pixelColor.G, pixelColor.B, 255),
		)

		// Switch the image buffer
		dataPtr = unsafe.Pointer(&(*currentPixelBuffer)[0])
		img = raylib.Image{
			Data:    dataPtr,
			Width:   renderWidth,
			Height:  renderHeight,
			Mipmaps: 1,
			Format:  raylib.UncompressedR8g8b8a8,
		}

		imageBufferTexture := raylib.LoadTextureFromImage(&img)

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.NewColor(0, 0, 0, 255))
		raylib.DrawTexturePro(imageBufferTexture, textureRectangle, finalRenderRectagle, raylib.Vector2{}, 0, raylib.White)

		// status bar overlay
		//raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))
		//raylib.DrawRectangle(0, screenHeight+10, screenWidth, screenHeight+statusBarHeight, raylib.NewColor(0, 0, 255, 255))
		//playerStatus := fmt.Sprintf("%.2f\n%.2f\n%.2f", player.x, player.y, player.angle)
		playerStatus := fmt.Sprintf("%.2d\n%.2d\n%.2d\n%.2d\n%.2d\n%.2d\n%.2d", mouseAbsX, mouseAbsY, pixelColor.R, pixelColor.G, pixelColor.B, lastMouseAbsX, lastMouseAbsY)
		raylib.DrawText(playerStatus, screenWidth-150, screenHeight+10, 15, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\nR\nG\nB\nLastMouseX\nLastMouseY")
		raylib.DrawText(playerStatusLegend, screenWidth-100, screenHeight+10, 15, raylib.White)
		raylib.DrawFPS(screenWidth-100, screenHeight+150)

		raylib.BeginMode3D(camera)

		// Draw cube with an applied texture

		currentTexture := raylib.LoadTextureFromImage(wallImage)

		for xx := float32(-1.0); xx <= 1.0; xx = xx + 1.0 {

			vec := raylib.Vector3{
				X: xx,
				Y: -1.0,
			}
			DrawCubeTexture(currentTexture, vec, 1.0, 1.0, 1.0, raylib.White)

			vec = raylib.Vector3{
				X: xx,
				Y: 1.0,
			}
			DrawCubeTexture(currentTexture, vec, 1.0, 1.0, 1.0, raylib.White)

			if xx != 0 {
				vec = raylib.Vector3{
					X: xx,
					Y: 0.0,
				}
				DrawCubeTexture(currentTexture, vec, 1.0, 1.0, 1.0, raylib.White)
			}
		}

		// Increase the angle (adjust speed as needed)
		angle += 0.005

		// Calculate new camera position to orbit around the target (center)
		radius := float32(4.0)
		camera.Position.X = float32(math.Cos(float64(angle))) * radius
		camera.Position.Z = float32(math.Sin(float64(angle))) * radius
		camera.Position.Y = 3.5 // Keep height constant, or change for a different effect

		camera.Target = raylib.Vector3{0, 0, 0}

		//raylib.DrawGrid(10, 1.0) // Draw a grid
		raylib.EndMode3D()

		raylib.EndDrawing()
		raylib.UnloadTexture(imageBufferTexture)
		//raylib.UnloadImage(wallImage)
	}
}

func absDiffInt(x, y int) int {
	if x < y {
		return y - x
	}
	return x - y
}

// Helper functions
func abs(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func round(a float64) float64 {
	if a-float64(int(a)) >= 0.5 {
		return float64(int(a) + 1)
	}
	return float64(int(a))
}

func setLineColor(img *raylib.Image, lastX, lastY, currentX, currentY int, pixelColor codingpixels.PixelColor) {

	// Digital Differential Analyzer (DDA) algorithm
	dx := float64(currentX - lastX)
	dy := float64(currentY - lastY)

	steps := int(max(abs(dx), abs(dy)))

	// Calculate the increment for each step
	xIncrement := dx / float64(steps)
	yIncrement := dy / float64(steps)

	x := float64(lastX)
	y := float64(lastY)

	for i := 0; i <= steps; i++ {
		setPixelWhite(img, int(round(x)), int(round(y)), pixelColor)
		x += xIncrement
		y += yIncrement
	}
}

func setPixelWhite(img *raylib.Image, x, y int, pixelColor codingpixels.PixelColor) {
	if x < 0 || x >= int(img.Width) || y < 0 || y >= int(img.Height) {
		return // out of bounds
	}

	width, height := int(img.Width), int(img.Height)
	bytesPerPixel := 3

	dataSlice := (*[1 << 30]byte)(img.Data)[: width*height*bytesPerPixel : width*height*bytesPerPixel]

	// Generate symmetric positions
	xs := []int{x, x, width - x - 1, width - x - 1}
	ys := []int{y, height - y - 1, y, height - y - 1}
	for ixi, xi := range xs {
		yi := ys[ixi]
		index := (yi*width + xi) * bytesPerPixel
		dataSlice[index] = pixelColor.R   //R
		dataSlice[index+1] = pixelColor.G //G
		dataSlice[index+2] = pixelColor.B //B
		//dataSlice[index+3] = 255          //A

		// flip rows and columns
		index = (yi + xi*width) * bytesPerPixel
		dataSlice[index] = pixelColor.R   //R
		dataSlice[index+1] = pixelColor.G //G
		dataSlice[index+2] = pixelColor.B //B
		//dataSlice[index+3] = 255          //A

	}
}

func prefillBufferWithImage(offsetX int, offsetY int, scale int, wallImage *raylib.Image, wallImageImage image.Image, currentPixelBuffer *[]uint32) {
	for rayX := offsetX; rayX < int(wallImage.Width)+offsetX; rayX++ {
		corX := int(int32(rayX)) - offsetX
		for rayY := offsetY; rayY < int(wallImage.Height)+offsetY; rayY++ {
			// Extract fill color and alpha
			corY := int(int32(rayY)) - offsetY
			fillColor := wallImageImage.At(corX, corY)
			r, g, b, a := fillColor.RGBA()
			indexingCurrentPixelBuffer := *currentPixelBuffer
			color := a>>8<<24 | b>>8<<16 | g>>8<<8 | r>>8
			scalePixel(indexingCurrentPixelBuffer, offsetX+corX*scale, offsetY+corY*scale, renderWidth, scale, color)
		}
	}
}

// scalePixel replicates a pixel at (startX, startY) into a block of size scale x scale
func scalePixel(buffer []uint32, startX, startY, renderWidth, scale int, color uint32) {
	for y := 0; y < scale; y++ {
		for x := 0; x < scale; x++ {
			offsetX := startX + x
			offsetY := startY + y
			index := offsetX + renderWidth*offsetY
			buffer[index] = color
		}
	}
}

// DrawCubeTexture draws a textured cube
// NOTE: Cube position is the center position
func DrawCubeTexture(texture raylib.Texture2D, position raylib.Vector3, width, height, length float32, color raylib.Color) {
	x := position.X
	y := position.Y
	z := position.Z

	// Set desired texture to be enabled while drawing following vertex data
	raylib.SetTexture(texture.ID)

	raylib.Begin(raylib.Quads)
	raylib.Color4ub(color.R, color.G, color.B, color.A)
	// Front Face
	raylib.Normal3f(0.0, 0.0, 1.0) // Normal Pointing Towards Viewer
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	// Back Face
	raylib.Normal3f(0.0, 0.0, -1.0) // Normal Pointing Away From Viewer
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	// Top Face
	raylib.Normal3f(0.0, 1.0, 0.0) // Normal Pointing Up
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad.
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x-width/2, y+height/2, z+length/2) // Bottom Left Of The Texture and Quad
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x+width/2, y+height/2, z+length/2) // Bottom Right Of The Texture and Quad
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad Bottom Face

	raylib.Normal3f(0.0, -1.0, 0.0) // Normal Pointing Down
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x-width/2, y-height/2, z-length/2) // Top Right Of The Texture and Quad
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x+width/2, y-height/2, z-length/2) // Top Left Of The Texture and Quad
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	// Right face
	raylib.Normal3f(1.0, 0.0, 0.0) // Normal Pointing Right
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	// Left Face
	raylib.Normal3f(-1.0, 0.0, 0.0) // Normal Pointing Left
	raylib.TexCoord2f(0.0, 0.0)
	raylib.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	raylib.TexCoord2f(1.0, 0.0)
	raylib.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	raylib.TexCoord2f(1.0, 1.0)
	raylib.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	raylib.TexCoord2f(0.0, 1.0)
	raylib.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad

	raylib.End()

	raylib.SetTexture(0)
}
