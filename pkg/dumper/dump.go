package dumper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"

	"github.com/smartatransit/scrapedumper/pkg/alias"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/postgres"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Dumper
type Dumper interface {
	Dump(ctx context.Context, r io.Reader, path string) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Uploader
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// RoundRobinDumpClient reads the scrape into disk, and then dumps that result into each dumper synchronously
type RoundRobinDumpClient struct {
	logger  *zap.Logger
	clients []Dumper
}

// NewRoundRobinDumpClient instantiates a new RoundRobin client
func NewRoundRobinDumpClient(logger *zap.Logger, clients ...Dumper) RoundRobinDumpClient {
	return RoundRobinDumpClient{
		logger,
		clients,
	}
}

func (c RoundRobinDumpClient) Dump(ctx context.Context, r io.Reader, path string) error {
	// this could potentially load a lot into memory, but we have to buffer it somehow so that we can read it into multiple
	// dump clients.  This could potentially be better if we use Go pipelining here, but for now i'm keeping it as is
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	for _, client := range c.clients {
		br := bytes.NewReader(b)
		err := client.Dump(ctx, br, path)
		if err != nil {
			return err
		}
	}
	return err
}

// LocalDumpHandler will write a scrape to the local file sysem
type LocalDumpHandler struct {
	path   string
	logger *zap.Logger
	fs     afero.Fs
}

// NewLocalDumpHandler instantiates a new local dump handler
func NewLocalDumpHandler(path string, logger *zap.Logger, fs afero.Fs) LocalDumpHandler {
	return LocalDumpHandler{
		path,
		logger,
		fs,
	}
}

func (c LocalDumpHandler) Dump(ctx context.Context, r io.Reader, path string) error {
	c.logger.Debug(fmt.Sprintf("Local dump to %s", path))
	location := filepath.Join(c.path, path)

	f, err := c.fs.Create(location)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		//avoid leaving a malformed JSON file behind if possible
		if f.Close() == nil {
			if rErr := c.fs.Remove(f.Name()); rErr != nil {
				err = rErr
			}
		}

		return err
	}

	return f.Close()
}

// NewS3DumpHandler instantiates a new S3 dump handler
func NewS3DumpHandler(uploader Uploader, bucket string, logger *zap.Logger) S3DumpHandler {
	return S3DumpHandler{
		uploader,
		bucket,
		logger,
	}
}

// S3DumpHandler will write a scrape to an s3 bucket
type S3DumpHandler struct {
	uploader Uploader
	bucket   string
	logger   *zap.Logger
}

func (c S3DumpHandler) Dump(ctx context.Context, r io.Reader, path string) error {
	c.logger.Debug(fmt.Sprintf("S3 dump to bucket %s, path %s", c.bucket, path))
	_, err := c.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
		Body:   r,
	})
	return err
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DynamoPuter
type DynamoPuter interface {
	BatchWriteItemWithContext(ctx aws.Context, input *dynamodb.BatchWriteItemInput, opts ...request.Option) (*dynamodb.BatchWriteItemOutput, error)
}

func NoOpMarshaller(r io.Reader, s string) ([]*dynamodb.BatchWriteItemInput, error) {
	x := make([]*dynamodb.BatchWriteItemInput, 1)
	return x, nil
}

type DynamoMarshalFunc = func(io.Reader, string) ([]*dynamodb.BatchWriteItemInput, error)

// DynamoDumpHandler will write a scrape into dynamo
// a DynamoMarshalFunc is required, which will transform the io.Reader into a BatchWriteItemInput
type DynamoDumpHandler struct {
	table       string
	logger      *zap.Logger
	dyn         DynamoPuter
	marshalFunc DynamoMarshalFunc
}

// NewDynamoDumpHandler instantiates a new dynamo dump handler
//a marshal func must be provided, which will transform the io.Reader provided into BatchWriteItems
func NewDynamoDumpHandler(logger *zap.Logger, table string, dyn DynamoPuter, marshalFunc DynamoMarshalFunc) DynamoDumpHandler {
	return DynamoDumpHandler{
		table,
		logger,
		dyn,
		marshalFunc,
	}
}

