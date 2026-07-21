package game

import (
	"image/color"
	"math"

	"gravitytrails/pkg/physics"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawDeformableSprite(screen *ebiten.Image) {
	preset := VehiclePresets[g.currentVehicleIndex]
	texture := vehicleTextures[preset.ID]
	if texture == nil {
		return
	}

	v := g.vehicle
	wVal := float32(texture.Bounds().Dx())
	hVal := float32(texture.Bounds().Dy())

	// Направление по горизонтали (от левого колеса к правому)
	dx := v.RightWheel.Pos.X - v.LeftWheel.Pos.X
	dy := v.RightWheel.Pos.Y - v.LeftWheel.Pos.Y
	wheelBase := math.Sqrt(dx*dx + dy*dy)

	var ux, uy float64
	if wheelBase > 0 {
		ux = dx / wheelBase
		uy = dy / wheelBase
	} else {
		ux = 1.0
		uy = 0.0
	}
	uVec := physics.Vector2{X: ux, Y: uy}

	// Направление вверх (перпендикуляр к uVec, в экранных координатах Y идет вниз)
	nVec := physics.Vector2{X: uy, Y: -ux}

	// Рассчитываем динамические масштабные коэффициенты
	scaleX := 150.0
	if spanDiff := preset.UvRightWheelX - preset.UvLeftWheelX; spanDiff > 0 {
		scaleX = wheelBase / spanDiff
	}

	// Рассчитываем вертикальный размер на основе расстояния между крышей и колесами
	avgChassisY := ((v.Chassis[0].Pos.X+v.Chassis[1].Pos.X)*nVec.X + (v.Chassis[0].Pos.Y+v.Chassis[1].Pos.Y)*nVec.Y) / 2.0
	avgWheelY := ((v.LeftWheel.Pos.X+v.RightWheel.Pos.X)*nVec.X + (v.LeftWheel.Pos.Y+v.RightWheel.Pos.Y)*nVec.Y) / 2.0
	heightDiff := avgChassisY - avgWheelY
	if heightDiff < 5.0 {
		heightDiff = 5.0 // Защита от деления на ноль/сплющивания
	}

	scaleY := 60.0
	if heightUvDiff := preset.UvLeftWheelY - preset.UvTopLeftY; heightUvDiff > 0 {
		scaleY = heightDiff / heightUvDiff
	}

	// Вычисляем координаты 4-х внешних углов изображения на экране на основе относительных офсетов
	// 0 - Top-Left (U=0, V=0)
	dst0 := v.Chassis[0].Pos.
		Add(uVec.Mul(-preset.UvTopLeftX * scaleX)).
		Add(nVec.Mul(preset.UvTopLeftY * scaleY))

	// 1 - Top-Right (U=1, V=0)
	dst1 := v.Chassis[1].Pos.
		Add(uVec.Mul((1.0 - preset.UvTopRightX) * scaleX)).
		Add(nVec.Mul(preset.UvTopRightY * scaleY))

	// 2 - Bottom-Right (U=1, V=1)
	dst2 := v.RightWheel.Pos.
		Add(uVec.Mul((1.0 - preset.UvRightWheelX) * scaleX)).
		Add(nVec.Mul(-(1.0 - preset.UvRightWheelY) * scaleY))

	// 3 - Bottom-Left (U=0, V=1)
	dst3 := v.LeftWheel.Pos.
		Add(uVec.Mul(-preset.UvLeftWheelX * scaleX)).
		Add(nVec.Mul(-(1.0 - preset.UvLeftWheelY) * scaleY))

	// 4 - Center (U=0.5, V=0.5)
	var dst4 physics.Vector2
	if v.Center != nil {
		dst4 = v.Center.Pos
	} else {
		dst4 = dst0.Add(dst1).Add(dst2).Add(dst3).Div(4.0)
	}

	destPositions := []physics.Vector2{dst0, dst1, dst2, dst3, dst4}

	// Полные UV координаты от 0 до wVal/hVal
	uvX := []float32{0, wVal, wVal, 0, wVal * 0.5}
	uvY := []float32{0, 0, hVal, hVal, hVal * 0.5}

	vertices := make([]ebiten.Vertex, 5)
	for i := 0; i < 5; i++ {
		pos := destPositions[i]
		vertices[i] = ebiten.Vertex{
			DstX:   float32(pos.X - g.camX),
			DstY:   float32(pos.Y),
			SrcX:   uvX[i],
			SrcY:   uvY[i],
			ColorR: 1.0, ColorG: 1.0, ColorB: 1.0, ColorA: 1.0,
		}
	}

	var indices []uint16
	// Проверяем целостность физических связей перед рендером каждого треугольника
	if v.Center != nil {
		if isBeamIntact(v, v.Chassis[0], v.Chassis[1]) && isBeamIntact(v, v.Chassis[0], v.Center) && isBeamIntact(v, v.Chassis[1], v.Center) {
			indices = append(indices, 0, 1, 4)
		}
		if isBeamIntact(v, v.Chassis[1], v.RightWheel) && isBeamIntact(v, v.Chassis[1], v.Center) {
			indices = append(indices, 1, 2, 4)
		}
		if isBeamIntact(v, v.LeftWheel, v.RightWheel) {
			indices = append(indices, 2, 3, 4)
		}
		if isBeamIntact(v, v.Chassis[0], v.LeftWheel) && isBeamIntact(v, v.Chassis[0], v.Center) {
			indices = append(indices, 3, 0, 4)
		}
	} else {
		indices = append(indices, 0, 1, 2, 0, 2, 3)
	}

	if len(indices) > 0 {
		screen.DrawTriangles(vertices, indices, texture, nil)
	}
}

func (g *Game) drawBeams(screen *ebiten.Image) {
	preset := VehiclePresets[g.currentVehicleIndex]
	hasTexture := vehicleTextures[preset.ID] != nil
	isCustomBike := preset.ID == "bike"

	for _, b := range g.world.Beams {
		if b.IsBroken {
			continue
		}

		// Если есть встроенная текстура (или это кастомный мотоцикл) и выключен оверлей рамы, не рисуем балки автомобиля
		if (hasTexture || isCustomBike) && !g.showFrameOnTop && isVehicleBeam(g.vehicle, b) {
			continue
		}

		x1 := float32(b.NodeA.Pos.X - g.camX)
		y1 := float32(b.NodeA.Pos.Y)
		x2 := float32(b.NodeB.Pos.X - g.camX)
		y2 := float32(b.NodeB.Pos.Y)

		strain := b.Stress
		absStrain := math.Abs(strain)

		t := absStrain / b.LimitFracture
		if t > 1.0 {
			t = 1.0
		}

		var c color.RGBA
		if strain > 0 {
			c = color.RGBA{uint8(60 + 195*t), uint8(160 - 120*t), uint8(90 - 40*t), 255}
		} else {
			c = color.RGBA{uint8(60 - 30*t), uint8(160 - 60*t), uint8(90 + 165*t), 255}
		}

		if absStrain > b.LimitPlastic {
			c.R = uint8(math.Min(float64(c.R)*1.2, 255))
			c.G = uint8(math.Min(float64(c.G)*1.2, 255))
			c.B = uint8(math.Min(float64(c.B)*1.2, 255))
		}

		thickness := float32(2.0)
		if strain < 0 {
			thickness += float32(absStrain * 12)
		}

		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, c, true)
	}
}

