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
func castRay(px float64, py float64, rayAngleDegrees float64, maxDepth float64) (float64, float64, float64) {
	for depth := 0.0; depth < maxDepth; depth += 0.01 {
		positionX := px + depth*math.Cos(rayAngleDegrees)
		positionY := py + depth*math.Sin(rayAngleDegrees)
		mapX := int(positionX)
		mapY := int(positionY)
		// Is ray inside the map?
		if mapX >= 0 && mapX < mapWidth && mapY >= 0 && mapY < mapHeight {
			// Was a wall hit?
			if mapData[mapY][mapX] == 1 {
				return depth, positionX, positionY
			}
		} else {
			return maxDepth, positionX, positionY
		}
	}
	return maxDepth, 0, 0
}

const (
	screenWidth      = 1980
	screenHeight     = 1080
	screenHeightHalf = screenHeight / 2
)

func main() {
	raylib.InitWindow(screenWidth, screenHeight, "Raycasting in Go")
	defer raylib.CloseWindow()

	loadMap("map.txt")

	// Player
	posX, posY := 3.0, 3.0
	dir := 0.0 // looking straight ahead

	// Camera settings
	fov := 90.0
	numRays := screenWidth
	maxDepth := 20.0

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
		raylib.ClearBackground(raylib.RayWhite)

		for ray := 0; ray < numRays; ray++ {
			rayAngleDegrees := (float64(ray)/float64(numRays)-0.5)*(fov*math.Pi/180) + dir
			dist, _, _ := castRay(posX, posY, rayAngleDegrees, maxDepth)
			wallHeight := float32(screenHeight / (dist + 0.0001))
			lineX := float32(ray) * (screenWidth / float32(numRays))
			raylib.DrawLine(int32(lineX), screenHeightHalf-int32(wallHeight/2), int32(lineX), screenHeightHalf+int32(wallHeight/2), raylib.Black)
		}

		raylib.EndDrawing()
	}
}
