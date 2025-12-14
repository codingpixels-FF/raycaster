package main

import (
	"bufio"
	"math"
	"os"
	_ "strconv"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	mapData             [][]int
	mapWidth, mapHeight int
)

func loadMap(filename string) {
	file, _ := os.Open(filename)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		row := make([]int, len(line))
		for i, ch := range line {
			if ch == '1' {
				row[i] = 1
			} else {
				row[i] = 0
			}
		}
		mapData = append(mapData, row)
	}
	mapHeight = len(mapData)
	mapWidth = len(mapData[0])
}

// castRay function returns distance to the wall
func castRay(rayX float64, rayY float64, rayAngleDegrees float64) (float64, float64, float64, bool) {

	depth := 0.01
	deltaX := depth * math.Cos(rayAngleDegrees)
	deltaY := depth * math.Sin(rayAngleDegrees)
	mapX := int(rayX)
	mapY := int(rayY)
	totalDepth := depth
	for rayX >= 0 && rayX < float64(mapWidth) && rayY >= 0 && rayY < float64(mapHeight) {
		totalDepth += depth
		rayX += deltaX
		mapX = int(rayX)
		// Check for collision on X
		if mapData[mapY][mapX] == 1 {
			return totalDepth, rayX, rayY, true
		}
		rayY += deltaY
		// check for collision on both (and assume it was Y)
		mapY = int(rayY)
		if mapData[mapY][mapX] == 1 {
			return totalDepth, rayX, rayY, false
		}
	}

	return 5, 0, 0, false
}

const (
	screenWidth      = 1980
	screenHeight     = 1080
	screenHeightHalf = screenHeight / 2
)

func main() {
	raylib.InitWindow(screenWidth, screenHeight, "Raycasting in Go")
	defer raylib.CloseWindow()

	wallTexture := raylib.LoadTexture("wall5.png") // or "wall.jpg"
	defer raylib.UnloadTexture(wallTexture)
	wallTextureDark := raylib.LoadTexture("wall5_dark.png") // or "wall.jpg"
	defer raylib.UnloadTexture(wallTexture)

	loadMap("map.txt")

	// Player
	posX, posY := 3.0, 3.0
	dir := 0.0 // looking straight ahead

	// Camera settings
	fov := 90.0
	numRays := screenWidth / 2

	raylib.SetTargetFPS(60)

	for !raylib.WindowShouldClose() {
		// Mouse
		deltaX := float64(raylib.GetMouseDelta().X)
		dir += deltaX * 0.01

		// Keyboard
		if raylib.IsKeyDown(raylib.KeyW) {
			posX += 0.1 * math.Cos(dir)
			posY += 0.1 * math.Sin(dir)
		}
		if raylib.IsKeyDown(raylib.KeyS) {
			posX -= 0.1 * math.Cos(dir)
			posY -= 0.1 * math.Sin(dir)
		}

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.NewColor(50, 50, 50, 255))
		raylib.DrawRectangle(0, screenHeightHalf, screenWidth, screenHeight, raylib.NewColor(135, 135, 135, 255))

		for ray := 0; ray < numRays; ray++ {
			rayAngleDegrees := (float64(ray)/float64(numRays)-0.5)*(fov*math.Pi/180) + dir
			dist, hitX, hitY, isHitOnX := castRay(posX, posY, rayAngleDegrees)
			wallHeight := float32(screenHeight / (dist + 0.0001))

			// Texture coordinate
			var texX float32
			if isHitOnX {
				texX = float32(hitY - math.Floor(hitY))
			} else {
				texX = float32(hitX - math.Floor(hitX))
			}

			texX = texX * float32(wallTexture.Width) // map to texture width

			// Calculate the rectangle to draw
			sliceWidth := float32(wallTexture.Width) / float32(numRays)

			// Source rectangle from texture
			srcRect := raylib.Rectangle{
				X:      texX,
				Y:      0,
				Width:  sliceWidth,
				Height: float32(wallTexture.Height),
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
				raylib.DrawTexturePro(wallTexture, srcRect, destRect, raylib.Vector2{}, 0, raylib.White)
			} else {
				raylib.DrawTexturePro(wallTextureDark, srcRect, destRect, raylib.Vector2{}, 0, raylib.White)
			}

		}

		raylib.EndDrawing()
	}
}
