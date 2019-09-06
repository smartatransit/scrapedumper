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

	Components []DumpConfig `json:"components"`
	// TODO: Move these out into just a generic map[string]string?
	Options map[string]string `json:"options"`
}

//ErrDumperValidationFailed indicates that a dumper's configuration was invalid
var ErrDumperValidationFailed = errors.New("dumper failed to build due to missing args")

//BuildDumper builds the dumper described by the given config option
func BuildDumper(log *zap.Logger, c DumpConfig) (dumper.Dumper, error) {
	switch c.Kind {
	case RoundRobinKind:
		componentDumpers := make([]dumper.Dumper, len(c.Components))
		if len(c.Components) == 0 {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no components provided: provide components using the config file, a command-line argument, or an environment variable", RoundRobinKind)
		}

		for i := range c.Components {
			var err error
			componentDumpers[i], err = BuildDumper(log, c.Components[i])
			if err != nil {
				return nil, err
			}
		}

		return dumper.NewRoundRobinDumpClient(log, componentDumpers...), nil
	case FileDumperKind:

		if _, ok := c.Options["local_output_location"]; !ok {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no file output location provided: provide a local output location using the config file, a command-line argument, or an environment variable", FileDumperKind)
		}

		return dumper.NewLocalDumpHandler(c.Options["local_output_location"], log, afero.NewOsFs()), nil
	case DynamoDBDumperKind:
		if _, ok := c.Options["dynamo_table_name"]; !ok {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no dynamo table name provided: provide a dynamo table name using the config file, a command-line argument, or an environment variable", DynamoDBDumperKind)
		}

		dynamoClient := dynamodb.New(session.Must(session.NewSession()))

		return dumper.NewDynamoDumpHandler(log, c.Options["dynamo_table_name"], dynamoClient, martaapi.DigestScheduleResponse), nil
	case S3DumperKind:
		if _, ok := c.Options["s3_bucket_name"]; !ok {
			return nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no s3 bucket name provided: provide an s3 bucket name using the config file, a command-line argument, or an environment variable", S3DumperKind)
		}

		s3Manager := s3manager.NewUploaderWithClient(s3.New(session.Must(session.NewSession())))

		return dumper.NewS3DumpHandler(s3Manager, c.Options["s3_bucket_name"], log), nil
	default:
		return nil, errors.Wrapf(ErrDumperValidationFailed, "unsupported dumper kind `%s`", string(c.Kind))
	}
}
