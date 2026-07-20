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
