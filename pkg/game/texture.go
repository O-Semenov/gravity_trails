package game

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

var fontFaceSource *text.GoTextFaceSource
var vehicleTextures = make(map[string]*ebiten.Image)

func init() {
	var err error
	fontFaceSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}

	// Генерация текстур кузовов при старте
	for _, preset := range VehiclePresets {
		vehicleTextures[preset.ID] = generateVehicleTexture(preset.ChassisWidth, preset.ChassisHeight, preset.ID)
	}
}

func generateVehicleTexture(w, h float64, presetID string) *ebiten.Image {
	img := ebiten.NewImage(int(w), int(h))
	fw := float32(w)
	fh := float32(h)

	var bodyColor, glassColor, lightColor color.RGBA
	switch presetID {
	case "buggy":
		bodyColor = color.RGBA{0, 160, 120, 255}
		glassColor = color.RGBA{0, 240, 255, 200}
		lightColor = color.RGBA{255, 220, 0, 255}
	case "sport":
		bodyColor = color.RGBA{220, 40, 40, 255}
		glassColor = color.RGBA{0, 255, 255, 220}
		lightColor = color.RGBA{255, 255, 255, 255}
	case "truck":
		bodyColor = color.RGBA{140, 80, 220, 255}
		glassColor = color.RGBA{0, 240, 255, 180}
		lightColor = color.RGBA{255, 150, 0, 255}
	default:
		bodyColor = color.RGBA{100, 100, 110, 255}
		glassColor = color.RGBA{180, 200, 220, 255}
		lightColor = color.RGBA{255, 255, 255, 255}
	}

	// Корпус
	vector.DrawFilledRect(img, 0, fh*0.4, fw, fh*0.6, bodyColor, true)
	// Кабина
	vector.DrawFilledRect(img, fw*0.25, 0, fw*0.45, fh*0.4, bodyColor, true)
	// Лобовое стекло
	vector.DrawFilledRect(img, fw*0.55, fh*0.08, fw*0.12, fh*0.32, glassColor, true)
	// Заднее стекло
	vector.DrawFilledRect(img, fw*0.28, fh*0.08, fw*0.1, fh*0.32, glassColor, true)
	// Передняя фара
	vector.DrawFilledRect(img, fw-6, fh*0.5, 6, fh*0.2, lightColor, true)
	// Задняя фара
	vector.DrawFilledRect(img, 0, fh*0.5, 4, fh*0.15, color.RGBA{255, 0, 0, 255}, true)
	// Неоновые обводки
	vector.StrokeRect(img, 0, fh*0.4, fw, fh*0.6, 1.5, color.RGBA{255, 255, 255, 100}, true)
	vector.StrokeRect(img, fw*0.25, 0, fw*0.45, fh*0.4, 1.5, color.RGBA{255, 255, 255, 100}, true)

	return img
}

func drawText(screen *ebiten.Image, txt string, x, y float64, size float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.LayoutOptions.LineSpacing = size * 1.2
	if clr != nil {
		op.ColorScale.ScaleWithColor(clr)
	}
	text.Draw(screen, txt, &text.GoTextFace{
		Source: fontFaceSource,
		Size:   size,
	}, op)
}
