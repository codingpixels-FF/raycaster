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
		offsetX := 20
		offsetY := 20
		prefillBUfferWithImage(offsetX, offsetY, wallImage, wallImageImage, currentPixelBuffer)
		// Mouse
		mouseAbsX := int(raylib.GetMouseX())
		mouseAbsY := int(raylib.GetMouseY())

		setPixelWhite(wallImage, mouseAbsX-offsetX, mouseAbsY-offsetY)
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
		playerStatus := fmt.Sprintf("%.2d\n%.2d", mouseAbsX, mouseAbsY)
		raylib.DrawText(playerStatus, 20, screenHeight+10, 65, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\n")
		raylib.DrawText(playerStatusLegend, 200, screenHeight+10, 65, raylib.White)
		raylib.DrawFPS(screenWidth-90, screenHeight+10)

		raylib.BeginMode3D(camera)

		// Draw cube with an applied texture
		vec := raylib.Vector3{
			X: -2.0,
			Y: 2.0,
		}

		currentTexture := raylib.LoadTextureFromImage(wallImage)

		DrawCubeTexture(currentTexture, vec, 2.0, 2.0, 2.0, raylib.White)
		vec = raylib.Vector3{
			X: 0.0,
			Y: 1.0,
		}
		DrawCubeTexture(currentTexture, vec, 2.0, 2.0, 2.0, raylib.White)

		vec = raylib.Vector3{
			X: 2.0,
			Y: 1.0,
		}
		// Increase the angle (adjust speed as needed)
		angle += 0.05

		// Calculate new camera position to orbit around the target (center)
		radius := float32(10.0)
		camera.Position.X = float32(math.Cos(float64(angle))) * radius
		camera.Position.Z = float32(math.Sin(float64(angle))) * radius
		camera.Position.Y = 10.0 // Keep height constant, or change for a different effect

		// Make sure to update the target if needed, or keep it fixed
		camera.Target = raylib.Vector3{0, 0, 0}

		raylib.DrawGrid(10, 1.0) // Draw a grid
		raylib.EndMode3D()
		raylib.DrawFPS(screenWidth-100, 10)
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

func prefillBUfferWithImage(offsetX int, offsetY int, wallImage *raylib.Image, wallImageImage image.Image, currentPixelBuffer *[]uint32) {
	for rayX := offsetX; rayX < int(wallImage.Width); rayX++ {
		corX := int(int32(rayX)) - offsetX
		for rayY := offsetY; rayY < int(wallImage.Height); rayY++ {
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
