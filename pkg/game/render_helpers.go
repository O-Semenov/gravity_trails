package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawTerrain(screen *ebiten.Image) {
	groundColor := color.RGBA{22, 27, 35, 255}
	surfaceColor := color.RGBA{0, 215, 120, 255}

	stepX := 4.0
	for x := 0.0; x < float64(g.screenWidth); x += stepX {
		worldX1 := x + g.camX
		worldX2 := x + stepX + g.camX

		y1 := g.world.GetHeight(worldX1)
		y2 := g.world.GetHeight(worldX2)

		vector.DrawFilledRect(screen, float32(x), float32(y1), float32(stepX), float32(float64(g.screenHeight)-y1), groundColor, true)
		vector.StrokeLine(screen, float32(x), float32(y1), float32(x+stepX), float32(y2), 3, surfaceColor, true)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	damage := g.getDamagePercent()
	velVec := g.vehicle.GetVelocity()
	speedKmh := velVec.Len() * 21.6
	if speedKmh < 0.2 {
		speedKmh = 0
	}

	uiX := float32(20)
	uiY := float32(25)
	uiWidth := float32(220)
	uiHeight := float32(18)

	vector.DrawFilledRect(screen, uiX, uiY, uiWidth, uiHeight, color.RGBA{45, 45, 55, 255}, true)
	dmgFactor := damage / 100.0
	var barColor color.RGBA
	if dmgFactor < 0.5 {
		barColor = color.RGBA{0, 215, 120, 255}
	} else if dmgFactor < 0.85 {
		barColor = color.RGBA{255, 200, 0, 255}
	} else {
		barColor = color.RGBA{255, 50, 50, 255}
	}
	vector.DrawFilledRect(screen, uiX, uiY, uiWidth*float32(dmgFactor), uiHeight, barColor, true)
	vector.StrokeRect(screen, uiX, uiY, uiWidth, uiHeight, 1.5, color.RGBA{80, 80, 95, 255}, true)

	drawText(screen, fmt.Sprintf("КОНСТРУКЦИЯ: %d%%", 100-int(damage)), float64(uiX), float64(uiY)-18, 12, nil)
	drawText(screen, fmt.Sprintf("СКОРОСТЬ:  %.0f км/ч", speedKmh), float64(uiX)+250, float64(uiY)-10, 12, nil)
	drawText(screen, fmt.Sprintf("ДИСТАНЦИЯ: %.1f м (РЕКОРД: %.1f м)", g.currentSessionMax, g.maxDistance), float64(uiX)+430, float64(uiY)-10, 12, nil)

	drawText(screen, "ESC: Меню | R: Сброс", 20, float64(g.screenHeight)-25, 11, color.RGBA{120, 120, 130, 255})
}
