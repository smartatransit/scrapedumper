package martaapi

//Direction enumerates all valid MARTA direction codes
type Direction string

const (
	North Direction = "N"
	South Direction = "S"
	East  Direction = "E"
	West  Direction = "W"
)

//Directions is for checking whether a string represents a valid Direction
var Directions = map[Direction]struct{}{
	North: struct{}{},
	South: struct{}{},
	East:  struct{}{},
	West:  struct{}{},
}

//Line enumerates all valid MARTA line names
type Line string

const (
	Green Line = "GREEN"
	Blue  Line = "BLUE"
	Gold  Line = "GOLD"
	Red   Line = "RED"
)

//Lines is for checking whether a string represents a valid Line
var Lines = map[Line]struct{}{
	Green: struct{}{},
	Blue:  struct{}{},
	Gold:  struct{}{},
	Red:   struct{}{},
}

//Station enumerates all valid MARTA station names
type Station string

const (
	FivePointsStation          Station = "FIVE POINTS STATION"
	OaklandCityStation         Station = "OAKLAND CITY STATION"
	AirportStation             Station = "AIRPORT STATION"
	BrookhavenStation          Station = "BROOKHAVEN STATION"
	KensingtonStation          Station = "KENSINGTON STATION"
	OmniDomeStation            Station = "OMNI DOME STATION"
	DecaturStation             Station = "DECATUR STATION"
	InmanParkStation           Station = "INMAN PARK STATION"
	SandySpringsStation        Station = "SANDY SPRINGS STATION"
	IndianCreekStation         Station = "INDIAN CREEK STATION"
	DunwoodyStation            Station = "DUNWOODY STATION"
	KingMemorialStation        Station = "KING MEMORIAL STATION"
	VineCityStation            Station = "VINE CITY STATION"
	NorthSpringsStation        Station = "NORTH SPRINGS STATION"
	MedicalCenterStation       Station = "MEDICAL CENTER STATION"
	ArtsCenterStation          Station = "ARTS CENTER STATION"
	EastPointStation           Station = "EAST POINT STATION"
	CivicCenterStation         Station = "CIVIC CENTER STATION"
	CollegeParkStation         Station = "COLLEGE PARK STATION"
	ChambleeStation            Station = "CHAMBLEE STATION"
	AvondaleStation            Station = "AVONDALE STATION"
	LindberghStation           Station = "LINDBERGH STATION"
	GarnettStation             Station = "GARNETT STATION"
	PeachtreeCenterStation     Station = "PEACHTREE CENTER STATION"
	WestLakeStation            Station = "WEST LAKE STATION"
	EastLakeStation            Station = "EAST LAKE STATION"
	GeorgiaStateStation        Station = "GEORGIA STATE STATION"
	BankheadStation            Station = "BANKHEAD STATION"
	BuckheadStation            Station = "BUCKHEAD STATION"
	WestEndStation             Station = "WEST END STATION"
	LakewoodStation            Station = "LAKEWOOD STATION"
	MidtownStation             Station = "MIDTOWN STATION"
	DoravilleStation           Station = "DORAVILLE STATION"
	LenoxStation               Station = "LENOX STATION"
	EdgewoodCandlerParkStation Station = "EDGEWOOD CANDLER PARK STATION"
	NorthAveStation            Station = "NORTH AVE STATION"
	AshbyStation               Station = "ASHBY STATION"
	HamiltonEHolmesStation     Station = "HAMILTON E HOLMES STATION"
)

//Stations is for checking whether a string represents a valid station
var Stations = map[Station]struct{}{
	FivePointsStation:          struct{}{},
	OaklandCityStation:         struct{}{},
	AirportStation:             struct{}{},
	BrookhavenStation:          struct{}{},
	KensingtonStation:          struct{}{},
	OmniDomeStation:            struct{}{},
	DecaturStation:             struct{}{},
	InmanParkStation:           struct{}{},
	SandySpringsStation:        struct{}{},
	IndianCreekStation:         struct{}{},
	DunwoodyStation:            struct{}{},
	KingMemorialStation:        struct{}{},
	VineCityStation:            struct{}{},
	NorthSpringsStation:        struct{}{},
	MedicalCenterStation:       struct{}{},
	ArtsCenterStation:          struct{}{},
	EastPointStation:           struct{}{},
	CivicCenterStation:         struct{}{},
	CollegeParkStation:         struct{}{},
	ChambleeStation:            struct{}{},
	AvondaleStation:            struct{}{},
	LindberghStation:           struct{}{},
	GarnettStation:             struct{}{},
	PeachtreeCenterStation:     struct{}{},
	WestLakeStation:            struct{}{},
	EastLakeStation:            struct{}{},
	GeorgiaStateStation:        struct{}{},
	BankheadStation:            struct{}{},
	BuckheadStation:            struct{}{},
	WestEndStation:             struct{}{},
	LakewoodStation:            struct{}{},
	MidtownStation:             struct{}{},
	DoravilleStation:           struct{}{},
	LenoxStation:               struct{}{},
	EdgewoodCandlerParkStation: struct{}{},
	NorthAveStation:            struct{}{},
	AshbyStation:               struct{}{},
	HamiltonEHolmesStation:     struct{}{},
}

//LineStations provides the stations on the line, in the order
//specified by LineDirections[line][0].
var LineStations = map[Line][]Station{
	Green: []Station{
		BankheadStation,
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
	Blue: []Station{
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
	Gold: []Station{
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
	Red: []Station{
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
	Green: []Direction{East, West},
	Blue:  []Direction{East, West},
	Gold:  []Direction{North, South},
	Red:   []Direction{North, South},
}
