package game

import (
	"gravitytrails/pkg/physics"
)

type Vehicle struct {
	LeftWheel     *physics.Node
	RightWheel    *physics.Node
	Chassis       []*physics.Node
	Center        *physics.Node
	Beams         []*physics.Beam
	MaxTorque     float64
	AirTorque     float64
	DriveReaction float64
	Spoiler       *physics.Node
	Bumper        *physics.Node
	SpoilerBeams []*physics.Beam
	BumperBeams  []*physics.Beam
}

// SpawnVehicle создает автомобиль типа "Багги" (по умолчанию, для обратной совместимости)
func SpawnVehicle(w *physics.World, startX, startY float64) *Vehicle {
	return SpawnVehiclePreset(w, startX, startY, VehiclePresets[0])
}

// SpawnVehiclePreset создает автомобиль по заданному пресету
func SpawnVehiclePreset(w *physics.World, startX, startY float64, p VehiclePreset) *Vehicle {
	v := &Vehicle{
		Chassis:       make([]*physics.Node, 0),
		Beams:         make([]*physics.Beam, 0),
		MaxTorque:     p.MaxTorque,
		AirTorque:     p.AirTorque,
		DriveReaction: p.DriveReaction,
	}

	midX := startX + p.WheelSpan/2.0

	// 1. Создаем узлы рамы/кабины (устойчивая трапеция)
	topW := p.ChassisWidth * 0.7 // кабина сужается кверху
	n1 := w.AddNode(midX-topW/2.0, startY, 0.5, false)
	n2 := w.AddNode(midX+topW/2.0, startY, 0.5, false)

	// Масса нижних узлов рамы в зависимости от класса
	chassisMass := 1.2
	if p.ID == "sport" {
		chassisMass = 0.8
	} else if p.ID == "truck" {
		chassisMass = 2.5
	}
	n3 := w.AddNode(midX+p.ChassisWidth/2.0, startY+p.ChassisHeight, chassisMass, false)
	n4 := w.AddNode(midX-p.ChassisWidth/2.0, startY+p.ChassisHeight, chassisMass, false)

	v.Chassis = append(v.Chassis, n1, n2, n3, n4)

	// Дополнительный узел жесткости в центре кабины
	v.Center = w.AddNode(midX, startY+p.ChassisHeight/2.0, chassisMass*0.7, false)
	v.Chassis = append(v.Chassis, v.Center)

	// 2. Колеса автомобиля
	v.LeftWheel = w.AddWheelNode(midX-p.WheelSpan/2.0, startY+p.ChassisHeight+35.0, p.WheelMass, p.WheelRadius)
	v.RightWheel = w.AddWheelNode(midX+p.WheelSpan/2.0, startY+p.ChassisHeight+35.0, p.WheelMass, p.WheelRadius)

	// 3. Соединение рамы кабины
	v.Beams = append(v.Beams, w.AddBeam(n1, n2, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n2, n3, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n3, n4, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n4, n1, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))

	// Внутренние ребра жесткости (крестовина кабины к центру)
	v.Beams = append(v.Beams, w.AddBeam(n1, v.Center, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n2, v.Center, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n3, v.Center, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n4, v.Center, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))

	// 4. Соединение подвески колес к кузову (крепление к нижней и верхней точкам)
	v.Beams = append(v.Beams, w.AddBeam(n4, v.LeftWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n1, v.LeftWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n3, v.RightWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))
	v.Beams = append(v.Beams, w.AddBeam(n2, v.RightWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))

	// Поддерживающая гибкая стяжка между осями колес
	v.Beams = append(v.Beams, w.AddBeam(v.LeftWheel, v.RightWheel, 0.4, p.LimitPlastic, p.LimitFracture))

	// 5. Создаем отрывающиеся детали (Спойлер и Бампер)
	spoilerX := midX - p.ChassisWidth/2.0 - 15.0
	spoilerY := startY - 10.0
	v.Spoiler = w.AddNode(spoilerX, spoilerY, 0.4, false)

	bumperX := midX + p.ChassisWidth/2.0 + 15.0
	bumperY := startY + p.ChassisHeight - 5.0
	v.Bumper = w.AddNode(bumperX, bumperY, 0.4, false)

	// Хрупкие крепления спойлера (LimitFracture = 0.1)
	sb1 := w.AddBeam(v.Spoiler, n1, 0.5, p.LimitPlastic, 0.1)
	sb2 := w.AddBeam(v.Spoiler, n4, 0.5, p.LimitPlastic, 0.1)
	v.SpoilerBeams = []*physics.Beam{sb1, sb2}
	v.Beams = append(v.Beams, sb1, sb2)

	// Хрупкие крепления бампера (LimitFracture = 0.1)
	bb1 := w.AddBeam(v.Bumper, n2, 0.5, p.LimitPlastic, 0.1)
	bb2 := w.AddBeam(v.Bumper, n3, 0.5, p.LimitPlastic, 0.1)
	v.BumperBeams = []*physics.Beam{bb1, bb2}
	v.Beams = append(v.Beams, bb1, bb2)

	return v
}
