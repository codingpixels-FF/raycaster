package main

import (
	"fmt"
	"image"
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

func main() {
	// Set the maximum number of CPU cores to use
	// Set to debug or trace level for detailed logs
	raylib.SetTraceLogLevel(raylib.LogError)
	raylib.InitWindow(screenWidth, screenHeight+statusBarHeight, "Image mapping on objects in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("wall_all_small.png") // Loaded in GPU memory (VRAM)
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
		prefillBUfferWithImage(wallImage, wallImageImage, currentPixelBuffer)
		// Mouse
		deltaX := int(raylib.GetMouseX())
		deltaY := int(raylib.GetMouseY())

		setPixelWhite(wallImage, deltaX, deltaY)
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
		raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))
		raylib.DrawRectangle(0, screenHeight+10, screenWidth, screenHeight+statusBarHeight, raylib.NewColor(0, 0, 255, 255))
		//playerStatus := fmt.Sprintf("%.2f\n%.2f\n%.2f", player.x, player.y, player.angle)
		playerStatus := fmt.Sprintf("%.2d\n%.2d", deltaX, deltaY)
		raylib.DrawText(playerStatus, 20, screenHeight+10, 65, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\n")
		raylib.DrawText(playerStatusLegend, 200, screenHeight+10, 65, raylib.White)
		raylib.DrawFPS(screenWidth-90, screenHeight+10)

		raylib.EndDrawing()
		raylib.UnloadTexture(imageBufferTexture)
	}
}

func setPixelWhite(img *raylib.Image, x, y int) {
	if x < 0 || x >= int(img.Width) || y < 0 || y >= int(img.Height) {
		return // out of bounds
	}

	bytesPerPixel := 4 // RGBA
	index := (y*int(img.Width) + x) * bytesPerPixel
	size := int(img.Width) * int(img.Height) * bytesPerPixel

	// Convert unsafe.Pointer to a byte slice with the correct length
	dataSlice := (*[1 << 30]byte)(img.Data)[:size:size]

	// Set pixel to white
	dataSlice[index] = 255   // R
	dataSlice[index+1] = 255 // G
	dataSlice[index+2] = 255 // B
	dataSlice[index+3] = 255 // A
}

func prefillBUfferWithImage(wallImage *raylib.Image, wallImageImage image.Image, currentPixelBuffer *[]uint32) {
	for rayX := 0; rayX < renderWidth; rayX++ {
		corX := int(int32(rayX) % wallImage.Width)
		for rayY := 0; rayY < renderHeight; rayY++ {
			// Extract fill color and alpha
			fillColor := wallImageImage.At(corX, int(int32(rayY)%wallImage.Height))
			r, g, b, a := fillColor.RGBA()
			// Calculate scaling factor based on distance
			//r, g, b = addPixelEffects(currentDistance, r, g, b, numberOfMirrorsInWay)

			indexingCurrentPixelBuffer := *currentPixelBuffer
			indexingCurrentPixelBuffer[rayX+renderWidth*rayY] = a>>8<<24 | b>>8<<16 | g/3*2>>8<<8 | r/3*2>>8

		}
	}
}
