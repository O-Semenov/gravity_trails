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

	// UV координаты для точного проецирования картинки на узлы (в процентах 0..1)
	UvTopLeftX    float64
	UvTopLeftY    float64
	UvTopRightX   float64
	UvTopRightY   float64
	UvLeftWheelX  float64
	UvLeftWheelY  float64
	UvRightWheelX float64
	UvRightWheelY float64
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
		// UV координаты: кабина в центре сверху, колеса по бокам снизу
		UvTopLeftX:    0.28,
		UvTopLeftY:    0.20,
		UvTopRightX:   0.72,
		UvTopRightY:   0.20,
		UvLeftWheelX:  0.18,
		UvLeftWheelY:  0.72,
		UvRightWheelX: 0.82,
		UvRightWheelY: 0.72,
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
		UvTopLeftX:    0.32,
		UvTopLeftY:    0.28,
		UvTopRightX:   0.68,
		UvTopRightY:   0.28,
		UvLeftWheelX:  0.18,
		UvLeftWheelY:  0.78,
		UvRightWheelX: 0.82,
		UvRightWheelY: 0.78,
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
		DriveReaction: 0.0,
		UvTopLeftX:    0.30,
		UvTopLeftY:    0.42,
		UvTopRightX:   0.70,
		UvTopRightY:   0.42,
		UvLeftWheelX:  0.17,
		UvLeftWheelY:  0.65,
		UvRightWheelX: 0.71,
		UvRightWheelY: 0.65,
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
		UvTopLeftX:    0.35,
		UvTopLeftY:    0.18,
		UvTopRightX:   0.65,
		UvTopRightY:   0.18,
		UvLeftWheelX:  0.15,
		UvLeftWheelY:  0.75,
		UvRightWheelX: 0.85,
		UvRightWheelY: 0.75,
	},
}
