package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
)

func init() {
	var err error
	EasternTime, err = time.LoadLocation("US/Eastern")
	if err != nil {
		panic("US/Eastern time zone not found")
	}
}

//EasternTime is the eastern timezone, where all MARTA times should be interpreted
var EasternTime *time.Location

//ArrivalEstimates implements SQL marshalling for an array of ArrivalEstimate's
type ArrivalEstimates map[string]string

//AddEstimate adds an estiamte
func (aes ArrivalEstimates) AddEstimate(eventTime time.Time, estimate time.Time) bool {
	evtStr := eventTime.Format(time.RFC3339)
	estStr := estimate.Format(time.RFC3339)

	// if this event time is already in the map, don't overwrite
	if _, ok := aes[evtStr]; ok {
		return false
	}

	aes[evtStr] = estStr

	return true
}

//Scan implements the db/sql.Scanner interface
func (ae *ArrivalEstimates) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	err := json.Unmarshal([]byte(str), ae)
	return err
}

//Value implements the db/sql.Valuer interface
func (ae ArrivalEstimates) Value() (driver.Value, error) {
	bs, err := json.Marshal(ae)
	return string(bs), err
}

//IdentifierFor creates a identifier for the given metadata
func IdentifierFor(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time, station marta.Station) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339), station)
}

//RunIdentifierFor creates a run identifier for the given metadata
func RunIdentifierFor(dir marta.Direction, line marta.Line, trainID string, runFirstEventMoment time.Time) string {
	return fmt.Sprintf("%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.Format(time.RFC3339))
}

//RunGroupIdentifierFor creates a run group identifier for the given metadata
func RunGroupIdentifierFor(dir marta.Direction, line marta.Line, trainID string) string {
	return fmt.Sprintf("%s_%s_%s", dir, line, trainID)
}
