package postgres

import (
	"errors"
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/jinzhu/gorm"
)

//Repository implements interactions with Postgres through GORM
type Repository interface {
	EnsureTables() error

	GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (startTime time.Time, lastUpdated time.Time, err error)
	EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) error
	AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error
	SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error
}

//NewRepository creates a new postgres respotitory
func NewRepository(
	db *gorm.DB,
) *RepositoryAgent {
	return &RepositoryAgent{
		DB: db,
	}
}

//RepositoryAgent implements Repository
type RepositoryAgent struct {
	DB *gorm.DB
}

//EnsureTables ensures that all necessary tables exist
func (a *RepositoryAgent) EnsureTables() error {
	return a.DB.AutoMigrate(&Arrival{}).Error
}

//GetLatestRunStartMomentFor from among all runs matching the specified data, this function selects
//the most recent one and returns it's earliest start time (used as part of the run identifier) as
//well as it's most recent one, which is used for determining whether it is stale. If no runs match
//the metadata, it returns two zero time.Time objects and no error
func (a *RepositoryAgent) GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (startTime time.Time, lastUpdated time.Time, err error) {
	row := a.DB.Model(&Arrival{}).
		Where("direction = ?", string(dir)).
		Where("line = ?", string(line)).
		Where("train_id = ?", trainID).
		Order("run_first_event_moment DESC").
		Order("most_revent_event_time DESC").Limit(1).
		Select("run_first_event_moment", "most_revent_event_time").Row()
	if err = a.DB.Error; err != nil {
		return time.Time{}, time.Time{}, err
	}

	if row == nil {
		return time.Time{}, time.Time{}, nil
	}

	err = row.Scan(&startTime, &lastUpdated)
	return startTime, lastUpdated, err
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) error {
	err := a.DB.Set("gorm:insert_option", "ON CONFLICT DO NOTHING").
		Create(&Arrival{
			Direction:           dir,
			Line:                line,
			TrainID:             trainID,
			RunFirstEventMoment: runFirstEventMoment,
			Station:             station,
		}).Error
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil
		}
		return err
	}
	return nil
}

//AddArrivalEstimate upserts the specified arrival estimate to the arrival record in question
func (a *RepositoryAgent) AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error {
	if err := a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return nil
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	row := a.DB.Model(&Arrival{}).
		Where("identifier = ?", ident).
		Select("arrival_estimates").Row()
	if a.DB.Error != nil {
		return a.DB.Error
	}

	if row == nil {
		return errors.New("TODO")
	}

	var ests ArrivalEstimates
	err := row.Scan(&ests)
	if err != nil {
		return err
	}

	ests = append(ests, estimate)

	err = a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_estimates", ests).
		Update("most_revent_event_time", estimate.EventTime).Error
	if err != nil {
		return err
	}

	return nil
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) error {
	if err := a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return nil
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	err := a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_time", estimate.EstimatedArrivalTime).
		Update("most_revent_event_time", estimate.EventTime).Error
	if err != nil {
		return err
	}

	return nil
}
