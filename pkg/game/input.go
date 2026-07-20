package game

import (
	"gravitytrails/pkg/physics"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) UpdateGame() error {
	// ESC возвращает в меню
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.saveHighScore()
		g.state = StateMenu
		return nil
	}

	// Сброс на клавишу R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.saveHighScore()
		g.initSandbox()
		g.draggedNode = nil
		g.camX = 0
	}

	// Сбрасываем трение колес на свободное качение (инерция)
	g.vehicle.LeftWheel.Friction = 0.01
	g.vehicle.RightWheel.Friction = 0.01

	// Проверяем контакт колес с землей
	leftOnGround := g.vehicle.LeftWheel.OnGround && !g.vehicle.IsWheelBroken(true)
	rightOnGround := g.vehicle.RightWheel.OnGround && !g.vehicle.IsWheelBroken(false)
	onGround := leftOnGround || rightOnGround

	// Получаем текущую горизонтальную скорость автомобиля
	currentVelocity := g.vehicle.GetVelocity()

	// Управление движением и наклоном
	if !g.isCrashed {
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			if onGround {
				if currentVelocity.X < -0.5 {
					// Тормозим, если двигаемся назад
					g.vehicle.LeftWheel.Friction = 0.75
					g.vehicle.RightWheel.Friction = 0.75
				} else {
					// Газ вперед
					g.vehicle.ApplyDriveTorque(g.world, g.vehicle.MaxTorque)
				}
			} else {
				// Наклон корпуса против часовой стрелки
				g.vehicle.ApplyAirControl(-g.vehicle.AirTorque)
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			if onGround {
				if currentVelocity.X > 0.5 {
					// Тормозим, если двигаемся вперед
					g.vehicle.LeftWheel.Friction = 0.75
					g.vehicle.RightWheel.Friction = 0.75
				} else {
					// Задний ход
					g.vehicle.ApplyDriveTorque(g.world, -0.6*g.vehicle.MaxTorque)
				}
			} else {
				// Наклон корпуса по часовой стрелке
				g.vehicle.ApplyAirControl(g.vehicle.AirTorque)
			}
		}
	}

	// Захват узла мышкой (с учетом позиции камеры)
	mx, my := ebiten.CursorPosition()
	cursorWorldPos := physics.Vector2{X: float64(mx) + g.camX, Y: float64(my)}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.draggedNode == nil {
			var bestNode *physics.Node
			bestDist := 25.0
			for _, node := range g.world.Nodes {
				dist := node.Pos.Dist(cursorWorldPos)
				if dist < bestDist {
					bestDist = dist
					bestNode = node
				}
			}
			g.draggedNode = bestNode
		}

		if g.draggedNode != nil {
			g.draggedNode.Pos = cursorWorldPos
			g.draggedNode.OldPos = cursorWorldPos
		}
	} else {
		g.draggedNode = nil
	}

	// Шаг физического симулятора (dt = 1/60)
	dt := 1.0 / 60.0
	g.world.Update(dt)

	// Проверяем состояние аварии
	g.checkCrashedState(dt)

	// Камера следует за центром масс
	if len(g.world.Nodes) > 0 {
		avgX := g.vehicle.GetCenterOfMass().X
		targetCamX := avgX - float64(g.screenWidth)/2
		g.camX += (targetCamX - g.camX) * 0.08

		currentDist := avgX / 10.0
		if currentDist < 0 {
			currentDist = 0
		}
		if currentDist > g.currentSessionMax {
			g.currentSessionMax = currentDist
		}
		if g.currentSessionMax > g.maxDistance {
			g.maxDistance = g.currentSessionMax
			activeMap := GameMaps[g.currentMapIndex]
			g.allHighScores[activeMap.ID] = g.maxDistance
		}
	}

	return nil
}
