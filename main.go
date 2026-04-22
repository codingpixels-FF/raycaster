package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"reflect"
	_ "strconv"
	"sync"
	"unsafe"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	mapData             [][]int
	mapWidth, mapHeight int
)

type ZBufferItem struct {
	Dist                 float64
	HitX                 float64
	HitY                 float64
	IsHitOnX             bool
	ItemID               int
	NumberOfMirrorsInWay int
}

func loadMap(filename string) {
	file, _ := os.Open(filename)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		row := make([]int, len(line))
		for i, ch := range line {
			if ch >= '0' && ch <= '9' {
				row[i] = int(ch - '0')
			} else {
				// handle non-digit characters if needed
				row[i] = 0
			}
		}
		mapData = append(mapData, row)
	}
	mapHeight = len(mapData)
	mapWidth = len(mapData[0])
}

// get radius of two 2D points
func getRadius(x, y, cx, cy float64) float64 {
	distance := math.Hypot(x-cx, y-cy)
	return distance
}

// Check if point (x0, y0) lies on the line defined by point (x1, y1) and angle theta
func getDistanceOfPointOnLineByAngle(x1, y1, cosTheta, sinTheta, x0, y0 float64) float64 {
	const epsilon = 0.005 // bigger number because the ray step is high

	// Direction vector of the line
	//rotrated by +PI/2
	dx := -sinTheta
	dy := cosTheta
	// Vector from the line point to the test point
	vx := x0 - x1
	vy := y0 - y1

	// Cross product
	cross := vx*dy - vy*dx

	if math.Abs(cross) > epsilon { // not the same line
		return 1.0
	}

	// Dot product to check if the point lies within the segment length
	dot := vx*dx + vy*dy
	if dot < -0.5 { //epsilon {
		return 1.0 // Point is behind the start point
	}

	if dot > 0.5 {
		return 1.0 // Point is beyond the segment length
	}

	// Check distance from start point to the point is less than or equal to 1
	distance := math.Hypot(vx, vy)
	if distance > 0.5 || distance < -0.5 {
		return 1.0
	}

	// Cross product against origin to determine approach from left or right
	cross = vx*sinTheta - vy*cosTheta
	if cross > 0 {
		return -distance
	}
	return +distance

}

