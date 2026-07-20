package game

import (
	"image/color"
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) DrawGame(screen *ebiten.Image) {
	// Премиальный темно-синий цвет неба
	screen.Fill(color.RGBA{15, 16, 22, 255})

	// 1. Отрисовка ландшафта
	g.drawTerrain(screen)

	// 1.5 Отрисовка деформируемого спрайта кузова
	g.drawDeformableSprite(screen)

	// 2. Отрисовка упругих связей (Beams), исключая балки кабины, перекрытые текстурой
	g.drawBeams(screen)

	// 3. Отрисовка узлов и колес
	g.drawNodesAndWheels(screen)

	// 3.5 Отрисовка спойлера и бампера
	g.drawDetachableParts(screen)

	// 4. Отрисовка HUD
	g.drawHUD(screen)

	// 5. Вывод предупреждений
	g.drawWarnings(screen)

	// 6. Оверлей при аварии
	g.drawCrashOverlay(screen)
}
