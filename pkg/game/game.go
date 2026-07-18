package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"

	"gravitytrails/pkg/physics"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

var fontFaceSource *text.GoTextFaceSource

func init() {
	var err error
	fontFaceSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
}

func drawText(screen *ebiten.Image, txt string, x, y float64, size float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.LayoutOptions.LineSpacing = size * 1.2
	if clr != nil {
		op.ColorScale.ScaleWithColor(clr)
	}
	text.Draw(screen, txt, &text.GoTextFace{
		Source: fontFaceSource,
		Size:   size,
	}, op)
}

type SaveData struct {
	HighScore  float64            `json:"high_score"` // Для обратной совместимости
	HighScores map[string]float64 `json:"high_scores"`
}

type GameMap struct {
	ID          string
	Name        string
	Description string
	TerrainFunc func(x float64) float64
}

var GameMaps = []GameMap{
	{
		ID:          "hills",
		Name:        "ХОЛМЫ",
		Description: "Обычный рельеф для тренировки.",
		TerrainFunc: func(x float64) float64 {
			return 450.0 + math.Sin(x*0.005)*90.0 + math.Sin(x*0.02)*18.0 + math.Sin(x*0.08)*4.0
		},
	},
	{
		ID:          "ramps",
		Name:        "ТРАМПЛИНЫ",
		Description: "Разгон и прыжки с трамплинов!",
		TerrainFunc: func(x float64) float64 {
			cycle := 1200.0
			dx := math.Mod(x, cycle)
			if dx < 0 {
				dx += cycle
			}
			baseY := 480.0
			if dx < 400 {
				return baseY
			} else if dx < 550 {
				t := (dx - 400) / 150.0
				return baseY - 120.0*(t*t)
			} else if dx < 580 {
				t := (dx - 550) / 30.0
				return 340.0 + 140.0*t
			} else if dx < 750 {
				t := (dx - 580) / 170.0
				return 480.0 + 80.0*math.Sin(t*math.Pi/2.0)
			} else if dx < 950 {
				t := (dx - 750) / 200.0
				return 560.0 - 100.0*(t*t*(3.0-2.0*t))
			} else {
				return baseY
			}
		},
	},
	{
		ID:          "mountains",
		Name:        "ГОРЫ",
		Description: "Крутые подъемы и спуски.",
		TerrainFunc: func(x float64) float64 {
			return 400.0 + math.Sin(x*0.004)*180.0 + math.Sin(x*0.015)*35.0 + math.Sin(x*0.06)*6.0
		},
	},
}

func getSmoothedHeight(x float64, baseFunc func(x float64) float64) float64 {
	spawnFlatX := 350.0
	spawnTransitionX := 150.0
	spawnBaseY := 480.0

	if x < spawnFlatX {
		return spawnBaseY
	} else if x < spawnFlatX+spawnTransitionX {
		t := (x - spawnFlatX) / spawnTransitionX
		blend := t * t * (3.0 - 2.0*t) // smoothstep
		return spawnBaseY + (baseFunc(x)-spawnBaseY)*blend
	}
	return baseFunc(x)
}

type GameState int

const (
	StateMenu GameState = iota
	StateGame
)

type Game struct {
	screenWidth         int
	screenHeight        int
	world               *physics.World
	vehicle             *Vehicle
	draggedNode         *physics.Node
	camX                float64 // Текущая позиция камеры по горизонтали
	isCrashed           bool    // Флаг аварии
	maxDistance         float64 // Текущий рекорд для активной карты
	allHighScores     map[string]float64
	crashedTimer      float64 // Таймер до фиксации крушения при перевороте (секунды)
	wheelLossTimer    float64 // Таймер до фиксации поражения при отрыве обоих колес
	currentSessionMax float64 // Максимальная дистанция в текущей сессии
	currentMapIndex     int     // Индекс активной карты
	currentVehicleIndex int     // Индекс активного автомобиля
	state               GameState
	hoveredVehicleIndex int
	hoveredMapIndex     int
}

func NewGame(width, height int) *Game {
	g := &Game{
		screenWidth:         width,
		screenHeight:        height,
		world:               physics.NewWorld(),
		camX:                0,
		currentMapIndex:     0,
		currentVehicleIndex: 0,
		state:               StateMenu,
		hoveredVehicleIndex: -1,
		hoveredMapIndex:     -1,
	}
	g.loadHighScore() // Загрузка рекорда
	g.initSandbox()
	return g
}