// castRay function returns distance to the wall
func castRay(rayX float64, rayY float64, rayAngleRadians float64) []ZBufferItem {
	rayXOrigin := rayX
	rayYOrigin := rayY
	var zbufferSlice []ZBufferItem

	depth := depthStepHighDetail
	dCosAngle := math.Cos(rayAngleRadians)
	dSinAngle := math.Sin(rayAngleRadians)
	deltaX := depth * dCosAngle
	deltaY := depth * dSinAngle
	mapX := int(rayX)
	mapY := int(rayY)

	totalDepth := depth
	isSelfNotAddedInThisReflection := false // ignore first pass, you are not able to see self without a mirror
	isCurrentObjectServedInThisTile := false
	NumberOfMirrorsInWay := 0
	for rayX >= 0 && rayX < float64(mapWidth) && rayY >= 0 && rayY < float64(mapHeight) {
		totalDepth += depth
		if totalDepth > 30 || NumberOfMirrorsInWay > 3 {
			return zbufferSlice
		}
		rayX += deltaX
		lastMapX := mapX
		lastMapY := mapY
		mapX = int(rayX)
		mapY = int(rayY) // needs to be recalculated in case of reflection

		// level of detail of floor and ceiling based on depth
		if totalDepth > 1 && totalDepth < 2 && depth == depthStepHighDetail {
			depth = depthStepMediumDetail
			deltaX = depth * dCosAngle
			deltaY = depth * dSinAngle
		}
		if totalDepth > 2 && totalDepth < 3 && depth == depthStepMediumDetail {
			depth = depthStepLowDetail
			deltaX = depth * dCosAngle
			deltaY = depth * dSinAngle
		}
		if true {
			// Ceiling and floor
			// Track roof and floor first
			insideTileX := float64(rayX - math.Floor(rayX))
			insideTileY := float64(rayY + deltaY - math.Floor(rayY+deltaY)) // adjust Y for future movement
			newZBufferItem := ZBufferItem{totalDepth, insideTileX, insideTileY, true, 0, NumberOfMirrorsInWay}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
		}

		// Check for collision on X
		if mapData[mapY][mapX] == 1 {
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, true, 1, NumberOfMirrorsInWay}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			return zbufferSlice
		}
		// check of mirrors on X - left right
		if mapData[mapY][mapX] == 2 {
			rayY += deltaY // correct the ray
			deltaX = -deltaX
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, true, 2, NumberOfMirrorsInWay}
			NumberOfMirrorsInWay += 1
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			rayX += deltaX
			rayY += deltaY
			isSelfNotAddedInThisReflection = true
			//change the origin - after reflection on mirror
			dCosAngle = -dCosAngle // cos 180 is 1/2 period
			//dSinAngle = math.Sin(rayAngleRadians)
			continue

		}
		rayY += deltaY
		// check for collision on both (and assume it was Y)
		mapY = int(rayY)
		if lastMapX != mapX || lastMapY != mapY { // only one object at one tile
			isCurrentObjectServedInThisTile = false
		}
		if mapData[mapY][mapX] == 1 {
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, false, 1, NumberOfMirrorsInWay}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			return zbufferSlice
		}

		// check of mirrors on Y
		if mapData[mapY][mapX] == 2 { //up down
			deltaY = -deltaY
			dSinAngle = -dSinAngle // sin 180 is 1/2 period
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, false, 2, NumberOfMirrorsInWay}
			NumberOfMirrorsInWay += 1
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			rayX += deltaX
			rayY += deltaY
			isSelfNotAddedInThisReflection = true
			continue
		}

		// Handle player reflection
		if totalDepth > 0.6 && isSelfNotAddedInThisReflection {
			lenFromSelf := getRadius(rayXOrigin, rayYOrigin, rayX, rayY)
			if lenFromSelf < 0.5 {
				// Save player's sprite if present
				textureDistance := getDistanceOfPointOnLineByAngle(rayXOrigin, rayYOrigin, dCosAngle, dSinAngle, rayX, rayY)
				if textureDistance < 0.5 && textureDistance > -0.5 {
					newZBufferItem := ZBufferItem{totalDepth, textureDistance, 0, false, 9, NumberOfMirrorsInWay} // player
					zbufferSlice = append(zbufferSlice, newZBufferItem)
					isSelfNotAddedInThisReflection = false
				}
			}
		}
		mapObjectId := mapData[mapY][mapX]
		// Check of other objects (sprites) on the map
		if mapObjectId > 2 && mapObjectId < 9 && isCurrentObjectServedInThisTile == false { // 9 is player
			// Only execute if we are in the radius of the root of the square
			mapSquareCenterX := float64(mapX) + 0.5
			mapSquareCenterY := float64(mapY) + 0.5

			textureDistance := getDistanceOfPointOnLineByAngle(mapSquareCenterX, mapSquareCenterY, dCosAngle, dSinAngle, rayX, rayY)

			if textureDistance < 0.5 && textureDistance > -0.5 {
				// Other objects sprites
				newZBufferItem := ZBufferItem{totalDepth, textureDistance, 0, false, mapObjectId, NumberOfMirrorsInWay} // NPC
				zbufferSlice = append(zbufferSlice, newZBufferItem)
				isCurrentObjectServedInThisTile = true
			}
		}
	}
	newZBufferItem := ZBufferItem{5, 0, 0, false, 1, NumberOfMirrorsInWay} // default wall
	zbufferSlice = append(zbufferSlice, newZBufferItem)
	return zbufferSlice
}

const (
	screenWidth           = 1980
	screenHeight          = 1080
	screenHeightHalf      = screenHeight / 2
	statusBarHeight       = screenHeight / 5
	renderWidth           = screenWidth / 4
	renderHeight          = screenHeight / 4
	renderHeightHalf      = renderHeight / 2
	depthStepLowDetail    = 0.01
	depthStepMediumDetail = 0.007
	depthStepHighDetail   = 0.005
)

