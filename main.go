package main

import (
	"fmt"
	"image"
	"math"
	_ "strconv"
	"sync"
	"unsafe"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth      = 1920
	screenHeight     = 1080
	screenPixelSize  = 4
	screenHeightHalf = screenHeight / 2
	statusBarHeight  = screenHeight / 2
	renderWidth      = screenWidth / 1
	renderHeight     = screenHeight / 1
	renderHeightHalf = renderHeight / 2
)

type PixelColor struct {
	R int
	G int
	B int
}

func main() {
	// Set the maximum number of CPU cores to use
	// Set to debug or trace level for detailed logs
	raylib.SetTraceLogLevel(raylib.LogError)
	raylib.InitWindow(screenWidth, screenHeight+statusBarHeight, "Image mapping on objects in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("repeat_triangle_64.png") // Loaded in GPU memory (VRAM)
	defer raylib.UnloadTexture(wallTexture)
	wallImage := raylib.LoadImageFromTexture(wallTexture)      // Loaded in CPU memory (RAM)
	raylib.ImageFormat(wallImage, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(wallImage)
	// Create a 2D array for the pixel data of one column
	wallImageImage := wallImage.ToImage()

	raylib.SetTargetFPS(6)
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

	// Define the camera to look into our 3d world
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

	angle := 0.0
	offsetX := 20
	offsetY := 20
	lastMouseAbsX := -1
	lastMouseAbsY := -1
	colorR := 0
	colorG := 0
	colorB := 0
	pixelColor := PixelColor{
		R: colorR,
		G: colorG,
		B: colorB,
	}
	colorPickerNumColors := 255
	colorPickerRectSizeX := int32(3)
	colorPickerRectSizeY := colorPickerRectSizeX * 4
	colorPickerColorPadding := int32(1)
	colorPickerStartPositionX := int32(100)
	colorPickerStartPositionY := colorPickerRectSizeX * 2

	for !raylib.WindowShouldClose() {
		var wg sync.WaitGroup
		// Check if we point to pixelBuffer2
		// Empty the other pixel buffer safely
		if &(*currentPixelBuffer)[0] == &pixelBuffer2[0] {
			currentPixelBuffer = &pixelBuffer1
			wg.Add(1)
			go func() {
				defer wg.Done()

				/*for i := range pixelBuffer2 {
					pixelBuffer2[i] = 0
				}*/
			}()
		} else {
			currentPixelBuffer = &pixelBuffer2
			wg.Add(1)
			go func() {
				defer wg.Done()
				/*for i := range pixelBuffer1 {
					pixelBuffer1[i] = 0
				}*/
			}()
		}

		prefillBUfferWithImage(offsetX, offsetY, wallImage, wallImageImage, currentPixelBuffer)

		//Draw color picker
		//for raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))

		// Parameters for color blocks

		ii := int32(0)
		for i := 0; i < colorPickerNumColors; i = i + 2 {
			ii++
			// Half-bright color
			newColor := raylib.NewColor(uint8(i), uint8(pixelColor.G), uint8(pixelColor.B), 255)
			raylib.DrawRectangle(
				colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding),
				colorPickerStartPositionY+colorPickerRectSizeY*1+colorPickerColorPadding,
				colorPickerRectSizeX,
				colorPickerRectSizeY,
				newColor,
			)

			// Half-bright color
			newColor = raylib.NewColor(uint8(pixelColor.R), uint8(i), uint8(pixelColor.B), 255)
			raylib.DrawRectangle(
				colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding),
				colorPickerStartPositionY+colorPickerRectSizeY*2+colorPickerColorPadding,
				colorPickerRectSizeX,
				colorPickerRectSizeY,
				newColor,
			)

			// Half-bright color
			newColor = raylib.NewColor(uint8(pixelColor.R), uint8(pixelColor.G), uint8(i), 255)
			raylib.DrawRectangle(
				colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding),
				colorPickerStartPositionY+colorPickerRectSizeY*3+colorPickerColorPadding,
				colorPickerRectSizeX,
				colorPickerRectSizeY,
				newColor,
			)
		}

		// Mouse
		mouseAbsX := int(raylib.GetMouseX())
		mouseAbsY := int(raylib.GetMouseY())
		mouseButton1Pressed := raylib.IsMouseButtonDown(raylib.MouseButtonLeft)
		mouseButton2Pressed := raylib.IsMouseButtonDown(raylib.MouseButtonRight)
		if mouseButton1Pressed {
			if lastMouseAbsX != -1 || lastMouseAbsY != -1 {
				setLineColor(wallImage, lastMouseAbsX-offsetX, lastMouseAbsY-offsetY, mouseAbsX-offsetX, mouseAbsY-offsetY, pixelColor)
			}

			//setPixelWhite(wallImage, mouseAbsX-offsetX, mouseAbsY-offsetY)
			lastMouseAbsX = mouseAbsX
			lastMouseAbsY = mouseAbsY
			if int32(mouseAbsX) > colorPickerStartPositionX && int32(mouseAbsX) < colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding) {
				startRY := colorPickerStartPositionY + colorPickerRectSizeY*1 + colorPickerColorPadding
				startGY := colorPickerStartPositionY + colorPickerRectSizeY*2 + colorPickerColorPadding
				startBY := colorPickerStartPositionY + colorPickerRectSizeY*3 + colorPickerColorPadding
				if int32(mouseAbsY) > startRY && int32(mouseAbsY) < startRY+colorPickerRectSizeY {
					colorR = mouseAbsX - int(colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding))
				}
				if int32(mouseAbsY) > startGY && int32(mouseAbsY) < startGY+colorPickerRectSizeY {
					colorG = mouseAbsX - int(colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding))
				}
				if int32(mouseAbsY) > startBY && int32(mouseAbsY) < startBY+colorPickerRectSizeY {
					colorB = mouseAbsX - int(colorPickerStartPositionX+ii*(colorPickerRectSizeX+colorPickerColorPadding))
				}
				pixelColor = PixelColor{
					R: colorR,
					G: colorG,
					B: colorB,
				}
			}
		}
		if mouseButton2Pressed {
			for yy := 0; yy <= int(wallImage.Height); yy++ {
				setLineColor(wallImage, 0, yy, int(wallImage.Width), yy, pixelColor)
			}
		}

		wallImageImage = wallImage.ToImage()

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
		defer raylib.UnloadImage(&img)

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.NewColor(0, 0, 0, 255))
		raylib.DrawTexturePro(imageBufferTexture, textureRectangle, finalRenderRectagle, raylib.Vector2{}, 0, raylib.White)

		// status bar overlay
		//raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))
		//raylib.DrawRectangle(0, screenHeight+10, screenWidth, screenHeight+statusBarHeight, raylib.NewColor(0, 0, 255, 255))
		//playerStatus := fmt.Sprintf("%.2f\n%.2f\n%.2f", player.x, player.y, player.angle)
		playerStatus := fmt.Sprintf("%.2d\n%.2d\n%.2d\n%.2d\n%.2d", mouseAbsX, mouseAbsY, colorR, colorG, colorB)
		raylib.DrawText(playerStatus, 20, screenHeight+10, 15, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\nR\nG\nB\n")
		raylib.DrawText(playerStatusLegend, 50, screenHeight+10, 15, raylib.White)

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
		angle += 0.05

		// Calculate new camera position to orbit around the target (center)
		radius := float32(4.0)
		camera.Position.X = float32(math.Cos(float64(angle))) * radius
		camera.Position.Z = float32(math.Sin(float64(angle))) * radius
		camera.Position.Y = 3.5 // Keep height constant, or change for a different effect

		camera.Target = raylib.Vector3{0, 0, 0}

		//raylib.DrawGrid(10, 1.0) // Draw a grid
		raylib.EndMode3D()
		raylib.DrawFPS(screenWidth-100, 10)
		raylib.EndDrawing()
		raylib.UnloadTexture(imageBufferTexture)
	}
}

