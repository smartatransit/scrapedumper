package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bipol/scrapedumper/pkg/marta"
	"github.com/jinzhu/gorm"
)

//ArrivalEstimates implements SQL marshalling for an array of ArrivalEstimate's
type ArrivalEstimates map[time.Time]time.Time

//Scan implements the db/sql.Scanner interface
func (ae *ArrivalEstimates) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	err := json.Unmarshal([]byte(str), ae)
	// ae.ensure()
	return err
}

func (ae *ArrivalEstimates) ensure() {
	if ae == nil {
		empty := ArrivalEstimates(map[time.Time]time.Time{})
		ae = &empty
	}
}

//Value implements the db/sql.Valuer interface
func (ae ArrivalEstimates) Value() (driver.Value, error) {
	// ae.ensure()
	bs, err := json.Marshal(ae)
	return string(bs), err
}

//Arrival encodes information about a particular arrival of a train at a station,
//including the actual arrival time and any arrival estimates made beforehand.
type Arrival struct {
	Identifier    string `gorm:"type:text;PRIMARY_KEY"`
	RunIdentifier string `gorm:"type:text"`

	MostRecentEventTime time.Time `gorm:"type:timestamp"`

	Direction           marta.Direction `gorm:"type:text;index:runs"`
	Line                marta.Line      `gorm:"type:text;index:runs"`
	TrainID             string          `gorm:"type:text;index:runs"`
	RunFirstEventMoment time.Time       `gorm:"type:timestamp;index:runs"` //TODO descending or something?
	Station             marta.Station   `gorm:"type:text"`

	ArrivalTime      time.Time        `gorm:"type:timestamp"`
	ArrivalEstimates ArrivalEstimates `gorm:"type:text"` //need a Valuer implementation
}

//BeforeCreate sets up the composite identifiers
func (a *Arrival) BeforeCreate(scope *gorm.Scope) (err error) {
	a.Identifier = IdentifierFor(a.Direction, a.Line, a.TrainID, a.RunFirstEventMoment, a.Station)
	a.RunIdentifier = RunIdentifierFor(a.Direction, a.Line, a.TrainID, a.RunFirstEventMoment)
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
