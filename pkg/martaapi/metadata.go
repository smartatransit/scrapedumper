package martaapi

//ClassifySequenceList makes a best effort to determine the Line
// and Direction for this sequence of Stations.
//
//The original values passed into the function may sometimes be
// returned in place of an inferred values if there's no possible
// correct value. However, in some cases where no correct value is
// possible, the algorithm may return a best guess anyways. This
// is because the algorithm tries to exit early if some trick allows
// it to see that there is at most correct answer.
//
//The implementation is meant to be generally robust to missing
// values, although it may not currently be maximally robust.
func ClassifySequenceList(stationSeq []Station, line Line, dir Direction) (Line, Direction) {
	goldScore := 0
	redScore := 0
	ewScore := 0
	nsScore := 0
	for i, station := range stationSeq {
		if station == BankheadStation {
			//if bankhead is the first station, then we're eastbound
			if i == 0 {
				return Green, West
			}

			//otherwise, it must be our destination
			return Green, East
		}

		if _, ok := goldOnlyStations[station]; ok {
			goldScore++
		}
		if _, ok := redOnlyStations[station]; ok {
			redScore++
		}
		if _, ok := ewOnlyStations[station]; ok {
			ewScore++
		}
		if _, ok := nsOnlyStations[station]; ok {
			nsScore++
		}
	}

	//if both _or_ neither got any score, then
	//we've got nothing to go on, so leave it
	//untouched
	if (ewScore == 0) == (nsScore == 0) {
		return line, dir
	}

	if ewScore > 0 {
		//If the sequence contains BankheadStation, then we
		//already exited. That doesn't mean we're not Green
		//line, but if we are there must not be any evidence

		eastScore := directionalityScore(stationSeq, eastCorePositions)
		if eastScore > 0 {
			return Blue, East
		} else if eastScore < 0 {
			return Blue, West
		} else {
			return Blue, dir
		}
	}

	northScore := directionalityScore(stationSeq, northCorePositions)
	if northScore > 0 {
		dir = North
	} else if northScore < 0 {
		dir = South
	}

	if (goldScore == 0) == (redScore == 0) {
		return line, dir
	}

	if goldScore > 0 {
		return Gold, dir
	}

	return Red, dir
}

var goldOnlyStations = map[Station]struct{}{
	LenoxStation:      struct{}{},
	BrookhavenStation: struct{}{},
	ChambleeStation:   struct{}{},
	DoravilleStation:  struct{}{},
}
var redOnlyStations = map[Station]struct{}{
	BuckheadStation:      struct{}{},
	MedicalCenterStation: struct{}{},
	DunwoodyStation:      struct{}{},
	SandySpringsStation:  struct{}{},
	NorthSpringsStation:  struct{}{},
}
var ewOnlyStations = map[Station]struct{}{
	BankheadStation:            struct{}{},
	HamiltonEHolmesStation:     struct{}{},
	WestLakeStation:            struct{}{},
	AshbyStation:               struct{}{},
	VineCityStation:            struct{}{},
	OmniDomeStation:            struct{}{},
	GeorgiaStateStation:        struct{}{},
	KingMemorialStation:        struct{}{},
	InmanParkStation:           struct{}{},
	EdgewoodCandlerParkStation: struct{}{},
	EastLakeStation:            struct{}{},
	DecaturStation:             struct{}{},
	AvondaleStation:            struct{}{},
	KensingtonStation:          struct{}{},
	IndianCreekStation:         struct{}{},
}
var nsOnlyStations = map[Station]struct{}{
	AirportStation:         struct{}{},
	CollegeParkStation:     struct{}{},
	EastPointStation:       struct{}{},
	LakewoodStation:        struct{}{},
	OaklandCityStation:     struct{}{},
	WestEndStation:         struct{}{},
	GarnettStation:         struct{}{},
	PeachtreeCenterStation: struct{}{},
	CivicCenterStation:     struct{}{},
	NorthAveStation:        struct{}{},
	MidtownStation:         struct{}{},
	ArtsCenterStation:      struct{}{},
	LindberghStation:       struct{}{},
	LenoxStation:           struct{}{},
	BrookhavenStation:      struct{}{},
	ChambleeStation:        struct{}{},
	DoravilleStation:       struct{}{},
	BuckheadStation:        struct{}{},
	MedicalCenterStation:   struct{}{},
	DunwoodyStation:        struct{}{},
	SandySpringsStation:    struct{}{},
	NorthSpringsStation:    struct{}{},
}

func directionalityScore(subjectSequence []Station, dir map[Station]int) (score int) {
	return dir[subjectSequence[len(subjectSequence)-1]] - dir[subjectSequence[0]]
}

var eastCorePositions = map[Station]int{
	AshbyStation:               -6,
	VineCityStation:            -5,
	OmniDomeStation:            -4,
	FivePointsStation:          -3,
	GeorgiaStateStation:        -2,
	KingMemorialStation:        -1,
	InmanParkStation:           0,
	EdgewoodCandlerParkStation: +1,
	EastLakeStation:            +2,
	DecaturStation:             +3,
	AvondaleStation:            +4,
	KensingtonStation:          +5,
	IndianCreekStation:         +6,
}
var northCorePositions = map[Station]int{
	AirportStation:         -7,
	CollegeParkStation:     -6,
	EastPointStation:       -5,
	LakewoodStation:        -4,
	OaklandCityStation:     -3,
	WestEndStation:         -2,
	GarnettStation:         -1,
	FivePointsStation:      +1,
	PeachtreeCenterStation: +2,
	CivicCenterStation:     +3,
	NorthAveStation:        +4,
	MidtownStation:         +5,
	ArtsCenterStation:      +6,
	LindberghStation:       +7,
}
