package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/bipol/scrapedumper/pkg/martaapi"
)

//ArrivalEstimates implements SQL marshalling for an array of ArrivalEstimate's
type ArrivalEstimates map[string]string

//AddEstimate adds an estimate. Returns true if the record was new.
func (aes ArrivalEstimates) AddEstimate(eventTime EasternTime, estimate EasternTime) bool {
	evtStr := eventTime.String()
	estStr := estimate.String()

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
func IdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime, station martaapi.Station) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.String(), station)
}

//RunIdentifierFor creates a run identifier for the given metadata
func RunIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string, runFirstEventMoment EasternTime) string {
	return fmt.Sprintf("%s_%s_%s_%s", dir, line, trainID, runFirstEventMoment.String())
}

//RunGroupIdentifierFor creates a run group identifier for the given metadata
func RunGroupIdentifierFor(dir martaapi.Direction, line martaapi.Line, trainID string) string {
	return fmt.Sprintf("%s_%s_%s", dir, line, trainID)
}
