package game

import (
	"gravitytrails/pkg/physics"
)

type Vehicle struct {
	LeftWheel  *physics.Node
	RightWheel *physics.Node
	Chassis    []*physics.Node
	Center     *physics.Node
	Beams      []*physics.Beam
	MaxTorque  float64
	AirTorque  float64
}

type VehiclePreset struct {
	ID            string
	Name          string
	Description   string
	ChassisWidth  float64
	ChassisHeight float64
	WheelRadius   float64
	WheelMass     float64
	WheelSpan     float64
	StiffnessCh   float64
	StiffnessSusp float64
	LimitPlastic  float64
	LimitFracture float64
	MaxTorque     float64
	AirTorque     float64
}

var VehiclePresets = []VehiclePreset{
	{
		ID:            "buggy",
		Name:          "БАГГИ",
		Description:   "Легкий, сбалансированный внедорожник. Хорошо управляется в воздухе.",
		ChassisWidth:  100,
		ChassisHeight: 40,
		WheelRadius:   22.0,
		WheelMass:     1.5,
		WheelSpan:     150,
		StiffnessCh:   0.95,
		StiffnessSusp: 0.35,
		LimitPlastic:  0.15,
		LimitFracture: 0.45,
		MaxTorque:     1400.0,
		AirTorque:     180.0,
	},
	{
		ID:            "sport",
		Name:          "СПОРТКАР",
		Description:   "Очень быстрый, низкий и жесткий. Легко разбивается при прыжках.",
		ChassisWidth:  110,
		ChassisHeight: 22,
		WheelRadius:   18.0,
		WheelMass:     1.2,
		WheelSpan:     160,
		StiffnessCh:   0.98,
		StiffnessSusp: 0.55,
		LimitPlastic:  0.10,
		LimitFracture: 0.32,
		MaxTorque:     2000.0,
		AirTorque:     240.0,
	},
	{
		ID:            "truck",
		Name:          "ГРУЗОВИК",
		Description:   "Тяжелый монстр-трак с прочной рамой и огромными колесами. Медленный.",
		ChassisWidth:  120,
		ChassisHeight: 60,
		WheelRadius:   30.0,
		WheelMass:     3.0,
		WheelSpan:     180,
		StiffnessCh:   0.90,
		StiffnessSusp: 0.25,
		LimitPlastic:  0.25,
		LimitFracture: 0.65,
		MaxTorque:     2800.0,
		AirTorque:     100.0,
	},
}

// SpawnVehicle создает автомобиль типа "Багги" (по умолчанию, для обратной совместимости)
func SpawnVehicle(w *physics.World, startX, startY float64) *Vehicle {
	return SpawnVehiclePreset(w, startX, startY, VehiclePresets[0])
}

// SpawnVehiclePreset создает автомобиль по заданному пресету
func SpawnVehiclePreset(w *physics.World, startX, startY float64, p VehiclePreset) *Vehicle {
	v := &Vehicle{
		Chassis:   make([]*physics.Node, 0),
		Beams:     make([]*physics.Beam, 0),
		MaxTorque: p.MaxTorque,
		AirTorque: p.AirTorque,
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

	return v
}

// ApplyDriveTorque применяет крутящий момент к колесам и реактивный момент к кузову
func (v *Vehicle) ApplyDriveTorque(w *physics.World, torque float64) {
	// Двигатель работает только при контакте колес с ландшафтом
	onGround := v.LeftWheel.OnGround || v.RightWheel.OnGround
	if !onGround {
		return
	}

	for _, wheel := range []*physics.Node{v.LeftWheel, v.RightWheel} {
		if wheel.OnGround {
			normal := w.GetNormal(wheel.Pos.X)
			tangent := physics.Vector2{X: -normal.Y, Y: normal.X}

			// Сила тяги, прикладываемая к колесу (двигатель толкает колесо вперед)
			driveForce := tangent.Mul(torque * wheel.Mass)
			wheel.Force = wheel.Force.Add(driveForce)

			// Реактивный вращающий момент на кузов
			var bottomAnchor, topAnchor *physics.Node
			if wheel == v.LeftWheel {
				bottomAnchor = v.Chassis[3]
				topAnchor = v.Chassis[0]
			} else {
				bottomAnchor = v.Chassis[2]
				topAnchor = v.Chassis[1]
			}

			// Сила реакции пары сил (50% от силы тяги)
			reactF := driveForce.Mul(0.5)
			bottomAnchor.Force = bottomAnchor.Force.Sub(reactF)
			topAnchor.Force = topAnchor.Force.Add(reactF)
		}
	}
}

// ApplyAirControl применяет вращающий момент в воздухе для контроля наклона (тангажа) автомобиля
func (v *Vehicle) ApplyAirControl(torque float64) {
	// Воздушный контроль активен только когда оба колеса оторвались от земли
	if v.LeftWheel.OnGround || v.RightWheel.OnGround {
		return
	}

	// Находим центр масс рамы автомобиля
	center := physics.Vector2{}
	count := 0.0
	for _, node := range v.Chassis {
		center = center.Add(node.Pos)
		count++
	}
	if count == 0 {
		return
	}
	center = center.Div(count)

	// Распределяем вращающий момент в виде касательных сил на все узлы рамы
	for _, node := range v.Chassis {
		r := node.Pos.Sub(center)
		dist := r.Len()
		if dist == 0 {
			continue
		}

		perp := physics.Vector2{X: -r.Y, Y: r.X}.Normalize()
		rotForce := perp.Mul(torque * node.Mass)
		node.Force = node.Force.Add(rotForce)
	}
}

// GetCenterOfMass возвращает центр масс автомобиля
func (v *Vehicle) GetCenterOfMass() physics.Vector2 {
	center := physics.Vector2{}
	for _, node := range v.Chassis {
		center = center.Add(node.Pos)
	}
	center = center.Add(v.LeftWheel.Pos)
	center = center.Add(v.RightWheel.Pos)
	return center.Div(float64(len(v.Chassis) + 2))
}

// GetVelocity возвращает усредненный вектор скорости автомобиля
func (v *Vehicle) GetVelocity() physics.Vector2 {
	vel := physics.Vector2{}
	for _, node := range v.Chassis {
		vel = vel.Add(node.Pos.Sub(node.OldPos))
	}
	vel = vel.Add(v.LeftWheel.Pos.Sub(v.LeftWheel.OldPos))
	vel = vel.Add(v.RightWheel.Pos.Sub(v.RightWheel.OldPos))
	return vel.Div(float64(len(v.Chassis) + 2))
}

// IsWheelBroken проверяет, оторвано ли конкретное колесо (разорвана ли его подвеска)
func (v *Vehicle) IsWheelBroken(left bool) bool {
	var wheel *physics.Node
	if left {
		wheel = v.LeftWheel
	} else {
		wheel = v.RightWheel
	}

	suspensionCount := 0
	for _, b := range v.Beams {
		if b.IsBroken {
			continue
		}

		var otherNode *physics.Node
		if b.NodeA == wheel {
			otherNode = b.NodeB
		} else if b.NodeB == wheel {
			otherNode = b.NodeA
		}

		if otherNode != nil {
			for _, chNode := range v.Chassis {
				if otherNode == chNode {
					suspensionCount++
					break
				}
			}
		}
	}
	return suspensionCount == 0
}
