package martaapi

import (
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ScheduleFinder
type ScheduleFinder interface {
	FindSchedules(ctx context.Context) (io.ReadCloser, error)
	Prefix() string
}

type Schedule struct {
	PrimaryKey     string
	SortKey        string
	Destination    string `json:"DESTINATION"`
	Direction      string `json:"DIRECTION"`
	EventTime      string `json:"EVENT_TIME"`
	Line           string `json:"LINE"`
	NextArrival    string `json:"NEXT_ARR"`
	Station        string `json:"STATION"`
	TrainID        string `json:"TRAIN_ID"`
	WaitingSeconds string `json:"WAITING_SECONDS"`
	WaitingTime    string `json:"WAITING_TIME"`
	TTL            int64  `json:"TTL"`
}

const (
	MartaBaseURI              = "http://developer.itsmarta.com"
	RealtimeTrainTimeEndpoint = "/RealtimeTrain/RestServiceNextTrain/GetRealtimeArrivals"
	BusEndpoint               = "/BRDRestService/RestBusRealTimeService/GetAllBus"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(doer Doer, apiKey string, logger *zap.Logger, endpoint string, prefix string) Client {
	return Client{
		doer,
		apiKey,
		logger,
		endpoint,
		prefix,
	}
}

// Client will hold all of the deps required to find schedules
type Client struct {
	Doer         Doer
	ApiKey       string
	logger       *zap.Logger
	Endpoint     string
	OutputPrefix string
}

func (c Client) Prefix() string {
	return c.OutputPrefix
}

func (c Client) buildRequest(method string, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return req, err
	}
	q := req.URL.Query()
	q.Add("apiKey", c.ApiKey)
	req.URL.RawQuery = q.Encode()
	return req, err
}

// FindSchedules will retrieve a set of schedules
func (c Client) FindSchedules(ctx context.Context) (io.ReadCloser, error) {
	var (
		err error
	)

	path := MartaBaseURI + c.Endpoint

	req, err := c.buildRequest("GET", path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Doer.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