func (g *Game) initSandbox() {
	g.world.Clear()
	g.isCrashed = false
	g.crashedTimer = 0.0
	g.wheelLossTimer = 0.0
	g.currentSessionMax = 0.0

	// Настраиваем процедурный ландшафт выбранной карты с плавным спавном
	activeMap := GameMaps[g.currentMapIndex]
	g.world.TerrainFunc = func(x float64) float64 {
		return getSmoothedHeight(x, activeMap.TerrainFunc)
	}

	// Координаты спавна машины
	startX := 150.0
	startY := 200.0

	// Создаем автомобиль выбранного пресета
	g.vehicle = SpawnVehiclePreset(g.world, startX, startY, VehiclePresets[g.currentVehicleIndex])
}

// loadHighScore считывает рекорд из save.json
func (g *Game) loadHighScore() {
	g.allHighScores = make(map[string]float64)
	for _, m := range GameMaps {
		g.allHighScores[m.ID] = 0.0
	}

	file, err := os.Open("save.json")
	if err != nil {
		g.maxDistance = 0.0
		return
	}
	defer file.Close()

	var data SaveData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		g.maxDistance = 0.0
		return
	}

	// Миграция старого рекорда
	if data.HighScore > 0.0 {
		g.allHighScores["hills"] = data.HighScore
	}

	for id, score := range data.HighScores {
		g.allHighScores[id] = score
	}

	activeMap := GameMaps[g.currentMapIndex]
	g.maxDistance = g.allHighScores[activeMap.ID]
}

