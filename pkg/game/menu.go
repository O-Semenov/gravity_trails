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
		if my >= 185 && my <= 221 {
			g.hoveredVehicleIndex = 0
		} else if my >= 235 && my <= 271 {
			g.hoveredVehicleIndex = 1
		} else if my >= 285 && my <= 321 {
			g.hoveredVehicleIndex = 2
		}
	}

	// Проверяем наведение на карты
	if mx >= 440 && mx <= 720 {
		if my >= 185 && my <= 221 {
			g.hoveredMapIndex = 0
		} else if my >= 235 && my <= 271 {
			g.hoveredMapIndex = 1
		} else if my >= 285 && my <= 321 {
			g.hoveredMapIndex = 2
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

