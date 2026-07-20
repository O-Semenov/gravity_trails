package game

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
	DriveReaction float64
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
		DriveReaction: 0.15,
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
		DriveReaction: 0.20,
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
		DriveReaction: 0.10,
	},
	{
		ID:            "bike",
		Name:          "МОТОЦИКЛ",
		Description:   "Сверхлегкий и маневренный. Легко козлит и переворачивается.",
		ChassisWidth:  70,
		ChassisHeight: 50,
		WheelRadius:   18.0,
		WheelMass:     0.8,
		WheelSpan:     90,
		StiffnessCh:   0.96,
		StiffnessSusp: 0.40,
		LimitPlastic:  0.12,
		LimitFracture: 0.35,
		MaxTorque:     1200.0,
		AirTorque:     300.0,
		DriveReaction: 0.40,
	},
}
