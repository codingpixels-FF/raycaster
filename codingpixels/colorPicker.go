package codingpixels

import (
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type PixelColor struct {
	R uint8
	G uint8
	B uint8
}

type ColorPicker struct {
	CurrentPixelColor       PixelColor
	NumColors               uint8
	SingleColorRectSizeX    int32
	SingleColorRectSizeY    int32
	SingleColorPaddingX     int32
	SingleColorPaddingY     int32
	StartPositionX          int32
	StartPositionY          int32
	ColorIncrement          uint8
	ColorPickerEndPositionX int32
	ColorPickerEndPositionY int32
}

func NewColorPicker(numColors uint8, singleColorRectSizeX int32, singleColorRectSizeY int32, singleColorPaddingX int32, singleColorPaddingY int32, startPositionX int32, startPositionY int32, colorIncrement uint8) *ColorPicker {
	if numColors > 254 {
		panic("Number of colors must be less than 255")
	}
	colorPicker := &ColorPicker{CurrentPixelColor: PixelColor{
		R: 0,
		G: 0,
		B: 0,
	}, NumColors: numColors, SingleColorRectSizeX: singleColorRectSizeX, SingleColorRectSizeY: singleColorRectSizeY, SingleColorPaddingX: singleColorPaddingX, SingleColorPaddingY: singleColorPaddingY, StartPositionX: startPositionX, StartPositionY: startPositionY, ColorIncrement: colorIncrement}
	colorPicker.calculateFullDimensions()
	return colorPicker
}

func (colorPicker *ColorPicker) Render() {
	columnCount := int32(0)
	singleColorSizeWithPaddingX := colorPicker.SingleColorRectSizeX + colorPicker.SingleColorPaddingX
	singleColorSizeWithPaddingY := colorPicker.SingleColorRectSizeY + colorPicker.SingleColorPaddingY

	// Background
	raylib.DrawRectangle(
		colorPicker.StartPositionX,
		colorPicker.StartPositionY,
		colorPicker.ColorPickerEndPositionX-colorPicker.StartPositionX+colorPicker.SingleColorPaddingX,
		colorPicker.ColorPickerEndPositionY-colorPicker.StartPositionY,
		raylib.NewColor(30, 30, 30, 255),
	)
	// RGB picker
	for colorNum := uint8(0); colorNum < colorPicker.NumColors; colorNum = colorNum + colorPicker.ColorIncrement {
		newColor := raylib.NewColor(colorNum, colorPicker.CurrentPixelColor.G, colorPicker.CurrentPixelColor.B, 255)
		// Red color
		if colorNum == colorPicker.CurrentPixelColor.R {
			newColor = raylib.NewColor(255, 255, 255, 255)
			colorValue := int(colorNum) + int(colorPicker.CurrentPixelColor.G) + int(colorPicker.CurrentPixelColor.B)
			if colorValue > 128*3 {
				newColor = raylib.NewColor(0, 0, 0, 255)
			}
		}
		raylib.DrawRectangle(
			colorPicker.StartPositionX+columnCount*(singleColorSizeWithPaddingX),
			colorPicker.StartPositionY+(singleColorSizeWithPaddingY)*0,
			colorPicker.SingleColorRectSizeX,
			colorPicker.SingleColorRectSizeY,
			newColor,
		)

		// Green color
		newColor = raylib.NewColor(colorPicker.CurrentPixelColor.R, colorNum, colorPicker.CurrentPixelColor.B, 255)
		if colorNum == colorPicker.CurrentPixelColor.G {
			newColor = raylib.NewColor(255, 255, 255, 255)
			colorValue := int(colorPicker.CurrentPixelColor.R) + int(colorNum) + int(colorPicker.CurrentPixelColor.B)
			if colorValue > 128*3 {
				newColor = raylib.NewColor(0, 0, 0, 255)
			}
		}
		raylib.DrawRectangle(
			colorPicker.StartPositionX+columnCount*singleColorSizeWithPaddingX,
			colorPicker.StartPositionY+singleColorSizeWithPaddingY*1,
			colorPicker.SingleColorRectSizeX,
			colorPicker.SingleColorRectSizeY,
			newColor,
		)

		// Blue color
		newColor = raylib.NewColor(colorPicker.CurrentPixelColor.R, colorPicker.CurrentPixelColor.G, colorNum, 255)
		if colorNum == colorPicker.CurrentPixelColor.B {
			newColor = raylib.NewColor(255, 255, 255, 255)
			colorValue := int(colorPicker.CurrentPixelColor.R) + int(colorPicker.CurrentPixelColor.G) + int(colorNum)
			if colorValue > 128*3 {
				newColor = raylib.NewColor(0, 0, 0, 255)
			}
		}
		raylib.DrawRectangle(
			colorPicker.StartPositionX+columnCount*singleColorSizeWithPaddingX,
			colorPicker.StartPositionY+singleColorSizeWithPaddingY*2,
			colorPicker.SingleColorRectSizeX,
			colorPicker.SingleColorRectSizeY,
			newColor,
		)
		columnCount++
	}

	// update dimensions
	colorPicker.ColorPickerEndPositionX = colorPicker.StartPositionX + columnCount*(singleColorSizeWithPaddingX) + colorPicker.SingleColorPaddingX
	colorPicker.ColorPickerEndPositionY = colorPicker.StartPositionY + singleColorSizeWithPaddingY*2
}

func (colorPicker *ColorPicker) UpdateColors(mouseAbsX int32, mouseAbsY int32) {
	singleColorSizeWithPaddingX := colorPicker.SingleColorRectSizeX + colorPicker.SingleColorPaddingX
	singleColorSizeWithPaddingY := colorPicker.SingleColorRectSizeY + colorPicker.SingleColorPaddingY
	startRY := colorPicker.StartPositionY
	startGY := colorPicker.StartPositionY + singleColorSizeWithPaddingY*1
	startBY := colorPicker.StartPositionY + singleColorSizeWithPaddingY*2
	if mouseAbsY > startRY && mouseAbsY < startRY+colorPicker.SingleColorRectSizeY {
		colorPicker.CurrentPixelColor.R = uint8((mouseAbsX-colorPicker.StartPositionX)/singleColorSizeWithPaddingX) * colorPicker.ColorIncrement
	}
	if mouseAbsY > startGY && mouseAbsY < startGY+colorPicker.SingleColorRectSizeY {
		colorPicker.CurrentPixelColor.G = uint8((mouseAbsX-colorPicker.StartPositionX)/singleColorSizeWithPaddingX) * colorPicker.ColorIncrement
	}
	if mouseAbsY > startBY && mouseAbsY < startBY+colorPicker.SingleColorRectSizeY {
		colorPicker.CurrentPixelColor.B = uint8((mouseAbsX-colorPicker.StartPositionX)/singleColorSizeWithPaddingX) * colorPicker.ColorIncrement
	}

}

func (colorPicker *ColorPicker) GetPixelColor() PixelColor {
	return colorPicker.CurrentPixelColor
}

func (colorPicker *ColorPicker) calculateFullDimensions() {
	columnCount := int32(0)
	singleColorSizeWithPaddingX := colorPicker.SingleColorRectSizeX + colorPicker.SingleColorPaddingX
	singleColorSizeWithPaddingY := colorPicker.SingleColorRectSizeY + colorPicker.SingleColorPaddingY
	for colorNum := uint8(0); colorNum < colorPicker.NumColors; colorNum = colorNum + colorPicker.ColorIncrement {
		columnCount++
	}
	colorPicker.ColorPickerEndPositionX = colorPicker.StartPositionX + columnCount*(singleColorSizeWithPaddingX) + colorPicker.SingleColorPaddingX
	colorPicker.ColorPickerEndPositionY = colorPicker.StartPositionY + singleColorSizeWithPaddingY*2
}
