package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	_ "strconv"
	"sync"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	mapData             [][]int
	mapWidth, mapHeight int
)

type ZBufferItem struct {
	Dist     float64
	HitX     float64
	HitY     float64
	IsHitOnX bool
	ItemID   int
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

	depth := 0.01
	dCosAngle := math.Cos(rayAngleRadians)
	dSinAngle := math.Sin(rayAngleRadians)
	deltaX := depth * dCosAngle
	deltaY := depth * dSinAngle
	mapX := int(rayX)
	mapY := int(rayY)

	totalDepth := depth
	isSelfNotAddedInThisReflection := false // ignore first pass, you are not able to see self without a mirror
	isCurrentObjectServedInThisTile := false
	for rayX >= 0 && rayX < float64(mapWidth) && rayY >= 0 && rayY < float64(mapHeight) {
		totalDepth += depth
		if totalDepth > 150 {
			return zbufferSlice
		}
		rayX += deltaX
		lastMaxX := mapX
		lastMaxY := mapY
		mapX = int(rayX)
		mapY = int(rayY) // needs to be recalculated in case of reflection

		if true {
			// track roof and floor first
			insideTileX := float64(rayX - math.Floor(rayX))
			insideTileY := float64(rayY + deltaY - math.Floor(rayY+deltaY)) // adjust Y for future movement
			newZBufferItem := ZBufferItem{totalDepth, insideTileX, insideTileY, true, 0}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
		}

		// Check for collision on X
		if mapData[mapY][mapX] == 1 {
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, true, 1}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			return zbufferSlice
		}
		// check of mirrors on X - left right
		if mapData[mapY][mapX] == 2 {
			rayY += deltaY // correct the ray
			deltaX = -deltaX
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, true, 2}
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
		if lastMaxX != mapX || lastMaxY != mapY { // only one object at one tile
			isCurrentObjectServedInThisTile = false
		}
		if mapData[mapY][mapX] == 1 {
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, false, 1}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			return zbufferSlice
		}

		// check of mirrors on Y
		if mapData[mapY][mapX] == 2 { //up down
			deltaY = -deltaY
			dSinAngle = -dSinAngle // sin 180 is 1/2 period
			newZBufferItem := ZBufferItem{totalDepth, rayX, rayY, false, 2}
			zbufferSlice = append(zbufferSlice, newZBufferItem)
			rayX += deltaX
			rayY += deltaY
			isSelfNotAddedInThisReflection = true
			continue
		}
		lenFromSelf := 100.0
		if totalDepth > 0.6 && isSelfNotAddedInThisReflection {
			lenFromSelf = getRadius(rayXOrigin, rayYOrigin, rayX, rayY)
		}
		mapObjectId := mapData[mapY][mapX]
		// check of other objects on the map
		if (mapObjectId > 2 && mapObjectId < 9 && isCurrentObjectServedInThisTile == false) || lenFromSelf <= 0.5 { // 9 is player
			// only execute if we are in the radius of the root of the squere
			var mapSquareCenterX float64
			var mapSquareCenterY float64
			if mapObjectId > 2 && mapObjectId < 9 {
				mapSquareCenterX = float64(mapX) + 0.5
				mapSquareCenterY = float64(mapY) + 0.5
			} else if lenFromSelf < 0.5 {
				mapSquareCenterX = rayXOrigin
				mapSquareCenterY = rayYOrigin
			}

			textureDistance := getDistanceOfPointOnLineByAngle(mapSquareCenterX, mapSquareCenterY, dCosAngle, dSinAngle, rayX, rayY)

			if textureDistance < 0.5 && textureDistance > -0.5 {
				if lenFromSelf <= 0.5 {
					newZBufferItem := ZBufferItem{totalDepth, textureDistance, 0, false, 9} // player
					zbufferSlice = append(zbufferSlice, newZBufferItem)
					isSelfNotAddedInThisReflection = false
					lenFromSelf = 100.0
				} else {
					newZBufferItem := ZBufferItem{totalDepth, textureDistance, 0, false, mapObjectId} // npc
					zbufferSlice = append(zbufferSlice, newZBufferItem)
					isCurrentObjectServedInThisTile = true
				}
			}
		}
	}
	newZBufferItem := ZBufferItem{5, 0, 0, false, 1} // default wall
	zbufferSlice = append(zbufferSlice, newZBufferItem)
	return zbufferSlice
}

const (
	screenWidth      = 1980
	screenHeight     = 1080
	screenHeightHalf = screenHeight / 2
	statusBarHeight  = screenHeight / 5
)

// Player properties
type Player struct {
	x, y  float64
	angle float64 // direction angle
}

// Define your result type based on castRay's return
type RayResult struct {
	index        int
	zBufferSlice []ZBufferItem // replace with actual type
}

