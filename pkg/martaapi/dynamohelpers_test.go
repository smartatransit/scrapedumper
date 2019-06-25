package martaapi_test

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/bipol/scrapedumper/pkg/martaapi"
)

var _ = Describe("Dynamohelpers", func() {
	Context("DigestScheduleResponse", func() {
		var (
			batchInput []*dynamodb.BatchWriteItemInput
			err        error
			r          io.Reader
		)
		BeforeEach(func() {
			err = nil
			batchInput = nil
			r = strings.NewReader(martaapi.ValidScheduleJSON)
		})
		JustBeforeEach(func() {
			batchInput, err = martaapi.DigestScheduleResponse(r, "t")
		})
		When("given an invalid response", func() {
			BeforeEach(func() {
				r = strings.NewReader("")
			})
			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		When("given a correct json", func() {
			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})
			It("returns correct request item", func() {
				Expect(len(batchInput[0].RequestItems["t"])).To(Equal(2))
				Expect(batchInput[0].RequestItems["t"][0].PutRequest.Item).To(MatchKeys(IgnoreExtras, Keys{
					"STATION":         PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("LAKEWOOD STATION"))})),
					"WAITING_SECONDS": PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("-16"))})),
					"WAITING_TIME":    PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("Boarding"))})),
					"PrimaryKey":      PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("LAKEWOOD STATION_Doraville_2019-05-11"))})),
					"LINE":            PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("GOLD"))})),
					"NEXT_ARR":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("05:48:14 PM"))})),
					"TRAIN_ID":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("304326"))})),
					"DESTINATION":     PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("Doraville"))})),
					"DIRECTION":       PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("N"))})),
				}))
				Expect(batchInput[0].RequestItems["t"][1].PutRequest.Item).To(MatchKeys(IgnoreExtras, Keys{
					"STATION":         PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("KENSINGTON STATION"))})),
					"WAITING_SECONDS": PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("-4"))})),
					"WAITING_TIME":    PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("Boarding"))})),
					"PrimaryKey":      PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("KENSINGTON STATION_Hamilton E Holmes_2019-05-11"))})),
					"LINE":            PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("BLUE"))})),
					"NEXT_ARR":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("05:48:26 PM"))})),
					"TRAIN_ID":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("103206"))})),
					"DESTINATION":     PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("Hamilton E Holmes"))})),
					"DIRECTION":       PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("W"))})),
				}))
			})
		})
	})
	Context("ScheduleToWriteRequest", func() {
		var (
			s   martaapi.Schedule
			wr  *dynamodb.WriteRequest
			err error
		)
		BeforeEach(func() {
			err = nil
			s = martaapi.Schedule{
				Destination:    "destination",
				Direction:      "direction",
				EventTime:      "5/14/2019 5:50:52 PM",
				Line:           "red",
				NextArrival:    "05:51:10 PM",
				Station:        "station",
				TrainID:        "train_id",
				WaitingSeconds: "-10",
				WaitingTime:    "Boarding",
			}
			wr = nil
		})
		JustBeforeEach(func() {
			wr, err = martaapi.ScheduleToWriteRequest(s, "table")
		})
		When("given an invalid eventtime", func() {
			BeforeEach(func() {
				s.EventTime = "invalid"
			})
			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})
		When("given a valid schedule", func() {
			It("returns valid fields on the write request", func() {
				item := wr.PutRequest.Item
				ttl := item["TTL"].N
				i, err := strconv.ParseInt(*ttl, 10, 64)
				Expect(err).To(BeNil())
				Expect(time.Unix(i, 0)).To(BeTemporally("~", time.Now().Add(30*24*time.Hour), time.Hour))
				Expect(item).To(MatchKeys(IgnoreExtras, Keys{
					"STATION":         PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("station"))})),
					"WAITING_SECONDS": PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("-10"))})),
					"WAITING_TIME":    PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("Boarding"))})),
					"PrimaryKey":      PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("station_destination_2019-05-14"))})),
					"LINE":            PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("red"))})),
					"NEXT_ARR":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("05:51:10 PM"))})),
					"TRAIN_ID":        PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("train_id"))})),
					"DESTINATION":     PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("destination"))})),
					"DIRECTION":       PointTo(MatchFields(IgnoreExtras, Fields{"S": Equal(aws.String("direction"))})),
				}))
			})
			It("does not return an error", func() {
				Expect(err).To(BeNil())
			})
		})
	})

})
