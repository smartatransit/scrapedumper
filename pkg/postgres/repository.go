package postgres

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/bipol/scrapedumper/pkg/marta"
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
	rows, err := a.DB.Model(&Arrival{}).
		Where("direction = ?", string(dir)).
		Where("line = ?", string(line)).
		Where("train_id = ?", trainID).
		Order("run_first_event_moment DESC").
		Order("most_recent_event_time DESC").Limit(1).
		Select("run_first_event_moment", "most_recent_event_time").Rows()
	if err = a.DB.Error; err != nil {
		err = errors.Wrapf(err, "failed to query latest run start moment for dir `%s` line `%s` and train `%s`", dir, line, trainID)
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&startTime, &lastUpdated)
		return
	}

	//not found, return time.Time{}, time.Time{}, nil
	return
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) (err error) {
	err = a.DB.Set("gorm:insert_option", "ON CONFLICT DO NOTHING").
		Create(&Arrival{
			Direction:           dir,
			Line:                line,
			TrainID:             trainID,
			RunFirstEventMoment: runFirstEventMoment,
			Station:             station,
		}).Error
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return
		}

		err = errors.Wrapf(err, "failed to ensure arrival for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}
	return nil
}

//AddArrivalEstimate upserts the specified arrival estimate to the arrival record in question
//TODO see if postgres supports array types that can be used to do this in a single query?
func (a *RepositoryAgent) AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) (err error) {
	if err = a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return nil
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	row := a.DB.Model(&Arrival{}).
		Where("identifier = ?", ident).
		Select("arrival_estimates").Row()
	if row == nil {
		err = errors.New("arrival record not found - call EnsureArrivalRecord before attempting to add an arrival estimate")
	}
	if a.DB.Error != nil {
		err = errors.Wrapf(a.DB.Error, "failed to get existing arrival estimates for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	var ests ArrivalEstimates
	err = row.Scan(&ests)
	if err != nil {
		err = errors.Wrapf(err, "failed to scan result for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	ests = append(ests, estimate)

	err = a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_estimates", ests).
		Update("most_recent_event_time", estimate.EventTime).Error
	if err != nil {
		err = errors.Wrapf(err, "failed to add arrival estimate for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	return
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, estimate ArrivalEstimate) (err error) {
	if err = a.EnsureArrivalRecord(dir, line, trainID, runFirstEventMoment, station); err != nil {
		return
	}

	ident := IdentifierFor(dir, line, trainID, runFirstEventMoment, station)
	err = a.DB.Model(&Arrival{Identifier: ident}).
		Update("arrival_time", estimate.EstimatedArrivalTime).
		Update("most_recent_event_time", estimate.EventTime).Error
	if err != nil {
		err = errors.Wrapf(err, "failed to set arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return err
	}

	return
}