// saveHighScore сохраняет рекорд в save.json
func (g *Game) saveHighScore() {
	activeMap := GameMaps[g.currentMapIndex]
	if g.currentSessionMax > g.allHighScores[activeMap.ID] {
		g.allHighScores[activeMap.ID] = g.currentSessionMax
	}
	g.maxDistance = g.allHighScores[activeMap.ID]

	data := SaveData{
		HighScore:  g.allHighScores["hills"], // Совместимость
		HighScores: g.allHighScores,
	}

	file, err := os.Create("save.json")
	if err != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

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

func (g *Game) Update() error {
	if g.state == StateMenu {
		return g.UpdateMenu()
	}
	return g.UpdateGame()
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

	// Сбрасываем трение колес на свободное качение
	g.vehicle.LeftWheel.Friction = 0.03
	g.vehicle.RightWheel.Friction = 0.03

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

// checkCrashedState проверяет состояние автомобиля и управляет таймерами спасения
func (g *Game) checkCrashedState(dt float64) {
	if g.isCrashed {
		return
	}

	// 1. Мгновенная смерть от сдавливания кабины (смятие по высоте на 60% и более)
	// Исходная высота кузова ~40px. Смерть наступает, если высота падает ниже 16px.
	leftHeight := g.vehicle.Chassis[0].Pos.Dist(g.vehicle.Chassis[3].Pos)
	rightHeight := g.vehicle.Chassis[1].Pos.Dist(g.vehicle.Chassis[2].Pos)
	if leftHeight < 16.0 || rightHeight < 16.0 {
		g.isCrashed = true
		g.saveHighScore()
		return
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

func (g *Game) Draw(screen *ebiten.Image) {
	if g.state == StateMenu {
		g.DrawMenu(screen)
	} else {
		g.DrawGame(screen)
	}
}

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

func (g *Game) DrawGame(screen *ebiten.Image) {
	// Премиальный темно-синий цвет неба
	screen.Fill(color.RGBA{15, 16, 22, 255})

	// 1. Отрисовка ландшафта
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

	// 2. Отрисовка упругих связей (Beams)
	for _, b := range g.world.Beams {
		if b.IsBroken {
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
			r := uint8(60 + 195*t)
			g := uint8(160 - 120*t)
			b := uint8(90 - 40*t)
			c = color.RGBA{r, g, b, 255}
		} else {
			r := uint8(60 - 30*t)
			g := uint8(160 - 60*t)
			b := uint8(90 + 165*t)
			c = color.RGBA{r, g, b, 255}
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

	// 3. Отрисовка узлов и колес
	for _, node := range g.world.Nodes {
		cx := float32(node.Pos.X - g.camX)
		cy := float32(node.Pos.Y)

		if node.IsWheel {
			isLeftWheel := node == g.vehicle.LeftWheel
			if g.vehicle.IsWheelBroken(isLeftWheel) {
				vector.DrawFilledCircle(screen, cx, cy, float32(node.Radius), color.RGBA{80, 80, 90, 180}, true)
				continue
			}

			radius := float32(node.Radius)
			vector.DrawFilledCircle(screen, cx, cy, radius, color.RGBA{35, 35, 45, 255}, true)
			vector.StrokeCircle(screen, cx, cy, radius-1.5, 3, color.RGBA{180, 180, 200, 255}, true)

			for i := 0; i < 4; i++ {
				angle := node.Rotation + float64(i)*math.Pi/2.0
				sx := cx + float32(math.Cos(angle))*(radius-3)
				sy := cy + float32(math.Sin(angle))*(radius-3)
				vector.StrokeLine(screen, cx, cy, sx, sy, 2, color.RGBA{120, 120, 140, 255}, true)
			}

			vector.DrawFilledCircle(screen, cx, cy, 5, color.RGBA{0, 255, 230, 255}, true)
			vector.DrawFilledCircle(screen, cx, cy, 3, color.RGBA{20, 20, 30, 255}, true)
		} else {
			var nodeColor color.RGBA
			if node == g.draggedNode {
				nodeColor = color.RGBA{0, 255, 230, 255}
			} else if node.IsStatic {
				nodeColor = color.RGBA{255, 215, 0, 255}
			} else {
				nodeColor = color.RGBA{200, 200, 210, 255}
			}

			vector.DrawFilledCircle(screen, cx, cy, 7, color.RGBA{30, 30, 40, 255}, true)
			vector.DrawFilledCircle(screen, cx, cy, 4, nodeColor, true)
		}
	}

	// 4. Отрисовка HUD
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

	// Маленькая аккуратная подсказка в левом нижнем углу
	drawText(screen, "ESC: Меню | R: Сброс", 20, float64(g.screenHeight)-25, 11, color.RGBA{120, 120, 130, 255})

	// 5. Вывод предупреждений
	if !g.isCrashed {
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

			msg := fmt.Sprintf("ПЕРЕВЕРНУТО! РАСКАЧАЙТЕ КУЗОВ [A / D]: %.1fс", math.Max(0.0, 1.5-g.crashedTimer))
			drawText(screen, msg, float64(boxX)+15, float64(boxY)+10, 12, txtColor)
		} else if g.wheelLossTimer > 0.0 {
			boxW := float32(340)
			boxH := float32(35)
			boxX := float32(g.screenWidth)/2 - boxW/2
			boxY := float32(80)
			vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{25, 10, 10, 200}, true)
			vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 1.5, color.RGBA{255, 150, 0, 255}, true)

			msg := fmt.Sprintf("ПОТЕРЯ ОБОИХ КОЛЕС! КОНЕЦ ЧЕРЕЗ: %.1fс", math.Max(0.0, 2.0-g.wheelLossTimer))
			drawText(screen, msg, float64(boxX)+15, float64(boxY)+10, 12, color.RGBA{255, 150, 0, 255})
		}
	}

	// 6. Оверлей при аварии
	if g.isCrashed {
		vector.DrawFilledRect(screen, 0, 0, float32(g.screenWidth), float32(g.screenHeight), color.RGBA{80, 0, 0, 140}, true)

		var msgCause string
		leftBroken := g.vehicle.IsWheelBroken(true)
		rightBroken := g.vehicle.IsWheelBroken(false)
		leftHeight := g.vehicle.Chassis[0].Pos.Dist(g.vehicle.Chassis[3].Pos)
		rightHeight := g.vehicle.Chassis[1].Pos.Dist(g.vehicle.Chassis[2].Pos)

		if leftHeight < 16.0 || rightHeight < 16.0 {
			msgCause = "Кабина водителя полностью раздавлена!"
		} else if leftBroken && rightBroken {
			msgCause = "Оторваны все колеса автомобиля!"
		} else {
			msgCause = "Машина перевернулась и загорелась!"
		}

		msgTitle := "=== ВНИМАНИЕ: КРУШЕНИЕ ==="
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}