// Player properties
type Player struct {
	x, y  float64
	angle float64 // direction angle
}

func drawFloorAndCeiling(bufferImage []uint32, textureImageImage image.Image, hitX float64, hitY float64, wallHeightHalfLast int, wallHeightHalfCurrent int, rayX int, numberOfMirrorsInWay int, currentDistance float64) {
	// hitX is from 0-1
	// hitY is from 0-1
	textureImageImageWidth := 256
	textureImageImageHeight := 256
	// get color of the texture
	hitXonTexture := hitX * float64(textureImageImageWidth)
	hitYonTexture := hitY * float64(textureImageImageHeight)
	fillColor := textureImageImage.At(int(hitXonTexture), int(hitYonTexture))
	r, g, b, a := fillColor.RGBA()
	// Calculate scaling factor based on distance
	// Calculate scaling factor based on distance
	r, g, b = addPixelEffects(currentDistance, r, g, b, numberOfMirrorsInWay)
	//fillColorUint32 := a>>8<<24 | r>>8<<16 | g>>8<<8 | b>>8 bufferImage[rayX+renderWidth*y] = a>>8<<24 | b>>8<<16 | g>>8<<8 | r>>8
	fillColorCeilingUint32 := a>>8<<24 | b/4>>8<<16 | g/4>>8<<8 | r/4>>8
	fillColorFloorUint32 := a>>8<<24 | b/2>>8<<16 | g/2>>8<<8 | r/2>>8

	floorStartY := renderHeightHalf + wallHeightHalfLast
	floorEndY := renderHeightHalf + wallHeightHalfCurrent
	ceilingStartY := renderHeightHalf - wallHeightHalfCurrent
	ceilingEndY := renderHeightHalf - wallHeightHalfLast

	for y := floorStartY; y < floorEndY; y++ {
		/*if floorStartY < renderHeightHalf {
			floorStartY = renderHeightHalf
		}
		if floorEndY < renderHeightHalf {
			floorEndY = renderHeightHalf
		}
		if floorStartY > renderHeight {
			floorStartY = renderHeight
		}
		if floorEndY > renderHeight {
			floorEndY = renderHeight
		}*/
		if y <= renderHeightHalf {
			continue // too far, error
		}
		if y >= renderHeight {
			break // too close to the camera
		}
		bufferImage[rayX+renderWidth*y] = fillColorFloorUint32
	}

	for y := ceilingStartY; y < ceilingEndY; y++ {
		/*if ceilingStartY < 0 {
			ceilingStartY = 0
		}
		if ceilingEndY < 0 {
			ceilingEndY = 0
		}
		if ceilingStartY > renderHeightHalf {
			ceilingStartY = renderHeightHalf
		}
		if ceilingEndY > renderHeightHalf {
			ceilingEndY = renderHeightHalf
		}*/
		if y >= renderHeightHalf {
			continue // too far, error
		}
		if y < 0 {
			break // too close to the camera
		}
		bufferImage[rayX+renderWidth*y] = fillColorCeilingUint32
	}
}

