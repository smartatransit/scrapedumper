package postgres

import (
	"database/sql/driver"
	"fmt"
	"time"
)

func init() {
	var err error
	EasternTimeZone, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic("US/Eastern time zone not found")
	}
}

//EasternTimeZone is the eastern timezone, where all MARTA times should be interpreted
var EasternTimeZone *time.Location

//EasternTime stores times in the Eastern timezone in postgres as strings, so that
//integrity is guaranteed regardless of the timezone of the connection.
type EasternTime time.Time

//String provides an RFC3339 representation of this EasternTime
func (ae EasternTime) String() string {
	return time.Time(ae).Format(time.RFC3339)
}

//ParseEasternTime parses a timestamp in the eastern timezone
func ParseEasternTime(str string) (EasternTime, error) {
	t, err := time.ParseInLocation(time.RFC3339, str, EasternTimeZone)
	return EasternTime(t), err
}

//Scan implements the db/sql.Scanner interface
func (ae *EasternTime) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	var err error
	*ae, err = ParseEasternTime(str)
	return err
}

//Value implements the db/sql.Valuer interface
func (ae EasternTime) Value() (driver.Value, error) {
	ae = EasternTime(time.Time(ae).In(EasternTimeZone))
	return ae.String(), nil
}
