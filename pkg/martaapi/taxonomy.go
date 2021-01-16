package martaapi

//Direction enumerates all valid MARTA direction codes
type Direction string

const (
	North Direction = "Northbound"
	South Direction = "Southbound"
	East  Direction = "Eastbound"
	West  Direction = "Westbound"
)

//Directions is for checking whether a string represents a valid Direction
var Directions = map[Direction]struct{}{
	North: {},
	South: {},
	East:  {},
	West:  {},
}

//Line enumerates all valid MARTA line names
type Line string

const (
	Green Line = "Green"
	Blue  Line = "Blue"
	Gold  Line = "Gold"
	Red   Line = "Red"
)

//Lines is for checking whether a string represents a valid Line
var Lines = map[Line]struct{}{
	Green: {},
	Blue:  {},
	Gold:  {},
	Red:   {},
}

//Station enumerates all valid MARTA station names
type Station string

const (
	AirportStation             Station = "Airport"
	ArtsCenterStation          Station = "Arts Center"
	AshbyStation               Station = "Ashby"
	AvondaleStation            Station = "Avondale"
	BankheadStation            Station = "Bankhead"
	BrookhavenStation          Station = "Brookhaven"
	BuckheadStation            Station = "Buckhead"
	ChambleeStation            Station = "Chamblee"
	CivicCenterStation         Station = "Civic Center"
	CollegeParkStation         Station = "College Park"
	DecaturStation             Station = "Decatur"
	DoravilleStation           Station = "Doraville"
	DunwoodyStation            Station = "Dunwoody"
	EastLakeStation            Station = "East Lake"
	EastPointStation           Station = "East Point"
	EdgewoodCandlerParkStation Station = "Edgewood-Candler Park"
	FivePointsStation          Station = "Five Points"
	GarnettStation             Station = "Garnett"
	GeorgiaStateStation        Station = "Georgia State"
	HamiltonEHolmesStation     Station = "H. E. Holmes"
	IndianCreekStation         Station = "Indian Creek"
	InmanParkStation           Station = "Inman Park"
	KensingtonStation          Station = "Kensington"
	KingMemorialStation        Station = "King Memorial"
	LakewoodStation            Station = "Lakewood"
	LenoxStation               Station = "Lenox"
	LindberghStation           Station = "Lindbergh Center"
	MedicalCenterStation       Station = "Medical Center"
	MidtownStation             Station = "Midtown"
	NorthAveStation            Station = "North Avenue"
	NorthSpringsStation        Station = "North Springs"
	OaklandCityStation         Station = "Oakland City"
	OmniDomeStation            Station = "Omni Dome"
	PeachtreeCenterStation     Station = "Peachtree Center"
	SandySpringsStation        Station = "Sandy Springs"
	VineCityStation            Station = "Vine City"
	WestEndStation             Station = "West End"
	WestLakeStation            Station = "West Lake"
)

//Stations is for checking whether a string represents a valid station
var Stations = map[Station]struct{}{
	FivePointsStation:          {},
	OaklandCityStation:         {},
	AirportStation:             {},
	BrookhavenStation:          {},
	KensingtonStation:          {},
	OmniDomeStation:            {},
	DecaturStation:             {},
	InmanParkStation:           {},
	SandySpringsStation:        {},
	IndianCreekStation:         {},
	DunwoodyStation:            {},
	KingMemorialStation:        {},
	VineCityStation:            {},
	NorthSpringsStation:        {},
	MedicalCenterStation:       {},
	ArtsCenterStation:          {},
	EastPointStation:           {},
	CivicCenterStation:         {},
	CollegeParkStation:         {},
	ChambleeStation:            {},
	AvondaleStation:            {},
	LindberghStation:           {},
	GarnettStation:             {},
	PeachtreeCenterStation:     {},
	WestLakeStation:            {},
	EastLakeStation:            {},
	GeorgiaStateStation:        {},
	BankheadStation:            {},
	BuckheadStation:            {},
	WestEndStation:             {},
	LakewoodStation:            {},
	MidtownStation:             {},
	DoravilleStation:           {},
	LenoxStation:               {},
	EdgewoodCandlerParkStation: {},
	NorthAveStation:            {},
	AshbyStation:               {},
	HamiltonEHolmesStation:     {},
}

//LineStations provides the stations on the line, in the order
//specified by LineDirections[line][0].
var LineStations = map[Line][]Station{
	Green: {
		BankheadStation,
		AshbyStation,
		VineCityStation,
		OmniDomeStation,
		FivePointsStation,
		GeorgiaStateStation,
		KingMemorialStation,
		InmanParkStation,
		EdgewoodCandlerParkStation,
	},
	Blue: {
		HamiltonEHolmesStation,
		WestLakeStation,
		AshbyStation,
		VineCityStation,
		OmniDomeStation,
		FivePointsStation,
		GeorgiaStateStation,
		KingMemorialStation,
		InmanParkStation,
		EdgewoodCandlerParkStation,
		EastLakeStation,
		DecaturStation,
		AvondaleStation,
		KensingtonStation,
		IndianCreekStation,
	},
	Gold: {
		AirportStation,
		CollegeParkStation,
		EastPointStation,
		LakewoodStation,
		OaklandCityStation,
		WestEndStation,
		GarnettStation,
		FivePointsStation,
		PeachtreeCenterStation,
		CivicCenterStation,
		NorthAveStation,
		MidtownStation,
		ArtsCenterStation,
		LindberghStation,
		LenoxStation,
		BrookhavenStation,
		ChambleeStation,
		DoravilleStation,
	},
	Red: {
		AirportStation,
		CollegeParkStation,
		EastPointStation,
		LakewoodStation,
		OaklandCityStation,
		WestEndStation,
		GarnettStation,
		FivePointsStation,
		PeachtreeCenterStation,
		CivicCenterStation,
		NorthAveStation,
		MidtownStation,
		ArtsCenterStation,
		LindberghStation,
		BuckheadStation,
		MedicalCenterStation,
		DunwoodyStation,
		SandySpringsStation,
		NorthSpringsStation,
	},
}

//LineDirections provides the directions that are compatible with
//a given line.
var LineDirections = map[Line][]Direction{
	Green: {East, West},
	Blue:  {East, West},
	Gold:  {North, South},
	Red:   {North, South},
}

//Termini allow for lookups of all the terminuseses of the different lines
var Termini = map[Line]map[Direction]Station{
	Green: {
		East: EdgewoodCandlerParkStation,
		West: BankheadStation,
	},
	Blue: {
		East: IndianCreekStation,
		West: HamiltonEHolmesStation,
	},
	Gold: {
		North: DunwoodyStation,
		South: AirportStation,
	},
	Red: {
		North: NorthSpringsStation,
		South: AirportStation,
	},
}
