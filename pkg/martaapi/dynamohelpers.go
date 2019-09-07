package martaapi

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func ScheduleToWriteRequest(s Schedule, t string) (*dynamodb.WriteRequest, error) {
	date, err := time.Parse(MartaAPIDatetimeFormat, s.EventTime)
	if err != nil {
		return nil, err
	}
	s.PrimaryKey = fmt.Sprintf("%s_%s", s.Station, s.Destination)
	s.SortKey = fmt.Sprintf("%s_%s", date.Format(time.RFC3339), s.TrainID)
	s.TTL = time.Now().Add(30 * 24 * time.Hour).Unix()
	attr, err := dynamodbattribute.MarshalMap(s)
	if err != nil {
		return nil, err
	}
	return &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: attr,
		},
	}, nil
}

func DigestScheduleResponse(r io.Reader, t string) ([]*dynamodb.BatchWriteItemInput, error) {
	var (
		inp []*dynamodb.BatchWriteItemInput
	)
	requestItems := make(map[string][]*dynamodb.WriteRequest)

	dec := json.NewDecoder(r)

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, err
	}

	for dec.More() {
		if len(requestItems[t]) == 25 {
			inp = append(inp, &dynamodb.BatchWriteItemInput{RequestItems: requestItems})
			requestItems = make(map[string][]*dynamodb.WriteRequest)
		}

		var s Schedule
		err = dec.Decode(&s)
		if err != nil {
			return nil, err
		}
		wr, err := ScheduleToWriteRequest(s, t)
		if err != nil {
			return nil, err
		}
		requestItems[t] = append(requestItems[t], wr)
	}

	inp = append(inp, &dynamodb.BatchWriteItemInput{RequestItems: requestItems})

	_, err = dec.Token()
	if err != nil {
		return nil, err
	}

	return inp, err
}
