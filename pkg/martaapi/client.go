package martaapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ScheduleFinder
type ScheduleFinder interface {
	FindSchedule() (Schedule, error)
}

type Schedule struct {
	Destination    string `json:"DESTINATION"`
	Direction      string `json:"DIRECTION"`
	EventTime      string `json:"EVENT_TIME"`
	Line           string `json:"LINE"`
	NextArrival    string `json:"NEXT_ARR"`
	Station        string `json:"STATION"`
	TrainID        string `json:"TRAIN_ID"`
	WaitingSeconds string `json:"WAITING_SECONDS"`
	WaitingTime    string `json:"WAITING_TIME"`
}

const (
	MartaBaseURI              = "http://developer.itsmarta.com"
	RealtimeTrainTimeEndpoint = "/RealtimeTrain/RestServiceNextTrain/GetRealtimeArrivals"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client will hold all of the deps required to find schedules
type Client struct {
	Doer   Doer
	ApiKey string
}

// FindSchedules will retrieve a set of schedules
func (c Client) FindSchedules() ([]Schedule, error) {
	var (
		schedules []Schedule
		err       error
	)

	//TODO: need to add api key addition here too
	// lazy, url package
	path := MartaBaseURI + RealtimeTrainTimeEndpoint

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return schedules, err
	}
	resp, err := c.Doer.Do(req)
	if err != nil {
		return schedules, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return schedules, err
	}

	err = json.Unmarshal(body, &schedules)
	if err != nil {
		return schedules, err
	}

	return schedules, nil
}
