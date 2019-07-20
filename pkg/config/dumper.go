package config

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

//DumperKind is an enum type used to specify which type of dumper is being configured
type DumperKind string

const (
	//RoundRobinKind creates a dumper that multiplexes to several other dumpers
	RoundRobinKind DumperKind = "ROUND_ROBIN"
	//FileDumperKind creates a dumper that writes to a local filesystem
	FileDumperKind DumperKind = "FILE"
	//S3DumperKind creates a dumper that writes to an S3 bucket
	S3DumperKind DumperKind = "S3"
	//DynamoDBDumperKind creates a dumper that writes to a dynamodb table
	DynamoDBDumperKind DumperKind = "DYNAMODB"
)

//DumpConfig specifies configuration for one dumper
type DumpConfig struct {
	Kind DumperKind `json:"kind"`

	Components          []DumpConfig `json:"components"`
	LocalOutputLocation string       `json:"local_output_location"`
	S3BucketName        string       `json:"s3_bucket_name"`
	DynamoTableName     string       `json:"dynamo_table_name"`
}

//BuildDumper builds the dumper described by the given config option
func BuildDumper(log *zap.Logger, c DumpConfig) (dumper.Dumper, error) {
	switch c.Kind {
	case RoundRobinKind:
		componentDumpers := make([]dumper.Dumper, len(c.Components))
		for i := range c.Components {
			var err error
			componentDumpers[i], err = BuildDumper(log, c.Components[i])
			if err != nil {
				return nil, err
			}
		}

		return dumper.NewRoundRobinDumpClient(log, componentDumpers...), nil
	case FileDumperKind:
		var localOutputLocation string
		if c.LocalOutputLocation == "" {
			return nil, errors.New("dumper kind FILE requested but no file output location provided: provide a local output location using the config file, a command-line argument, or an environment variable")
		} else {
			localOutputLocation = c.LocalOutputLocation
		}

		return dumper.NewLocalDumpHandler(localOutputLocation, log, afero.NewOsFs()), nil
	case DynamoDBDumperKind:
		var dynamoTableName string
		if c.DynamoTableName == "" {
			return nil, errors.New("dumper kind DYNAMODB requested but no dynamo table name provided: provide a dynamo table name using the config file, a command-line argument, or an environment variable")
		} else {
			dynamoTableName = c.DynamoTableName
		}

		dynamoClient := dynamodb.New(session.Must(session.NewSession()))

		//TODO expose flexibility in the marshalFunc?
		return dumper.NewDynamoDumpHandler(log, dynamoTableName, dynamoClient, martaapi.DigestScheduleResponse), nil
	case S3DumperKind:
		var s3BucketName string
		if c.S3BucketName == "" {
			return nil, errors.New("dumper kind S3 requested but no s3 bucket name provided: provide an s3 bucket name using the config file, a command-line argument, or an environment variable")
		} else {
			s3BucketName = c.S3BucketName
		}

		s3Manager := s3manager.NewUploaderWithClient(s3.New(session.Must(session.NewSession())))

		return dumper.NewS3DumpHandler(s3Manager, s3BucketName, log), nil
	default:
		return nil, fmt.Errorf("unsupported dumper kind `%s`", string(c.Kind))
	}
}
