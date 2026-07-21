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
	dst4 := v.Center.Pos

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

	if len(indices) > 0 {
		screen.DrawTriangles(vertices, indices, texture, nil)
	}
}

func (g *Game) drawBeams(screen *ebiten.Image) {
	preset := VehiclePresets[g.currentVehicleIndex]
	hasTexture := vehicleTextures[preset.ID] != nil

	for _, b := range g.world.Beams {
		if b.IsBroken {
			continue
		}

		// Если есть встроенная текстура и выключен оверлей рамы, не рисуем балки автомобиля
		if hasTexture && !g.showFrameOnTop && isVehicleBeam(g.vehicle, b) {
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

func (g *Game) drawDetachableParts(screen *ebiten.Image) {
	preset := VehiclePresets[g.currentVehicleIndex]
	if preset.ID == "bike" {
		return
	}

	hasTexture := vehicleTextures[preset.ID] != nil
	if !hasTexture {
		return
	}

	v := g.vehicle
	var bodyColor color.Color
	bodyColor = color.RGBA{0, 0, 0, 0}

	if v.Spoiler != nil {
		scx := float32(v.Spoiler.Pos.X - g.camX)
		scy := float32(v.Spoiler.Pos.Y)

		if isBeamIntact(v, v.Spoiler, v.Chassis[0]) {
			x1 := float32(v.Chassis[0].Pos.X - g.camX)
			y1 := float32(v.Chassis[0].Pos.Y)
			vector.StrokeLine(screen, scx, scy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}
		if isBeamIntact(v, v.Spoiler, v.Chassis[3]) {
			x1 := float32(v.Chassis[3].Pos.X - g.camX)
			y1 := float32(v.Chassis[3].Pos.Y)
			vector.StrokeLine(screen, scx, scy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}

		vector.StrokeLine(screen, scx-12, scy, scx+12, scy, 4, color.RGBA{30, 30, 30, 255}, true)
	}

	if v.Bumper != nil {
		bcx := float32(v.Bumper.Pos.X - g.camX)
		bcy := float32(v.Bumper.Pos.Y)

		if isBeamIntact(v, v.Bumper, v.Chassis[1]) {
			x1 := float32(v.Chassis[1].Pos.X - g.camX)
			y1 := float32(v.Chassis[1].Pos.Y)
			vector.StrokeLine(screen, bcx, bcy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}
		if isBeamIntact(v, v.Bumper, v.Chassis[2]) {
			x1 := float32(v.Chassis[2].Pos.X - g.camX)
			y1 := float32(v.Chassis[2].Pos.Y)
			vector.StrokeLine(screen, bcx, bcy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}

		vector.StrokeLine(screen, bcx, bcy-10, bcx, bcy+10, 5, bodyColor, true)
	}
}

func (g *Game) drawCustomBike(screen *ebiten.Image) {
	v := g.vehicle
	n1 := v.Chassis[0] // seat
	n2 := v.Chassis[1] // steering head
	n3 := v.Chassis[2] // engine front bottom
	n4 := v.Chassis[3] // swingarm pivot
	center := v.Center

	x1, y1 := float32(n1.Pos.X-g.camX), float32(n1.Pos.Y)
	x2, y2 := float32(n2.Pos.X-g.camX), float32(n2.Pos.Y)
	x3, y3 := float32(n3.Pos.X-g.camX), float32(n3.Pos.Y)
	x4, y4 := float32(n4.Pos.X-g.camX), float32(n4.Pos.Y)
	xc, yc := float32(center.Pos.X-g.camX), float32(center.Pos.Y)

	wL := v.LeftWheel
	wR := v.RightWheel
	xL, yL := float32(wL.Pos.X-g.camX), float32(wL.Pos.Y)
	xR, yR := float32(wR.Pos.X-g.camX), float32(wR.Pos.Y)

	// Colors
	trellisColor := color.RGBA{255, 60, 0, 255}      // KTM orange frame
	swingarmColor := color.RGBA{100, 100, 105, 255}  // Silver-grey swingarm
	engineColor := color.RGBA{45, 45, 50, 255}       // Dark crankcase
	engineFinColor := color.RGBA{140, 140, 150, 255} // Cooling fins
	tankColor := color.RGBA{255, 80, 0, 255}         // Tank orange
	seatColor := color.RGBA{20, 20, 25, 255}         // Black seat
	exhaustColor := color.RGBA{170, 170, 180, 255}   // Chrome muffler
	springColor := color.RGBA{230, 30, 30, 255}      // Red shock spring

	// Precompute fork directions for triple clamps and handlebars
	var hbX, hbY, uxF, uyF float32
	dxFork := xR - x2
	dyFork := yR - y2
	distFork := float32(math.Sqrt(float64(dxFork*dxFork + dyFork*dyFork)))
	if distFork > 0 {
		uxF = dxFork / distFork
		uyF = dyFork / distFork
		hbX = x2 - 5*uxF - 6*uyF
		hbY = y2 - 5*uyF + 6*uxF
	} else {
		hbX = x2 - 5
		hbY = y2 - 5
	}

	// 1. Rear swingarm
	if !wL.IsStatic {
		vector.StrokeLine(screen, x4, y4, xL, yL, 7.5, swingarmColor, true)
		vector.DrawFilledCircle(screen, x4, y4, 4.5, color.RGBA{30, 30, 35, 255}, true)
		vector.DrawFilledCircle(screen, xL, yL, 4.5, color.RGBA{30, 30, 35, 255}, true)
	}

	// 2. Rear shock absorber (Mono-shock)
	dxShock := xL - x1
	dyShock := yL - y1
	distShock := float32(math.Sqrt(float64(dxShock*dxShock + dyShock*dyShock)))
	if distShock > 0 {
		ux := dxShock / distShock
		uy := dyShock / distShock
		nx := -uy
		ny := ux

		vector.StrokeLine(screen, x1, y1, xL, yL, 3.0, color.RGBA{160, 160, 170, 255}, true)
		vector.DrawFilledCircle(screen, x1, y1, 3.5, color.RGBA{35, 35, 40, 255}, true)

		startDist := float32(8.0)
		endDist := distShock - 10.0
		springLen := endDist - startDist

		if springLen > 5.0 {
			coils := 7
			steps := coils * 2
			springWidth := float32(5.5)

			var lastX, lastY float32
			for i := 0; i <= steps; i++ {
				t := float32(i) / float32(steps)
				d := startDist + t*springLen
				px := x1 + d*ux
				py := y1 + d*uy

				if i > 0 && i < steps {
					side := float32(1.0)
					if i%2 == 1 {
						side = -1.0
					}
					px += side * springWidth * nx
					py += side * springWidth * ny
				}

				if i > 0 {
					vector.StrokeLine(screen, lastX, lastY, px, py, 3.0, springColor, true)
				}
				lastX = px
				lastY = py
			}
		}
	}

	// 3. Exhaust pipe
	exStartX := xc
	exStartY := yc + 6.0
	exMidX := x4
	exMidY := y4 + 8.0
	exEndX := x4 - 22.0
	exEndY := y4 - 6.0
	vector.StrokeLine(screen, exStartX, exStartY, exMidX, exMidY, 3.0, exhaustColor, true)
	vector.StrokeLine(screen, exMidX, exMidY, exEndX, exEndY, 5.5, exhaustColor, true)
	vector.DrawFilledCircle(screen, exEndX, exEndY, 2.0, color.RGBA{20, 20, 25, 255}, true)

	// 4. Engine
	vector.DrawFilledRect(screen, xc-14.0, yc-6.0, 28.0, 18.0, engineColor, true)
	cylX := xc + 4.0
	cylY := yc - 8.0
	vector.DrawFilledRect(screen, cylX-7.0, cylY-9.0, 14.0, 10.0, engineColor, true)
	for i := 0; i < 4; i++ {
		finY := cylY - 8.0 + float32(i)*2.8
		vector.StrokeLine(screen, cylX-9.5, finY, cylX+9.5, finY, 1.5, engineFinColor, true)
	}
	vector.DrawFilledCircle(screen, xc-5.0, yc+5.0, 6.0, color.RGBA{184, 115, 51, 255}, true) // Bronze clutch cover

	// 5. Trellis frame
	frameThickness := float32(3.5)
	vector.StrokeLine(screen, x1, y1, x2, y2, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x2, y2, x3, y3, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x3, y3, x4, y4, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x4, y4, x1, y1, frameThickness, trellisColor, true)

	vector.StrokeLine(screen, x1, y1, xc, yc, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x2, y2, xc, yc, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x3, y3, xc, yc, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x4, y4, xc, yc, frameThickness, trellisColor, true)

	vector.StrokeLine(screen, x1, y1, x3, y3, frameThickness, trellisColor, true)
	vector.StrokeLine(screen, x2, y2, x4, y4, frameThickness, trellisColor, true)

	// 6. Gas Tank & Seat
	tankTopX := x1*0.4 + x2*0.6
	tankTopY := y1*0.4 + y2*0.6 - 11.0
	vector.DrawFilledCircle(screen, tankTopX, tankTopY+3.0, 9.0, tankColor, true)
	vector.StrokeCircle(screen, tankTopX, tankTopY+3.0, 9.0, 1.0, color.RGBA{255, 255, 255, 100}, true)
	vector.StrokeLine(screen, x1, y1-3.0, tankTopX, tankTopY, 5.0, tankColor, true)
	vector.StrokeLine(screen, x2, y2-2.0, tankTopX, tankTopY, 11.0, tankColor, true)

	vector.StrokeLine(screen, x1-9.0, y1-2.0, x1+4.0, y1, 4.5, seatColor, true)

	// 7. Front Fork
	if distFork > 0 {
		forkGoldLen := distFork * 0.55
		goldEndX := x2 + forkGoldLen*uxF
		goldEndY := y2 + forkGoldLen*uyF

		vector.StrokeLine(screen, goldEndX, goldEndY, xR, yR, 4.0, color.RGBA{210, 210, 215, 255}, true) // Chrome
		vector.StrokeLine(screen, x2, y2, goldEndX, goldEndY, 6.0, color.RGBA{218, 165, 32, 255}, true)  // Gold

		// Triple clamps
		vector.StrokeLine(screen, x2-4.5*uyF, y2+4.5*uxF, x2+4.5*uyF, y2-4.5*uyF, 3.0, color.RGBA{45, 45, 50, 255}, true)
		vector.StrokeLine(screen, x2+10*uxF-4.5*uyF, y2+10*uyF+4.5*uxF, x2+10*uxF+4.5*uyF, y2+10*uyF-4.5*uxF, 3.0, color.RGBA{45, 45, 50, 255}, true)

		// Handlebars
		vector.StrokeLine(screen, x2, y2, hbX, hbY, 2.5, color.RGBA{40, 40, 45, 255}, true)
		vector.DrawFilledCircle(screen, hbX, hbY, 2.5, color.RGBA{15, 15, 15, 255}, true)
	}

	// 8. Dynamic Rider
	pelvisX := x1
	pelvisY := y1 - 4.0
	shoulderX := xc + 4.0
	shoulderY := yc - 20.0
	headX := shoulderX + 2.0
	headY := shoulderY - 14.0

	jacketColor := color.RGBA{30, 30, 35, 255}
	vector.StrokeLine(screen, pelvisX, pelvisY, shoulderX, shoulderY, 7.5, jacketColor, true)
	vector.DrawFilledCircle(screen, headX, headY, 6.0, color.RGBA{255, 215, 0, 255}, true)
	vector.DrawFilledCircle(screen, headX+3.0, headY+1.0, 3.0, color.RGBA{0, 220, 255, 220}, true) // visor

	vector.StrokeLine(screen, shoulderX, shoulderY, hbX, hbY, 3.2, jacketColor, true)
	vector.StrokeLine(screen, pelvisX, pelvisY, x4, y4, 4.2, jacketColor, true)
	vector.StrokeLine(screen, x4, y4, x4+5.0, y4+3.0, 3.5, seatColor, true)

	// 9. Attachments: Fender & Fairing
	if v.Spoiler != nil {
		scx := float32(v.Spoiler.Pos.X - g.camX)
		scy := float32(v.Spoiler.Pos.Y)
		if isBeamIntact(v, v.Spoiler, n1) || isBeamIntact(v, v.Spoiler, n4) {
			vector.StrokeLine(screen, x1, y1-3.0, scx, scy, 4.5, tankColor, true)
			vector.StrokeLine(screen, x4, y4, scx, scy, 1.8, color.RGBA{45, 45, 50, 255}, true)
		}
	}

	if v.Bumper != nil {
		bcx := float32(v.Bumper.Pos.X - g.camX)
		bcy := float32(v.Bumper.Pos.Y)
		if isBeamIntact(v, v.Bumper, n2) || isBeamIntact(v, v.Bumper, n3) {
			vector.StrokeLine(screen, x2, y2, bcx, bcy, 5.5, tankColor, true)
			vector.StrokeLine(screen, x3, y3, bcx, bcy, 2.5, color.RGBA{45, 45, 50, 255}, true)
			vector.DrawFilledCircle(screen, bcx+2.0, bcy, 3.0, color.RGBA{255, 255, 0, 255}, true)

			// Headlight beam lines
			vector.StrokeLine(screen, bcx+3.0, bcy, bcx+35.0, bcy-8.0, 1.0, color.RGBA{255, 255, 0, 45}, true)
			vector.StrokeLine(screen, bcx+3.0, bcy, bcx+35.0, bcy+8.0, 1.0, color.RGBA{255, 255, 0, 45}, true)
		}
	}
}