func absDiffInt(x, y int) int {
	if x < y {
		return y - x
	}
	return x - y
}

func setLineColor(img *raylib.Image, lastX, lastY, currentX, currentY int, pixelColor PixelColor) {
	/*if lastX < 0 || lastX >= int(img.Width) || lastY < 0 || lastY >= int(img.Height) {
		return // out of bounds
	}
	if currentX < 0 || currentX >= int(img.Width) || currentY < 0 || currentY >= int(img.Height) {
		return // out of bounds
	}*/
	//setPixelWhite(img, currentX, currentY)
	//return
	// Bresenham's line
	dx := absDiffInt(currentX, lastX)
	dy := absDiffInt(currentY, lastY)

	var sx, sy int
	if lastX < currentX {
		sx = 1
	} else {
		sx = -1
	}

	if lastY < currentY {
		sy = 1
	} else {
		sy = -1
	}

	balanceErr := dx - dy

	x, y := lastX, lastY

	for {
		setPixelWhite(img, x, y, pixelColor)
		if x == currentX && y == currentY {
			break
		}
		e2 := 2 * balanceErr
		if e2 > -dy {
			balanceErr -= dy
			x += sx
		}
		if e2 < dx {
			balanceErr += dx
			y += sy
		}
	}
}

func setPixelWhite(img *raylib.Image, x, y int, pixelColor PixelColor) {
	if x < 0 || x >= int(img.Width) || y < 0 || y >= int(img.Height) {
		return // out of bounds
	}

	/*bytesPerPixel := 4 // RGBA

	reverseX := (int(img.Width) - 1 - x)
	reverseY := (int(img.Height) - 1 - y)

	size := int(img.Width) * int(img.Height) * bytesPerPixel

	// Convert unsafe.Pointer to a byte slice with the correct length
	dataSlice := (*[1 << 30]byte)(img.Data)[:size:size]

	// Set pixel to white
	indexLeftTop := (y*int(img.Width) + x) * bytesPerPixel
	dataSlice[indexLeftTop] = 255   // R
	dataSlice[indexLeftTop+1] = 255 // G
	dataSlice[indexLeftTop+2] = 255 // B
	dataSlice[indexLeftTop+3] = 255 // A

	// Set pixel to white
	indexRightBottom := (reverseY*int(img.Width) + reverseX) * bytesPerPixel
	dataSlice[indexRightBottom] = 255   // R
	dataSlice[indexRightBottom+1] = 255 // G
	dataSlice[indexRightBottom+2] = 255 // B
	dataSlice[indexRightBottom+3] = 255 // A

	// Set pixel to white
	indexRightTop := (y*int(img.Width) + reverseX) * bytesPerPixel
	dataSlice[indexRightTop] = 255   // R
	dataSlice[indexRightTop+1] = 255 // G
	dataSlice[indexRightTop+2] = 255 // B
	dataSlice[indexRightTop+3] = 255 // A

	// Set pixel to white
	indexLeftBottom := (reverseY*int(img.Width) + x) * bytesPerPixel
	dataSlice[indexLeftBottom] = 255   // R
	dataSlice[indexLeftBottom+1] = 255 // G
	dataSlice[indexLeftBottom+2] = 255 // B
	dataSlice[indexLeftBottom+3] = 255 // A*/

	// TODO FLIPP X FOR Y and LX FOR LY

	width, height := int(img.Width), int(img.Height)
	bytesPerPixel := 4

	dataSlice := (*[1 << 30]byte)(img.Data)[: width*height*bytesPerPixel : width*height*bytesPerPixel]

	// Generate symmetric positions

	xs := []int{x, width - 1 - x, y, height - 1 - y}

	for ixi, xi := range xs {
		for iyi, yi := range xs {
			if ixi == iyi { // Variations without repetition
				continue
			}
			if xi >= 0 && xi < width && yi >= 0 && yi < height {
				index := (yi*width + xi) * bytesPerPixel
				dataSlice[index] = byte(pixelColor.R)   //R
				dataSlice[index+1] = byte(pixelColor.G) //G
				dataSlice[index+2] = byte(pixelColor.B) //B
				dataSlice[index+3] = 255                //A
			}
		}
	}

}

func prefillBUfferWithImage(offsetX int, offsetY int, wallImage *raylib.Image, wallImageImage image.Image, currentPixelBuffer *[]uint32) {
	for rayX := offsetX; rayX < int(wallImage.Width)+offsetX; rayX++ {
		corX := int(int32(rayX)) - offsetX
		for rayY := offsetY; rayY < int(wallImage.Height)+offsetY; rayY++ {
			// Extract fill color and alpha
			fillColor := wallImageImage.At(corX, int(int32(rayY))-offsetY)
			r, g, b, a := fillColor.RGBA()
			// Calculate scaling factor based on distance
			//r, g, b = addPixelEffects(currentDistance, r, g, b, numberOfMirrorsInWay)

			indexingCurrentPixelBuffer := *currentPixelBuffer
			indexingCurrentPixelBuffer[rayX+renderWidth*rayY] = a>>8<<24 | b>>8<<16 | g/3*2>>8<<8 | r/3*2>>8

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