func drawSprite(bufferImage []uint32, textureImageImage image.Image, hitX float64, currentDistance float64, rayX int, isMirror bool, isWall bool, isHitOnX bool, numberOfMirrorsInWay int) {
	// hitX is from 0-1
	// hitY is from 0-1
	textureImageImageWidth := 256
	textureImageImageHeight := 256
	// get color of the texture
	hitXonTexture := hitX * float64(textureImageImageWidth)

	wallHeightCurrent := float32(renderHeight / (currentDistance + 0.0001))

	spriteStartY := int(renderHeightHalf - wallHeightCurrent/2)
	spriteEndY := int(renderHeightHalf + wallHeightCurrent/2)
	texCorY := 0

	texSizeY := float64(spriteEndY - spriteStartY)

	for y := spriteStartY; y < spriteEndY; y++ {
		texCorY++

		// handle when the sprite is too close to the camera
		// We are unable to draw outside the render frame
		if y < 0 {
			continue // Top part of the sprite is exceeding camera view
		}
		if y >= renderHeight {
			break // Lower part of the sprite is exceeding camera view, we are done
		}

		hitYonTexture := float64(texCorY) / texSizeY * float64(textureImageImageHeight)

		// Extract fill color and alpha
		fillColor := textureImageImage.At(int(hitXonTexture), int(hitYonTexture))
		r, g, b, a := fillColor.RGBA()
		// Calculate scaling factor based on distance
		r, g, b = addPixelEffects(currentDistance, r, g, b, numberOfMirrorsInWay)
		if isWall {

			if isHitOnX {
				bufferImage[rayX+renderWidth*y] = a>>8<<24 | b>>8<<16 | g/3*2>>8<<8 | r/3*2>>8
			} else {
				bufferImage[rayX+renderWidth*y] = a>>8<<24 | b/4*3>>8<<16 | g/3*2>>8<<8 | r/3*2>>8
			}

		} else {
			if a == 0xffff { // simple transparency
				if isMirror {
					if isHitOnX {
						bufferImage[rayX+renderWidth*y] = a>>8<<24 | b>>8<<16 | g/3*2>>8<<8 | r/3*2>>8
					} else {
						bufferImage[rayX+renderWidth*y] = a>>8<<24 | b/3*2>>8<<16 | g/3*2>>8<<8 | r/3*2>>8
					}
				} else {
					// 2D sprite
					bufferImage[rayX+renderWidth*y] = a>>8<<24 | b>>8<<16 | g>>8<<8 | r>>8
				}
			}
		}
	}
}

func addPixelEffects(currentDistance float64, r uint32, g uint32, b uint32, numberOfMirrorsInWay int) (uint32, uint32, uint32) {
	randomValue := rand.Float64() + 0.5
	scale := 1.0 / (1.0 + randomValue*currentDistance/5)
	r = uint32(scale * (float64(r)))
	g = uint32(scale * (float64(g)))
	b = uint32(scale * (float64(b)))

	for _ = range numberOfMirrorsInWay {
		r = uint32(0.7 * float32(r))
		g = uint32(0.7 * float32(g))
		b = uint32(0.7 * float32(b))
	}
	if currentDistance < 2*randomValue {
		b = uint32(float64(b) * 1.2)
	}
	if currentDistance < 2.5*randomValue {
		b = uint32(float64(b) * 1.2)
	}
	if currentDistance < 3*randomValue {
		b = uint32(float64(b) * 1.2)
	}
	return r, g, b
}

