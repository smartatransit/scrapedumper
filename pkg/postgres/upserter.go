package postgres

import (
	"time"

	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/pkg/errors"
)

//Upserter upserts a record to the database, while attempting to
//reconcile separate records from the same train run
//go:generate counterfeiter . Upserter
type Upserter interface {
	AddRecordToDatabase(martaapi.Schedule) error
}

//NewUpserter creates a new postgres upserter
func NewUpserter(
	repo Repository,
	runLifetime time.Duration,
) *UpserterAgent {
	return &UpserterAgent{
		repo:        repo,
		runLifetime: runLifetime,
	}
}

//UpserterAgent implements Upserter
type UpserterAgent struct {
	repo        Repository
	runLifetime time.Duration
}

func newRunRequired(
	runFirstEventMoment EasternTime,
	mostRecentEventMoment EasternTime,
	eventTime time.Time,
	runLifetime time.Duration,
) bool {
	return time.Time(runFirstEventMoment) == (time.Time{}) ||
		time.Time(mostRecentEventMoment).Before(eventTime.Add(-runLifetime))
}

//AddRecordToDatabase upserts a record to the database, while
//attempting to reconcile separate records from the same train run
func (a *UpserterAgent) AddRecordToDatabase(rec martaapi.Schedule) (err error) {
	goEventTime, err := time.ParseInLocation(martaapi.MartaAPIDatetimeFormat, rec.EventTime, EasternTimeZone)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse record event time `%s`", rec.EventTime)
		return
	}
	eventTime := EasternTime(goEventTime)

	runFirstEventMoment, mostRecentEventMoment, err := a.repo.GetLatestRunStartMomentFor(martaapi.Direction(rec.Direction), martaapi.Line(rec.Line), rec.TrainID, eventTime)
	if err != nil {
		err = errors.Wrapf(err, "failed to get latest run start moment for record `%s`", rec.String())
		return
	}

	//if the run didn't match, or if the latest run is stale,
	//then this is the start of a new run
	if newRunRequired(
		runFirstEventMoment,
		mostRecentEventMoment,
		goEventTime,
		a.runLifetime,
	) {
		runFirstEventMoment = eventTime

		//
		//
		//
		// TODO: the entire paradigm of dumpers that handle a single
		//   martaapi.Schedule needs to change - ClassifySequenceList
		//   needs to see the whole picture.
		//
		// Idea: change the dumper interface to accept a []martaapi.Schedule
		//   then create a ScheduleDumper that only uses martaapi.Schedule,
		//   and a new Dumper implementation that naively uses a ScheduleDumper
		//
		//
		//

		stationSeq := make([]martaapi.Station, len(rec))
		correctedLine, correctedDirection := martaapi.ClassifySequenceList(
			nil, //TODO
			martaapi.Line(rec.Line),
			martaapi.Direction(rec.Direction),
		)

		if err = a.repo.CreateRunRecord(
			martaapi.Direction(rec.Direction),
			martaapi.Line(rec.Line),
			rec.TrainID,
			runFirstEventMoment,
			correctedLine,
			correctedDirection,
		); err != nil {
			err = errors.Wrapf(err, "failed to create run record for `%s`", rec.String())
			return
		}
	}

	if err = a.repo.EnsureArrivalRecord(
		martaapi.Direction(rec.Direction),
		martaapi.Line(rec.Line),
		rec.TrainID,
		runFirstEventMoment,
		martaapi.Station(rec.Station),
	); err != nil {
		err = errors.Wrapf(err, "failed to ensure pre-existing arrival record for `%s`", rec.String())
		return
	}

	if rec.HasArrived() {
		//NOTE this is a good first pass, but it is a potential source of error
		//to assume that the arrival time equals the first event time where the
		//train appears to have arrived. There may be smarter ways to infer the
		//arrival moment.
		arrivalTime := eventTime

		err = a.repo.SetArrivalTime(
			martaapi.Direction(rec.Direction),
			martaapi.Line(rec.Line),
			rec.TrainID,
			runFirstEventMoment,
			martaapi.Station(rec.Station),
			eventTime,
			arrivalTime,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to set arrival time from record `%s`", rec.String())
			return
		}
	} else if rec.IsArriving() {
		// we don't have an estimate to add, but we also don't want to set the arrival
		// time until the state changes again, so for now we ignore the record.
	} else {
		var goEstimate time.Time
		goEstimate, err = time.ParseInLocation(martaapi.MartaAPITimeFormat, rec.NextArrival, EasternTimeZone)
		if err != nil {
			err = errors.Wrapf(err, "failed to parse record estimated arrival time `%s`", rec.NextArrival)
			return
		}

		//take the time part of estimate together with the date part of runFirstEventMoment
		goRunFirstEventMoment := time.Time(runFirstEventMoment)
		estimate := EasternTime(time.Date(
			goRunFirstEventMoment.Year(), goRunFirstEventMoment.Month(), goRunFirstEventMoment.Day(),
			goEstimate.Hour(), goEstimate.Minute(), goEstimate.Second(), goEstimate.Nanosecond(),
			EasternTimeZone,
		))

		err = a.repo.AddArrivalEstimate(
			martaapi.Direction(rec.Direction),
			martaapi.Line(rec.Line),
			rec.TrainID,
			runFirstEventMoment,
			martaapi.Station(rec.Station),
			eventTime,
			estimate,
		)
		if err != nil {
			err = errors.Wrapf(err, "failed to add arrival estimate from record `%s`", rec.String())
			return
		}
	}

	return nil
}
