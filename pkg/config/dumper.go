package config

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	"github.com/bipol/scrapedumper/pkg/dumper"
	"github.com/bipol/scrapedumper/pkg/martaapi"
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

//ErrDumperValidationFailed indicates that a dumper's configuration was invalid
var ErrDumperValidationFailed = errors.New("dumper failed to build due to missing args")

//BuildDumper builds the dumper described by the given config option
func BuildDumper(log *zap.Logger, c DumpConfig) (dumper.Dumper, error) {
	switch c.Kind {
	case RoundRobinKind:
		log.Debug("building a roundrobin dumper")
		componentDumpers := make([]dumper.Dumper, len(c.Components))
		if len(c.Components) == 0 {
			log.Debug("no components")
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no components provided: provide components using the config file, a command-line argument, or an environment variable", RoundRobinKind)
		}

		for i := range c.Components {
			log.Debug("building component", i)
			var err error
			componentDumpers[i], err = BuildDumper(log, c.Components[i])
			if err != nil {
				return nil, err
			}
			log.Debug("component built", i)
		}

		return dumper.NewRoundRobinDumpClient(log, componentDumpers...), nil
	case FileDumperKind:
		log.Debug("building a file dumper")
		if c.LocalOutputLocation == "" {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no file output location provided: provide a local output location using the config file, a command-line argument, or an environment variable", FileDumperKind)
		}

		return dumper.NewLocalDumpHandler(c.LocalOutputLocation, log, afero.NewOsFs()), nil
	case DynamoDBDumperKind:
		if c.DynamoTableName == "" {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no dynamo table name provided: provide a dynamo table name using the config file, a command-line argument, or an environment variable", DynamoDBDumperKind)
		}

		dynamoClient := dynamodb.New(session.Must(session.NewSession()))

		return dumper.NewDynamoDumpHandler(log, c.DynamoTableName, dynamoClient, martaapi.DigestScheduleResponse), nil
	case S3DumperKind:
		if c.S3BucketName == "" {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no s3 bucket name provided: provide an s3 bucket name using the config file, a command-line argument, or an environment variable", S3DumperKind)
		}

		s3Manager := s3manager.NewUploaderWithClient(s3.New(session.Must(session.NewSession())))

		return dumper.NewS3DumpHandler(s3Manager, c.S3BucketName, log), nil
	default:
		return nil, errors.Wrapf(ErrDumperValidationFailed, "unsupported dumper kind `%s`", string(c.Kind))
	}
}
