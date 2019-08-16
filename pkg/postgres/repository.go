package postgres

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/bipol/scrapedumper/pkg/marta"
)

//Repository implements interactions with Postgres through GORM
//go:generate counterfeiter . Repository
type Repository interface {
	EnsureTables() error

	GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (runFirstEventMoment time.Time, mostRecentEventTime time.Time, err error)
	EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) (err error)
	AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, estimate time.Time) (err error)
	SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, arrival time.Time) (err error)
}

//NewRepository creates a new postgres respotitory
func NewRepository(
	db *sql.DB,
) *RepositoryAgent {
	return &RepositoryAgent{
		DB: db,
	}
}

//RepositoryAgent implements Repository
type RepositoryAgent struct {
	DB *sql.DB
}

//EnsureTables ensures that all necessary tables exist
func (a *RepositoryAgent) EnsureTables() error {
	_, err := a.DB.Exec(`
CREATE TABLE IF NOT EXISTS "arrivals"
(	"identifier" text,
	"run_identifier" text,
	"run_group_identifier" text,
	"most_recent_event_time" timestamp,
	"direction" text,
	"line" text,
	"train_id" text,
	"run_first_event_moment" timestamp,
	"station" text,
	"arrival_time" timestamp,
	"arrival_estimates" text,
	PRIMARY KEY ("identifier")
)`)

	//TODO create indexes?

	return errors.Wrap(err, "failed to ensure arrivals table")
}

//GetLatestRunStartMomentFor from among all runs matching the specified data, this function selects
//the most recent one and returns it's earliest start time (used as part of the run identifier) as
//well as it's most recent one, which is used for determining whether it is stale. If no runs match
//the metadata, it returns two zero time.Time objects and no error
func (a *RepositoryAgent) GetLatestRunStartMomentFor(dir marta.Direction, line marta.Line, trainID string) (runFirstEventMoment time.Time, mostRecentEventTime time.Time, err error) {
	row := a.DB.QueryRow(`
SELECT run_first_event_moment, most_recent_event_time
FROM "arrivals"
WHERE run_group_identifier = $1
ORDER BY run_first_event_moment DESC, most_recent_event_time DESC, "arrivals"."identifier" ASC
LIMIT 1`,
		RunGroupIdentifierFor(dir, line, trainID),
	)

	err = row.Scan(&runFirstEventMoment, &mostRecentEventTime)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "failed to query latest run start moment for dir `%s` line `%s` and train `%s`", dir, line, trainID)
		return
	}
	return
}

//EnsureArrivalRecord ensures that a record exists for the specified arrival
func (a *RepositoryAgent) EnsureArrivalRecord(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) (err error) {
	_, err = a.DB.Exec(`
INSERT INTO "arrivals"
("identifier", "run_identifier", "run_group_identifier", "most_recent_event_time", "direction", "line", "train_id", "run_first_event_moment", "station", "arrival_time", "arrival_estimates")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT DO NOTHING`,
		IdentifierFor(dir, line, trainID, runFirstEventMoment, station),
		RunIdentifierFor(dir, line, trainID, runFirstEventMoment),
		RunGroupIdentifierFor(dir, line, trainID),
		runFirstEventMoment, //most_recent_event_time
		dir,
		line,
		trainID,
		runFirstEventMoment,
		station,
		time.Time{}, //arrival_time
		ArrivalEstimates(map[string]string{}),
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to ensure arrival for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}
	return
}

//AddArrivalEstimate upserts the specified arrival estimate to the arrival record in question
//NOTE: this method is NOT thread-safe.
func (a *RepositoryAgent) AddArrivalEstimate(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, estimate time.Time) (err error) {
	row := a.DB.QueryRow(`
SELECT arrival_estimates
FROM "arrivals"
WHERE "arrivals"."identifier" = $1
LIMIT 1`,
		IdentifierFor(dir, line, trainID, runFirstEventMoment, station),
	)

	var arrivalEstimates ArrivalEstimates
	err = row.Scan(&arrivalEstimates)
	if err != nil {
		err = errors.Wrapf(err, "failed to get arrival for `%s` line `%s` and train `%s`", dir, line, trainID)
		return
	}
	if ok := arrivalEstimates.AddEstimate(eventTime, estimate); !ok {
		return
	}

	_, err = a.DB.Exec(`
UPDATE "arrivals"
SET ("arrival_estimates", "most_recent_event_time")
  = ($1, $2)
WHERE "arrivals"."identifier" = $3`,
		arrivalEstimates,
		eventTime,
		IdentifierFor(dir, line, trainID, runFirstEventMoment, station),
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to add arrival estimate for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return
	}

	return
}

//SetArrivalTime upserts the specified actual arrival time to the arrival record in question
func (a *RepositoryAgent) SetArrivalTime(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station, eventTime time.Time, arrivalTime time.Time) (err error) {
	_, err = a.DB.Exec(`
UPDATE "arrivals"
SET ("arrival_time", "most_recent_event_time") = ($1, $2)
WHERE "arrivals"."identifier" = $3
  AND "arrival_time" = $4`,
		arrivalTime,
		eventTime,
		IdentifierFor(dir, line, trainID, runFirstEventMoment, station),
		time.Time{}, //don't overwrite existing arrival times
	)
	if err != nil {
		err = errors.Wrapf(err, "failed to set arrival time for dir `%s` line `%s` train `%s` first event moment `%s` and station `%s`", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
		return err
	}

	return
}
