package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) DrawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 11, 15, 255})
	mx, my := ebiten.CursorPosition()

	// Заголовок
	titleFace := &text.GoTextFace{Source: fontFaceSource, Size: 32}
	tw, _ := text.Measure("GRAVITY TRAILS", titleFace, 0)
	titleX := (800.0 - tw) / 2.0
	drawText(screen, "GRAVITY TRAILS", titleX, 45, 32, color.RGBA{0, 255, 230, 255})

	// Подзаголовок
	subFace := &text.GoTextFace{Source: fontFaceSource, Size: 12}
	stw, _ := text.Measure("Симулятор езды с мягкой физикой кузова", subFace, 0)
	subtitleX := (800.0 - stw) / 2.0
	drawText(screen, "Симулятор езды с мягкой физикой кузова", subtitleX, 90, 12, color.RGBA{140, 150, 160, 255})

	// Разделительная линия
	vector.StrokeLine(screen, 100, 115, 700, 115, 1.0, color.RGBA{0, 255, 230, 80}, true)

	btnFace := &text.GoTextFace{
		Source: fontFaceSource,
		Size:   12,
	}

	// 1. КОЛОНКА МАШИН
	drawText(screen, "1. ВЫБЕРИТЕ МАШИНУ", 80, 150, 14, color.RGBA{255, 215, 0, 255})

	vehicleX := float32(80)
	vehicleW := float32(280)
	vehicleH := float32(36)
	vehicleYStart := float32(185)

	for i, vp := range VehiclePresets {
		by := vehicleYStart + float32(i)*50.0
		isActive := g.currentVehicleIndex == i
		isHovered := g.hoveredVehicleIndex == i

		var fillColor, borderColor color.RGBA
		if isActive {
			fillColor = color.RGBA{15, 45, 45, 255}
			borderColor = color.RGBA{0, 255, 230, 255}
		} else if isHovered {
			fillColor = color.RGBA{45, 45, 55, 255}
			borderColor = color.RGBA{120, 120, 140, 255}
		} else {
			fillColor = color.RGBA{25, 25, 30, 255}
			borderColor = color.RGBA{70, 70, 80, 255}
		}

		vector.DrawFilledRect(screen, vehicleX, by, vehicleW, vehicleH, fillColor, true)
		vector.StrokeRect(screen, vehicleX, by, vehicleW, vehicleH, 1.5, borderColor, true)

		txtW, txtH := text.Measure(vp.Name, btnFace, 0)
		tx := vehicleX + (vehicleW-float32(txtW))/2.0
		ty := by + (vehicleH-float32(txtH))/2.0

		txtColor := color.RGBA{200, 200, 210, 255}
		if isActive {
			txtColor = color.RGBA{0, 255, 230, 255}
		}
		drawText(screen, vp.Name, float64(tx), float64(ty), 12, txtColor)
	}

	// Описание выбранной/наведенной машины
	descVehicle := VehiclePresets[g.currentVehicleIndex]
	if g.hoveredVehicleIndex != -1 {
		descVehicle = VehiclePresets[g.hoveredVehicleIndex]
	}
	drawText(screen, descVehicle.Description, 80, 345, 11, color.RGBA{170, 175, 185, 255})

	// 2. КОЛОНКА КАРТ
	drawText(screen, "2. ВЫБЕРИТЕ КАРТУ", 440, 150, 14, color.RGBA{255, 215, 0, 255})

	mapX := float32(440)
	mapW := float32(280)
	mapH := float32(36)
	mapYStart := float32(185)

	for i, mp := range GameMaps {
		by := mapYStart + float32(i)*50.0
		isActive := g.currentMapIndex == i
		isHovered := g.hoveredMapIndex == i

		var fillColor, borderColor color.RGBA
		if isActive {
			fillColor = color.RGBA{15, 45, 45, 255}
			borderColor = color.RGBA{0, 255, 230, 255}
		} else if isHovered {
			fillColor = color.RGBA{45, 45, 55, 255}
			borderColor = color.RGBA{120, 120, 140, 255}
		} else {
			fillColor = color.RGBA{25, 25, 30, 255}
			borderColor = color.RGBA{70, 70, 80, 255}
		}

		vector.DrawFilledRect(screen, mapX, by, mapW, mapH, fillColor, true)
		vector.StrokeRect(screen, mapX, by, mapW, mapH, 1.5, borderColor, true)

		record := g.allHighScores[mp.ID]
		label := mp.Name
		if record > 0.0 {
			label = fmt.Sprintf("%s (Рекорд: %.1fм)", mp.Name, record)
		}

		txtW, txtH := text.Measure(label, btnFace, 0)
		tx := mapX + (mapW-float32(txtW))/2.0
		ty := by + (mapH-float32(txtH))/2.0

		txtColor := color.RGBA{200, 200, 210, 255}
		if isActive {
			txtColor = color.RGBA{0, 255, 230, 255}
		}
		drawText(screen, label, float64(tx), float64(ty), 12, txtColor)
	}

	// Описание выбранной/наведенной карты
	descMap := GameMaps[g.currentMapIndex]
	if g.hoveredMapIndex != -1 {
		descMap = GameMaps[g.hoveredMapIndex]
	}
	drawText(screen, descMap.Description, 440, 345, 11, color.RGBA{170, 175, 185, 255})

	// 3. БЛОК УПРАВЛЕНИЯ И СПРАВКИ
	vector.DrawFilledRect(screen, 80, 420, 640, 70, color.RGBA{18, 19, 24, 255}, true)
	vector.StrokeRect(screen, 80, 420, 640, 70, 1.0, color.RGBA{40, 42, 50, 255}, true)

	controlText := "Управление:\n" +
		" - Стрелки / WASD: Движение на земле | Наклон и раскачка корпуса в воздухе\n" +
		" - Левая кнопка мыши (ЛКМ): Перетаскивание и швыряние узлов кабины\n" +
		" - Клавиша R: Сброс машины во время заезда | ESC: Вернуться в это меню"
	drawText(screen, controlText, 95, 430, 11, color.RGBA{150, 155, 165, 255})

	// 4. КНОПКА СТАРТА
	startX := float32(290)
	startY := float32(515)
	startW := float32(220)
	startH := float32(45)

	isHoveredStart := float32(mx) >= startX && float32(mx) <= startX+startW && float32(my) >= startY && float32(my) <= startY+startH

	var startFill, startBorder color.RGBA
	if isHoveredStart {
		startFill = color.RGBA{0, 60, 60, 255}
		startBorder = color.RGBA{0, 255, 230, 255}
	} else {
		startFill = color.RGBA{0, 35, 35, 255}
		startBorder = color.RGBA{0, 180, 160, 255}
	}

	vector.DrawFilledRect(screen, startX, startY, startW, startH, startFill, true)
	vector.StrokeRect(screen, startX, startY, startW, startH, 2.0, startBorder, true)

	startFace := &text.GoTextFace{Source: fontFaceSource, Size: 15}
	btnTxt := "ПОЕХАЛИ! (Enter)"
	btnTxtW, btnTxtH := text.Measure(btnTxt, startFace, 0)
	btnTx := startX + (startW-float32(btnTxtW))/2.0
	btnTy := startY + (startH-float32(btnTxtH))/2.0

	drawText(screen, btnTxt, float64(btnTx), float64(btnTy), 15, color.RGBA{0, 255, 230, 255})
}