func (g *Game) drawCustomBike(screen *ebiten.Image) {
	v := g.vehicle
	n1 := v.Chassis[0] // seat
	n2 := v.Chassis[1] // steering head
	n3 := v.Chassis[2] // swingarm pivot (index 2 for 3-node chassis)

	x1, y1 := float32(n1.Pos.X-g.camX), float32(n1.Pos.Y)
	x2, y2 := float32(n2.Pos.X-g.camX), float32(n2.Pos.Y)
	x3, y3 := float32(n3.Pos.X-g.camX), float32(n3.Pos.Y)

	wL := v.LeftWheel
	wR := v.RightWheel
	xL, yL := float32(wL.Pos.X-g.camX), float32(wL.Pos.Y)
	xR, yR := float32(wR.Pos.X-g.camX), float32(wR.Pos.Y)

	// Colors
	frameColor := color.RGBA{255, 80, 0, 255}       // Orange frame
	swingarmColor := color.RGBA{100, 100, 105, 255}  // Silver-grey swingarm
	forkColor := color.RGBA{200, 200, 205, 255}      // Silver fork

	// 1. Rear swingarm (connects swingarm pivot to rear wheel)
	if !wL.IsStatic && !g.vehicle.IsWheelBroken(true) {
		vector.StrokeLine(screen, x3, y3, xL, yL, 6.0, swingarmColor, true)
	}

	// 2. Front Fork & Handlebars (connects steering head to front wheel)
	dxFork := xR - x2
	dyFork := yR - y2
	distFork := float32(math.Sqrt(float64(dxFork*dxFork + dyFork*dyFork)))
	if distFork > 0 && !wR.IsStatic && !g.vehicle.IsWheelBroken(false) {
		uxF := dxFork / distFork
		uyF := dyFork / distFork

		// Draw fork line
		vector.StrokeLine(screen, x2, y2, xR, yR, 5.0, forkColor, true)

		// Draw simple handlebars
		hbX := x2 - 5*uxF - 6*uyF
		hbY := y2 - 5*uyF + 6*uxF
		vector.StrokeLine(screen, x2, y2, hbX, hbY, 2.5, color.RGBA{40, 40, 45, 255}, true)
		vector.DrawFilledCircle(screen, hbX, hbY, 2.5, color.RGBA{15, 15, 15, 255}, true)
	}

	// 3. Simple Triangular Frame (n1 - n2 - n3)
	frameThickness := float32(4.5)
	vector.StrokeLine(screen, x1, y1, x2, y2, frameThickness, frameColor, true)
	vector.StrokeLine(screen, x2, y2, x3, y3, frameThickness, frameColor, true)
	vector.StrokeLine(screen, x3, y3, x1, y1, frameThickness, frameColor, true)
}
