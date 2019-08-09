package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/jinzhu/gorm"
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
	Identifier         string `gorm:"type:text;PRIMARY_KEY"`
	RunIdentifier      string `gorm:"type:text;index:run_id_idx"`
	RunGroupIdentifier string `gorm:"type:text;index:run_group_id_idx"`

	MostRecentEventTime time.Time `gorm:"type:timestamp"`

	Direction           marta.Direction `gorm:"type:text;index:runs"`
	Line                marta.Line      `gorm:"type:text;index:runs"`
	TrainID             string          `gorm:"type:text;index:runs"`
	RunFirstEventMoment time.Time       `gorm:"type:timestamp;index:runs"` //TODO descending or something?
	Station             marta.Station   `gorm:"type:text"`

	ArrivalTime      time.Time        `gorm:"type:timestamp"`
	ArrivalEstimates ArrivalEstimates `gorm:"type:jsonb"` //need a Valuer implementation
}

//BeforeCreate sets up the composite identifiers
func (a *Arrival) BeforeCreate(scope *gorm.Scope) (err error) {
	a.Identifier = IdentifierFor(a.Direction, a.Line, a.TrainID, a.RunFirstEventMoment, a.Station)
	a.RunIdentifier = RunIdentifierFor(a.Direction, a.Line, a.TrainID, a.RunFirstEventMoment)
	a.RunGroupIdentifier = RunGroupIdentifierFor(a.Direction, a.Line, a.TrainID)
	return
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