func main() {
	raylib.InitWindow(screenWidth, screenHeight+statusBarHeight, "Raycasting in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("wall_all.png")
	defer raylib.UnloadTexture(wallTexture)
	wizardTexture := raylib.LoadTexture("raw_wizzard1.png")
	defer raylib.UnloadTexture(wizardTexture)
	warrior1Texture := raylib.LoadTexture("raw_warrior1.png")
	defer raylib.UnloadTexture(warrior1Texture)
	mirrorTexture := raylib.LoadTexture("raw_mirror.png")
	defer raylib.UnloadTexture(mirrorTexture)
	raw_warrior2 := raylib.LoadTexture("raw_warrior2.png")
	defer raylib.UnloadTexture(raw_warrior2)

	loadMap("map.txt")

	// Player
	player := Player{
		x:     3.0,
		y:     3.0,
		angle: 0, // looking straight ahead
	}

	// Camera settings
	fov := math.Pi / 2.0 // 90 degrees
	numRays := screenWidth / 2

	raylib.SetTargetFPS(60)

	for !raylib.WindowShouldClose() {
		var wg sync.WaitGroup
		resultChan := make(chan RayResult, numRays)
		rayResults := make([][]ZBufferItem, numRays)
		// Mouse
		deltaX := float64(raylib.GetMouseDelta().X)
		player.angle += deltaX * 0.01

		// Keyboard
		// Forward
		if raylib.IsKeyDown(raylib.KeyW) {
			player.x += 0.1 * math.Cos(player.angle)
			player.y += 0.1 * math.Sin(player.angle)
		}

		// Backward
		if raylib.IsKeyDown(raylib.KeyS) {
			player.x -= 0.1 * math.Cos(player.angle)
			player.y -= 0.1 * math.Sin(player.angle)
		}

		// Strafe left
		if raylib.IsKeyDown(raylib.KeyA) {
			player.x += 0.1 * math.Cos(player.angle-math.Pi/2)
			player.y += 0.1 * math.Sin(player.angle-math.Pi/2)
		}

		// Strafe right
		if raylib.IsKeyDown(raylib.KeyD) {
			player.x += 0.1 * math.Cos(player.angle+math.Pi/2)
			player.y += 0.1 * math.Sin(player.angle+math.Pi/2)
		}
		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.NewColor(50, 50, 50, 255))
		raylib.DrawRectangle(0, screenHeightHalf, screenWidth, screenHeight, raylib.NewColor(135, 135, 135, 255))

		for ray := 0; ray < numRays; ray++ {
			wg.Add(1)
			go func(ray int) {
				defer wg.Done()
				rayAngleRad := (float64(ray)/float64(numRays)-0.5)*fov + player.angle // ray angles in Rad
				//dist, hitX, hitY, isHitOnX := castRay(player.x, player.y, rayAngleRad)
				zbufferSlice := castRay(player.x, player.y, rayAngleRad)
				resultChan <- RayResult{index: ray, zBufferSlice: zbufferSlice}
			}(ray)
		}

		// Close channel once all goroutines are done
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect rayResults
		for rayResult := range resultChan {
			rayResults[rayResult.index] = rayResult.zBufferSlice
		}

		for ray, zBufferItems := range rayResults {

			//for rayResult := range resultChan {
			//	ray := rayResult.index
			// zBufferItems := rayResult.zBufferSlice
			lastWallHeight := float32(0.0)

			for i := len(zBufferItems) - 1; i >= 0; i-- {
				zbufferItem := zBufferItems[i]
				itemId := zbufferItem.ItemID
				itemDist := zbufferItem.Dist
				hitX := zbufferItem.HitX
				hitY := zbufferItem.HitY
				isHitOnX := zbufferItem.IsHitOnX
				// Calculate the angle difference
				rayAngleRad := (float64(ray)/float64(numRays)-0.5)*fov + player.angle
				angleDiff := rayAngleRad - player.angle

				// Perspective correction
				dist := itemDist * math.Cos(angleDiff)
				wallHeight := float32(screenHeight / (dist + 0.0001))

				// Texture coordinate
				if itemId == 0 {
					if lastWallHeight == 0.0 {
						lastWallHeight = wallHeight
						continue
					}

					var textureBright raylib.Texture2D

					textureBright = wallTexture

					var texX float32
					texX = float32(hitX) * float32(textureBright.Width) // map to texture width
					var texY float32
					texY = float32(hitY) * float32(textureBright.Height) // map to texture width

					// Calculate the rectangle to draw
					sliceWidthX := float32(textureBright.Width) / float32(numRays)
					//sliceHeightY := float32(textureBright.Height) / float32(lastWallHeight/2-wallHeight/2)

					// Source rectangle from texture
					srcRect := raylib.Rectangle{
						X:      texX,
						Y:      texY,
						Width:  sliceWidthX,
						Height: wallHeight/2 - lastWallHeight/2,
					}
					// Destination rectangle FLOOR
					destRectFloor := raylib.Rectangle{
						X:      float32(ray) * (screenWidth / float32(numRays)),
						Y:      screenHeightHalf + lastWallHeight/2,
						Width:  float32(screenWidth/numRays) + 1,
						Height: wallHeight/2 - lastWallHeight/2 + 1,
					}
					// Destination rectangle ROOF
					destRectRoof := raylib.Rectangle{
						X:      float32(ray) * (screenWidth / float32(numRays)),
						Y:      screenHeightHalf - wallHeight/2,
						Width:  float32(screenWidth/numRays) + 1,
						Height: wallHeight/2 - lastWallHeight/2 + 1,
					}
					// Destination rectangle FLOOR
					// Draw textured slice
					raylib.DrawTexturePro(wallTexture, srcRect, destRectFloor, raylib.Vector2{}, 0, raylib.LightGray)
					raylib.DrawTexturePro(wallTexture, srcRect, destRectRoof, raylib.Vector2{}, 0, raylib.DarkGray)
					lastWallHeight = wallHeight
				} else if itemId >= 1 && itemId <= 2 {
					var textureBright raylib.Texture2D
					var textureDark raylib.Texture2D
					if itemId == 1 {
						textureBright = wallTexture
						textureDark = wallTexture //wallTextureDark
					}
					if itemId == 2 {
						textureBright = mirrorTexture
						textureDark = mirrorTexture
					}

					var texX float32
					if isHitOnX {
						texX = float32(hitY - math.Floor(hitY))
					} else {
						texX = float32(hitX - math.Floor(hitX))
					}
					texX = texX * float32(textureBright.Width) // map to texture width

					// Calculate the rectangle to draw
					sliceWidth := float32(textureBright.Width) / float32(numRays)

					// Source rectangle from texture
					srcRect := raylib.Rectangle{
						X:      texX,
						Y:      0,
						Width:  sliceWidth,
						Height: float32(textureBright.Height),
					}

					// Destination rectangle
					destRect := raylib.Rectangle{
						X:      float32(ray) * (screenWidth / float32(numRays)),
						Y:      screenHeightHalf - wallHeight/2,
						Width:  float32(screenWidth / numRays),
						Height: wallHeight,
					}
					// Draw textured slice
					if isHitOnX {
						raylib.DrawTexturePro(textureBright, srcRect, destRect, raylib.Vector2{}, 0, raylib.Blue)
					} else {
						raylib.DrawTexturePro(textureDark, srcRect, destRect, raylib.Vector2{}, 0, raylib.DarkBlue)
					}
				} else {
					// sprites
					// self player
					var textureSprite raylib.Texture2D
					if itemId == 8 {
						textureSprite = wizardTexture
					}
					if itemId == 9 {
						textureSprite = warrior1Texture
					}
					if itemId == 7 {
						textureSprite = raw_warrior2
					}

					// hitx is -0.5 to 0.5 of the textureSprite

					texX := 0.5 - float32(hitX)                // map to textureSprite width
					texX = texX * float32(textureSprite.Width) // map to textureSprite width

					// Calculate the rectangle to draw
					sliceWidth := float32(textureSprite.Width) / float32(numRays)

					// Source rectangle from textureSprite
					srcRect := raylib.Rectangle{
						X:      texX,
						Y:      0,
						Width:  sliceWidth,
						Height: float32(textureSprite.Height),
					}

					// Destination rectangle
					destRect := raylib.Rectangle{
						X:      float32(ray) * (screenWidth / float32(numRays)),
						Y:      screenHeightHalf - wallHeight/2,
						Width:  float32(screenWidth / numRays),
						Height: wallHeight,
					}
					raylib.DrawTexturePro(textureSprite, srcRect, destRect, raylib.Vector2{}, 0, raylib.White)
				}
			}

		}
		// status bar overlay
		raylib.DrawRectangle(0, screenHeight, screenWidth, screenHeight+10, raylib.NewColor(0, 0, 128, 255))
		raylib.DrawRectangle(0, screenHeight+10, screenWidth, screenHeight+statusBarHeight, raylib.NewColor(0, 0, 255, 255))
		playerStatus := fmt.Sprintf("%.2f\n%.2f\n%.2f", player.x, player.y, player.angle)
		raylib.DrawText(playerStatus, 20, screenHeight+10, 65, raylib.White)

		playerStatusLegend := fmt.Sprintf("x\ny\n>")
		raylib.DrawText(playerStatusLegend, 200, screenHeight+10, 65, raylib.White)
		raylib.DrawFPS(screenWidth-90, screenHeight+10)

		raylib.EndDrawing()
	}
}
