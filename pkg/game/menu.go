package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) selectMap(index int) {
	if index < 0 || index >= len(GameMaps) {
		return
	}
	if g.currentMapIndex == index {
		return
	}
	g.saveHighScore()
	g.currentMapIndex = index
	activeMap := GameMaps[g.currentMapIndex]
	g.maxDistance = g.allHighScores[activeMap.ID]

	g.initSandbox()
	g.draggedNode = nil
	g.camX = 0
}

func (g *Game) UpdateMenu() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	mx, my := ebiten.CursorPosition()

	// Сбрасываем ховеры
	g.hoveredVehicleIndex = -1
	g.hoveredMapIndex = -1

	// Проверяем наведение на машины
	if mx >= 80 && mx <= 360 {
		for i := 0; i < len(VehiclePresets); i++ {
			by := 175 + i*45
			if my >= by && my <= by+36 {
				g.hoveredVehicleIndex = i
				break
			}
		}
	}

	// Проверяем наведение на карты
	if mx >= 440 && mx <= 720 {
		for i := 0; i < len(GameMaps); i++ {
			by := 175 + i*45
			if my >= by && my <= by+36 {
				g.hoveredMapIndex = i
				break
			}
		}
	}

	// Обработка кликов
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.hoveredVehicleIndex != -1 {
			g.currentVehicleIndex = g.hoveredVehicleIndex
		}
		if g.hoveredMapIndex != -1 {
			g.currentMapIndex = g.hoveredMapIndex
		}
		// Клик по кнопке СТАРТ
		if mx >= 290 && mx <= 510 && my >= 515 && my <= 560 {
			g.state = StateGame
			g.initSandbox()
			g.camX = 0
			activeMap := GameMaps[g.currentMapIndex]
			g.maxDistance = g.allHighScores[activeMap.ID]
		}
	}

	// Быстрый запуск с клавиатуры
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StateGame
		g.initSandbox()
		g.camX = 0
		activeMap := GameMaps[g.currentMapIndex]
		g.maxDistance = g.allHighScores[activeMap.ID]
	}

	return nil
}