func main() {
	// Set the maximum number of CPU cores to use
	// Set to debug or trace level for detailed logs
	raylib.SetTraceLogLevel(raylib.LogError)
	raylib.InitWindow(screenWidth, screenHeight+statusBarHeight, "Raycasting in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("wall_all_small.png") // Loaded in GPU memory (VRAM)
	defer raylib.UnloadTexture(wallTexture)
	wallImage := raylib.LoadImageFromTexture(wallTexture)      // Loaded in CPU memory (RAM)
	raylib.ImageFormat(wallImage, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(wallImage)
	// Create a 2D array for the pixel data of one column
	wallImageImage := wallImage.ToImage()

	wizardTexture := raylib.LoadTexture("256_wizzard1.png")
	defer raylib.UnloadTexture(wizardTexture)
	wizardImage := raylib.LoadImageFromTexture(wizardTexture)    // Loaded in CPU memory (RAM)
	raylib.ImageFormat(wizardImage, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(wizardImage)
	wizardImageImage := wizardImage.ToImage()

	playerFrontTexture := raylib.LoadTexture("256_sorcerer1.png")
	defer raylib.UnloadTexture(playerFrontTexture)
	playerFrontImage := raylib.LoadImageFromTexture(playerFrontTexture) // Loaded in CPU memory (RAM)
	raylib.ImageFormat(playerFrontImage, raylib.UncompressedR8g8b8a8)   // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(playerFrontImage)
	playerFrontImageImage := playerFrontImage.ToImage()

	playerHandsTexture := raylib.LoadTexture("256_sorcerer1_hand.png")
	defer raylib.UnloadTexture(playerHandsTexture)
	//playerHandsImage := raylib.LoadImageFromTexture(playerHandsTexture) // Loaded in CPU memory (RAM)
	//raylib.ImageFormat(playerHandsImage, raylib.UncompressedR8g8b8a8)   // Format image to RGBA 32bit (required for texture update)
	//defer raylib.UnloadImage(playerHandsImage)
	//playerHandsImageImage := playerHandsImage.ToImage()

	mirrorTexture := raylib.LoadTexture("256_mirror.png")
	defer raylib.UnloadTexture(mirrorTexture)
	mirrorImage := raylib.LoadImageFromTexture(mirrorTexture)    // Loaded in CPU memory (RAM)
	raylib.ImageFormat(mirrorImage, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(mirrorImage)
	mirrorImageImage := mirrorImage.ToImage()

	warrior2Texture := raylib.LoadTexture("256_warrior2.png")
	defer raylib.UnloadTexture(warrior2Texture)
	warrior2Image := raylib.LoadImageFromTexture(warrior2Texture)  // Loaded in CPU memory (RAM)
	raylib.ImageFormat(warrior2Image, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(warrior2Image)
	warrior2ImageImage := warrior2Image.ToImage()

	warrior1Texture := raylib.LoadTexture("256_warrior1.png")
	defer raylib.UnloadTexture(warrior1Texture)
	warrior1Image := raylib.LoadImageFromTexture(warrior1Texture)  // Loaded in CPU memory (RAM)
	raylib.ImageFormat(warrior1Image, raylib.UncompressedR8g8b8a8) // Format image to RGBA 32bit (required for texture update)
	defer raylib.UnloadImage(warrior1Image)
	warrior1ImageImage := warrior1Image.ToImage()

	loadMap("map.txt")

	// Player
	player := Player{
		x:     15.5,
		y:     1.5,
		angle: math.Pi / 2, // looking straight ahead
	}

	// Camera settings
	fov := math.Pi / 2.0 // 90 degrees

	raylib.SetTargetFPS(30)
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

	playerHandRanderRectagle := raylib.Rectangle{
		X:      screenWidth / 2,
		Y:      0,
		Width:  screenWidth / 2,
		Height: screenHeight,
	}

	playerHandstextureRectangle := raylib.Rectangle{
		X:      0,
		Y:      0,
		Width:  256,
		Height: 256,
	}

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

	raylib.HideCursor()
	raylib.DisableCursor()
	raylib.SetMousePosition(screenWidth/2, screenHeight/2)

	for !raylib.WindowShouldClose() {
		var wg sync.WaitGroup
		// Check if we point to pixelBuffer2
		// Empty the other pixel buffer safely
		if &(*currentPixelBuffer)[0] == &pixelBuffer2[0] {
			currentPixelBuffer = &pixelBuffer1
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range pixelBuffer2 {
					pixelBuffer2[i] = 0
				}
			}()
		} else {
			currentPixelBuffer = &pixelBuffer2
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range pixelBuffer1 {
					pixelBuffer1[i] = 0
				}
			}()
		}

		// Mouse

		deltaX := float64(raylib.GetMouseDelta().X)
		player.angle += deltaX * 0.01
		raylib.SetMousePosition(screenWidth/2, screenHeight/2)

		// Keyboard
		// Turn left
		if raylib.IsKeyDown(raylib.KeyLeft) || raylib.IsKeyDown(raylib.KeyQ) {
			player.angle -= 0.05
		}

		// Turn right
		if raylib.IsKeyDown(raylib.KeyRight) || raylib.IsKeyDown(raylib.KeyE) {
			player.angle += 0.05
		}
		factor := 0.1

		if raylib.IsKeyDown(raylib.KeyLeftShift) {
			factor = 0.5
		}

		futurePlayerDeltaX := 0.0
		futurePlayerDeltaY := 0.0
		// Forward
		if raylib.IsKeyDown(raylib.KeyW) {
			futurePlayerDeltaX += math.Cos(player.angle)
			futurePlayerDeltaY += math.Sin(player.angle)
		}

		// Backward
		if raylib.IsKeyDown(raylib.KeyS) {
			futurePlayerDeltaX -= math.Cos(player.angle)
			futurePlayerDeltaY -= math.Sin(player.angle)
		}

		// Strafe left
		if raylib.IsKeyDown(raylib.KeyA) {
			futurePlayerDeltaX += math.Cos(player.angle - math.Pi/2)
			futurePlayerDeltaY += math.Sin(player.angle - math.Pi/2)
		}

		// Strafe right
		if raylib.IsKeyDown(raylib.KeyD) {
			futurePlayerDeltaX += math.Cos(player.angle + math.Pi/2)
			futurePlayerDeltaY += math.Sin(player.angle + math.Pi/2)
		}
		playerPadding := 0.35
		directionXNorm := futurePlayerDeltaX
		directionYNorm := futurePlayerDeltaY
		futurePlayerAbsX := player.x + directionXNorm*factor
		futurePlayerAbsY := player.y + directionYNorm*factor

		/*
			Checking for collision around player
			***
			*P*
			***
		*/
		isCollisionXFree := true
		isCollisionYFree := true
		for {
			countOfCollision := 0
			for paddingY := -playerPadding; paddingY <= playerPadding; paddingY += 2 * playerPadding {
				for paddingX := -playerPadding; paddingX <= playerPadding; paddingX += 2 * playerPadding {
					mapX := int(futurePlayerAbsX + paddingX)
					mapY := int(player.y + paddingY)

					mapTileIndex := mapData[mapY][mapX]
					if mapTileIndex != 0 {
						isCollisionXFree = false
						countOfCollision += 1
					}

					mapX = int(player.x + paddingX)
					mapY = int(futurePlayerAbsY + paddingY)
					mapTileIndex = mapData[mapY][mapX]
					if mapTileIndex != 0 {
						isCollisionYFree = false
						countOfCollision += 1
					}

				}
			}
			// if both false, jump around in order not to get stuck on the edge
			if isCollisionXFree || isCollisionYFree {
				break
			} else if countOfCollision == 2 { // Edge only
				// move player back and let him move only on one axis (Y)
				println("Collision XY Free")
				player.x -= futurePlayerDeltaX * factor
				player.y -= futurePlayerDeltaY * factor
				isCollisionYFree = true
				break
			} else {
				// not on the edge, player is stuck correctly behind walls
				break
			}
		}

		if isCollisionXFree {
			player.x = futurePlayerAbsX
		}
		if isCollisionYFree {
			player.y = futurePlayerAbsY
		}

		for ray := 0; ray < renderWidth; ray++ {
			wg.Add(1)
			go func(ray int) {
				defer wg.Done()

				rayAngleRad := (float64(ray)/float64(renderWidth)-0.5)*fov + player.angle // ray angles in Rad
				zbufferSlice := castRay(player.x, player.y, rayAngleRad)

				lastDistance := float64(-1)

				for i := len(zbufferSlice) - 1; i >= 0; i-- {
					zbufferItem := zbufferSlice[i]
					itemId := zbufferItem.ItemID
					itemDist := zbufferItem.Dist
					hitX := zbufferItem.HitX
					hitY := zbufferItem.HitY
					isHitOnX := zbufferItem.IsHitOnX
					numberOfMirrorsInWay := zbufferItem.NumberOfMirrorsInWay
					// Calculate the angle difference
					angleDiff := rayAngleRad - player.angle

					// Perspective correction
					currentDistance := itemDist * math.Cos(angleDiff)

					// Texture coordinate
					if itemId == 0 {
						if lastDistance == -1 {
							lastDistance = currentDistance
							continue
						}

						// Check if there is at least one pixel difference
						wallHeightHalfCurrent := int(float32(renderHeight/(currentDistance+0.0001)) / 2)
						wallHeightHalfLast := int(float32(renderHeight/(lastDistance+0.0001)) / 2)

						if wallHeightHalfCurrent != wallHeightHalfLast {
							lastDistance = currentDistance
							drawFloorAndCeiling(*currentPixelBuffer, wallImageImage, hitX, hitY, wallHeightHalfLast, wallHeightHalfCurrent, ray, numberOfMirrorsInWay, currentDistance)
						}

					} else if itemId >= 1 && itemId <= 2 {

						var imageOnWall image.Image
						if itemId == 1 {
							imageOnWall = wallImageImage
						}
						if itemId == 2 {
							imageOnWall = mirrorImageImage
						}

						var texX float32
						if isHitOnX {
							texX = float32(hitY - math.Floor(hitY))
						} else {
							texX = float32(hitX - math.Floor(hitX))
						}

						// Draw textured slice
						isWall := true
						isMirror := false
						if itemId == 2 {
							isMirror = true
							isWall = false
						}

						drawSprite(*currentPixelBuffer, imageOnWall, float64(texX), currentDistance, ray, isMirror, isWall, isHitOnX, numberOfMirrorsInWay)

					} else {
						// sprites

						var textureImage image.Image
						if itemId == 6 {
							textureImage = warrior1ImageImage
						} else if itemId == 8 {
							textureImage = wizardImageImage
						} else if itemId == 9 {
							// self player
							textureImage = playerFrontImageImage
						} else if itemId == 7 {
							textureImage = warrior2ImageImage
						}

						texX := 0.5 - float32(hitX) // map to textureSprite width

						drawSprite(*currentPixelBuffer, textureImage, float64(texX), currentDistance, ray, false, false, false, numberOfMirrorsInWay)
					}
				}
			}(ray)
		}

		// Close channel once all goroutines are done
		wg.Wait()

		raylib.UpdateTexture(imageBufferTexture, uint32SliceToRGBA2(*currentPixelBuffer))
		//defer raylib.UnloadImage(&img)

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.NewColor(0, 0, 0, 255))
		raylib.DrawTexturePro(imageBufferTexture, textureRectangle, finalRenderRectagle, raylib.Vector2{}, 0, raylib.White)

		raylib.DrawTexturePro(playerHandsTexture, playerHandstextureRectangle, playerHandRanderRectagle, raylib.Vector2{}, 0, raylib.White)

		// status bar overlay
		raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))
		raylib.DrawRectangle(0, screenHeight+10, screenWidth, screenHeight+statusBarHeight, raylib.NewColor(0, 0, 255, 255))
		playerStatus := fmt.Sprintf("%.2f\n%.2f\n%.2f", player.x, player.y, player.angle)
		raylib.DrawText(playerStatus, 20, screenHeight+10, 65, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\n>")
		raylib.DrawText(playerStatusLegend, 200, screenHeight+10, 65, raylib.White)
		raylib.DrawFPS(screenWidth-90, screenHeight+10)

		raylib.EndDrawing()
		//raylib.UnloadTexture(imageBufferTexture)
	}
}

func uint32SliceToRGBA(slice []uint32) []color.RGBA {
	var rgbaSlice []color.RGBA
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&rgbaSlice))
	hdr.Data = uintptr(unsafe.Pointer(&slice[0]))
	hdr.Len = len(slice)
	hdr.Cap = len(slice)
	return rgbaSlice
}

// Convert []uint32 to []color.RGBA without copying
func uint32SliceToRGBA2(slice []uint32) []color.RGBA {
	rgba := make([]color.RGBA, len(slice))
	for i, v := range slice {
		rgba[i] = color.RGBA{
			A: uint8(v >> 24),
			B: uint8(v >> 16),
			G: uint8(v >> 8),
			R: uint8(v),
		}
	}
	return rgba
}
