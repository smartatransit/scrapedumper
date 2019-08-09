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
type ArrivalEstimates map[time.Time]time.Time

//SingleEstimate produces an ArrivalEstimates with one estimate
func SingleEstimate(eventTime time.Time, estimate time.Time) ArrivalEstimates {
	return ArrivalEstimates(map[time.Time]time.Time{eventTime: estimate})
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

//Arrival encodes information about a particular arrival of a train at a station,
//including the actual arrival time and any arrival estimates made beforehand.
type Arrival struct {
	Identifier         string
	RunIdentifier      string
	RunGroupIdentifier string

	MostRecentEventTime time.Time

	Direction           marta.Direction
	Line                marta.Line
	TrainID             string
	RunFirstEventMoment time.Time
	Station             marta.Station

	ArrivalTime      time.Time
	ArrivalEstimates ArrivalEstimates
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
