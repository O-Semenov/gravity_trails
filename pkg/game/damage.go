package game

import (
	"math"
)

// checkCrashedState проверяет состояние автомобиля и управляет таймерами спасения
func (g *Game) checkCrashedState(dt float64) {
	if g.isCrashed {
		return
	}

	// 1. Мгновенная смерть от сдавливания кабины (смятие по высоте на 60% и более)
	if len(g.vehicle.Chassis) >= 4 {
		leftHeight := g.vehicle.Chassis[0].Pos.Dist(g.vehicle.Chassis[3].Pos)
		rightHeight := g.vehicle.Chassis[1].Pos.Dist(g.vehicle.Chassis[2].Pos)
		if leftHeight < 16.0 || rightHeight < 16.0 {
			g.isCrashed = true
			g.saveHighScore()
			return
		}
	} else if len(g.vehicle.Chassis) == 3 {
		// Для треугольной рамы проверяем критическую деформацию сторон треугольника
		sideA := g.vehicle.Chassis[0].Pos.Dist(g.vehicle.Chassis[1].Pos)
		sideB := g.vehicle.Chassis[1].Pos.Dist(g.vehicle.Chassis[2].Pos)
		sideC := g.vehicle.Chassis[2].Pos.Dist(g.vehicle.Chassis[0].Pos)
		if sideA < 15.0 || sideB < 15.0 || sideC < 15.0 {
			g.isCrashed = true
			g.saveHighScore()
			return
		}
	}

	// 2. Мягкая смерть при перевороте на крышу (таймер 1.5 секунды)
	roofTouched := g.vehicle.Chassis[0].OnGround || g.vehicle.Chassis[1].OnGround
	leftWheelTouched := g.vehicle.LeftWheel.OnGround && !g.vehicle.IsWheelBroken(true)
	rightWheelTouched := g.vehicle.RightWheel.OnGround && !g.vehicle.IsWheelBroken(false)

	// Если крыша на земле, а колеса НЕТ - запускаем таймер
	if roofTouched && !leftWheelTouched && !rightWheelTouched {
		g.crashedTimer += dt
		if g.crashedTimer >= 1.5 {
			g.isCrashed = true
			g.saveHighScore()
			return
		}
	} else {
		// Сбрасываем таймер, если встали на колеса
		g.crashedTimer = 0.0
	}

	// 3. Смерть при отрыве обоих колес (таймер 2.0 секунды)
	leftBroken := g.vehicle.IsWheelBroken(true)
	rightBroken := g.vehicle.IsWheelBroken(false)
	if leftBroken && rightBroken {
		g.wheelLossTimer += dt
		if g.wheelLossTimer >= 2.0 {
			g.isCrashed = true
			g.saveHighScore()
			return
		}
	} else {
		g.wheelLossTimer = 0.0
	}
}

// getDamagePercent вычисляет процент повреждений автомобиля
func (g *Game) getDamagePercent() float64 {
	if g.isCrashed {
		return 100.0
	}

	totalDamage := 0.0
	brokenCount := 0
	chassisBeamsCount := 0

	for _, b := range g.vehicle.Beams {
		// Оцениваем прочность только деталей кузова и подвески (без межосевого прутка)
		if b.NodeA.IsWheel && b.NodeB.IsWheel {
			continue
		}
		chassisBeamsCount++

		if b.IsBroken {
			brokenCount++
		} else {
			// Насколько деформация превышает предел упругости
			strain := math.Abs(b.Stress)
			if strain > b.LimitPlastic {
				// Относительная близость к разрыву
				p := (strain - b.LimitPlastic) / (b.LimitFracture - b.LimitPlastic)
				if p > totalDamage {
					totalDamage = p
				}
			}
		}
	}

	// Вычисляем процент сломанных деталей рамы
	brokenPercent := (float64(brokenCount) / float64(chassisBeamsCount)) * 100.0
	strainPercent := totalDamage * 100.0

	if brokenPercent > strainPercent {
		return brokenPercent
	}
	return strainPercent
}
