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

	if p.ID == "bike" {
		// 1. Узлы рамы мотоцикла (3-узловая треугольная рама)
		// n1: сиденье (Top-Left)
		n1 := w.AddNode(midX-25.0, startY+5.0, 0.6, false)
		// n2: рулевая колонка (Top-Right)
		n2 := w.AddNode(midX+25.0, startY-5.0, 0.6, false)
		// n3: точка крепления маятника (Bottom)
		n3 := w.AddNode(midX-10.0, startY+p.ChassisHeight, 1.2, false)

		v.Chassis = append(v.Chassis, n1, n2, n3)
		v.Center = nil // Нет центрального узла для треугольника

		// 2. Колеса
		// Заднее колесо (LeftWheel) смещено назад
		v.LeftWheel = w.AddWheelNode(midX-p.WheelSpan/2.0, startY+p.ChassisHeight+25.0, p.WheelMass, p.WheelRadius)
		// Переднее колесо (RightWheel) смещено вперед
		v.RightWheel = w.AddWheelNode(midX+p.WheelSpan/2.0, startY+p.ChassisHeight+25.0, p.WheelMass, p.WheelRadius)

		// 3. Жесткая рама мотоцикла (треугольник)
		v.Beams = append(v.Beams, w.AddBeam(n1, n2, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
		v.Beams = append(v.Beams, w.AddBeam(n2, n3, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
		v.Beams = append(v.Beams, w.AddBeam(n3, n1, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))

		// 4. Подвеска
		// Задняя подвеска: маятник к n3
		v.Beams = append(v.Beams, w.AddBeam(n3, v.LeftWheel, p.StiffnessCh, p.LimitPlastic, p.LimitFracture))
		// Задний амортизатор: пружина к n1
		v.Beams = append(v.Beams, w.AddBeam(n1, v.LeftWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))

		// Передняя подвеска: вилка от рулевой колонки (n2)
		v.Beams = append(v.Beams, w.AddBeam(n2, v.RightWheel, p.StiffnessSusp, p.LimitPlastic, p.LimitFracture))
		// Направляющая вилки от нижней точки рамы (n3)
		v.Beams = append(v.Beams, w.AddBeam(n3, v.RightWheel, p.StiffnessSusp*1.1, p.LimitPlastic, p.LimitFracture))

	} else {
		// 1. Создаем узлы рамы/кабины (5-узловая устойчивая трапеция)
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
	}

	return v
}