func (c DynamoDumpHandler) Dump(ctx context.Context, r io.Reader, path string) error {
	c.logger.Debug(fmt.Sprintf("Dynamo dump to table %s", c.table))
	inps, err := c.marshalFunc(r, c.table)
	if err != nil {
		return err
	}
	for _, i := range inps {
		_, err = c.dyn.BatchWriteItemWithContext(ctx, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// PostgresDumpHandler will write a scrape into postgres
type PostgresDumpHandler struct {
	logger   *zap.Logger
	upserter postgres.Upserter
	aliaser  alias.AliasLookup
}

// NewPostgresDumpHandler instantiates a new dynamo dump handler
func NewPostgresDumpHandler(logger *zap.Logger, upserter postgres.Upserter, aliaser alias.AliasLookup) PostgresDumpHandler {
	return PostgresDumpHandler{
		logger,
		upserter,
		aliaser,
	}
}

func (c PostgresDumpHandler) Dump(ctx context.Context, r io.Reader, path string) error {
	c.logger.Debug("Postgres dump")

	var records []martaapi.Schedule
	err := json.NewDecoder(r).Decode(&records)
	if err != nil {
		return err
	}

	//group them by train ID
	var runs = map[string][]martaapi.Schedule{}
	for _, rec := range records {
		runs[rec.TrainID] = append(runs[rec.TrainID], rec)
	}

	type correction struct {
		line martaapi.Line
		dir  martaapi.Direction
	}

	var (
		// As we loop over stuff, we're going to keep track of _all_
		// the various line, direction, and station names we come across.
		// Then, we'll loop over those and resolve the database IDs of
		// their respective records by fuzzy-matching.
		seenLineNames      = map[string]struct{}{}
		seenDirectionNames = map[string]struct{}{}
		seenStationNames   = map[string]struct{}{}

		lineNameResolutions      = map[string]*uint{}
		directionNameResolutions = map[string]*uint{}
		stationNameResolutions   = map[string]*uint{}

		// This, on the other hand, is not for correcting the _spelling_
		// of the line and direction names, but for checking whether
		// the train is actually moving in the direction it claims to be
		// moving
		corrections = map[string]correction{}
	)

	for tid, run := range runs {
		stationSeq := make([]martaapi.Station, len(run))
		for i := range run {
			stationSeq[i] = martaapi.Station(run[i].Station)
			seenStationNames[run[i].Station] = struct{}{}
		}
		line, dir := martaapi.ClassifySequenceList(
			stationSeq,
			martaapi.Line(run[0].Line),
			martaapi.Direction(run[0].Direction),
		)

		seenLineNames[string(line)] = struct{}{}
		seenDirectionNames[string(dir)] = struct{}{}

		corrections[tid] = correction{
			line: line,
			dir:  dir,
		}
	}

	if c.aliaser != nil {
		lineNameResolutions = c.ResolveAliases(ctx, seenLineNames, "lines")
		directionNameResolutions = c.ResolveAliases(ctx, seenDirectionNames, "directions")
		stationNameResolutions = c.ResolveAliases(ctx, seenStationNames, "stations")
	}

	for _, rec := range records {
		corr := corrections[rec.TrainID]
		var lineID, stationID, directionID *uint

		if c.aliaser != nil {
			lineID = lineNameResolutions[string(corr.line)]
			stationID = directionNameResolutions[string(corr.dir)]
			directionID = stationNameResolutions[rec.Station]
		}

		err := c.upserter.AddRecordToDatabase(
			rec,
			martaapi.Line(corr.line),
			martaapi.Direction(corr.dir),

			lineID,
			stationID,
			directionID,
		)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to upsert MARTA API response to postgres: %s", err.Error()))
		}
	}

	return nil
}

func (c PostgresDumpHandler) ResolveAliases(ctx context.Context, queries map[string]struct{}, kind string) map[string]*uint {
	resolutions := map[string]*uint{}
	for q := range queries {
		resolution, err := c.aliaser.FindNamedElementByRoughName(kind, q)
		if err != nil {
			c.logger.Error(fmt.Sprintf("couldn't match %s name `%s`: %s", kind, q, err.Error()))
			continue
		}

		resolutions[q] = resolution
	}

	return resolutions
}
