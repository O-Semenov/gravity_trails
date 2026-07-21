package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawWarnings(screen *ebiten.Image) {
	if !g.isCrashed {
		preset := VehiclePresets[g.currentVehicleIndex]
		if g.crashedTimer > 0.0 {
			pulse := math.Sin(g.crashedTimer*15.0) > 0.0
			var txtColor color.RGBA
			if pulse {
				txtColor = color.RGBA{255, 50, 50, 255}
			} else {
				txtColor = color.RGBA{255, 200, 0, 255}
			}

			boxW := float32(420)
			boxH := float32(35)
			boxX := float32(g.screenWidth)/2 - boxW/2
			boxY := float32(80)
			vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{25, 10, 10, 200}, true)
			vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 1.5, txtColor, true)

			var msg string
			if preset.ID == "bike" {
				msg = fmt.Sprintf("МОТОЦИКЛ ПЕРЕВЕРНУТ! РАСКАЧАЙТЕ РАМУ [A / D]: %.1fс", math.Max(0.0, 1.5-g.crashedTimer))
			} else {
				msg = fmt.Sprintf("ПЕРЕВЕРНУТО! РАСКАЧАЙТЕ КУЗОВ [A / D]: %.1fс", math.Max(0.0, 1.5-g.crashedTimer))
			}
			drawText(screen, msg, float64(boxX)+15, float64(boxY)+10, 12, txtColor)
		} else if g.wheelLossTimer > 0.0 {
			boxW := float32(340)
			boxH := float32(35)
			boxX := float32(g.screenWidth)/2 - boxW/2
			boxY := float32(80)
			vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{25, 10, 10, 200}, true)
			vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 1.5, color.RGBA{255, 150, 0, 255}, true)

			var msg string
			if preset.ID == "bike" {
				msg = fmt.Sprintf("ПОТЕРЯ ОБОИХ КОЛЕС! КОНЕЦ ЧЕРЕЗ: %.1fс", math.Max(0.0, 2.0-g.wheelLossTimer))
			} else {
				msg = fmt.Sprintf("ПОТЕРЯ ОБОИХ КОЛЕС! КОНЕЦ ЧЕРЕЗ: %.1fс", math.Max(0.0, 2.0-g.wheelLossTimer))
			}
			drawText(screen, msg, float64(boxX)+15, float64(boxY)+10, 12, color.RGBA{255, 150, 0, 255})
		}
	}
}

func (g *Game) drawCrashOverlay(screen *ebiten.Image) {
	if g.isCrashed {
		vector.DrawFilledRect(screen, 0, 0, float32(g.screenWidth), float32(g.screenHeight), color.RGBA{80, 0, 0, 140}, true)

		preset := VehiclePresets[g.currentVehicleIndex]
		var msgCause string
		leftBroken := g.vehicle.IsWheelBroken(true)
		rightBroken := g.vehicle.IsWheelBroken(false)
		leftHeight := g.vehicle.Chassis[0].Pos.Dist(g.vehicle.Chassis[3].Pos)
		rightHeight := g.vehicle.Chassis[1].Pos.Dist(g.vehicle.Chassis[2].Pos)

		if leftHeight < 16.0 || rightHeight < 16.0 {
			if preset.ID == "bike" {
				msgCause = "Рама мотоцикла полностью разрушена!"
			} else {
				msgCause = "Кабина водителя полностью раздавлена!"
			}
		} else if leftBroken && rightBroken {
			if preset.ID == "bike" {
				msgCause = "Оторваны оба колеса мотоцикла!"
			} else {
				msgCause = "Оторваны все колеса автомобиля!"
			}
		} else {
			if preset.ID == "bike" {
				msgCause = "Мотоцикл перевернулся и упал!"
			} else {
				msgCause = "Машина перевернулась и загорелась!"
			}
		}

		var msgTitle string
		if preset.ID == "bike" {
			msgTitle = "=== ВНИМАНИЕ: АВАРИЯ ==="
		} else {
			msgTitle = "=== ВНИМАНИЕ: КРУШЕНИЕ ==="
		}
		msgBody2 := "Пройденная дистанция: %.1f метров"
		msgPrompt := "Нажмите клавишу [ R ], чтобы перезапустить заезд"

		boxW := float32(400)
		boxH := float32(140)
		boxX := float32(g.screenWidth)/2 - boxW/2
		boxY := float32(g.screenHeight)/2 - boxH/2
		vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{18, 18, 24, 245}, true)
		vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{255, 50, 50, 255}, true)

		drawText(screen, msgTitle, float64(boxX)+95, float64(boxY)+20, 13, color.RGBA{255, 50, 50, 255})
		drawText(screen, msgCause, float64(boxX)+40, float64(boxY)+50, 12, color.RGBA{240, 240, 245, 255})
		drawText(screen, fmt.Sprintf(msgBody2, g.currentSessionMax), float64(boxX)+75, float64(boxY)+75, 12, color.RGBA{200, 200, 210, 255})
		drawText(screen, msgPrompt, float64(boxX)+40, float64(boxY)+105, 12, color.RGBA{255, 200, 0, 255})
	}
}
