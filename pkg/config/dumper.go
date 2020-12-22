package config

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
	gpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/smartatransit/scrapedumper/pkg/alias"
	"github.com/smartatransit/scrapedumper/pkg/dumper"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/postgres"

	//database/sql driver
	_ "github.com/lib/pq"
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
	//PostgresDumperKind creates a dumper that writes to a postgres table
	PostgresDumperKind DumperKind = "POSTGRES"
)

//DumpConfig specifies configuration for one dumper
type DumpConfig struct {
	Kind DumperKind `json:"kind"`

	Components               []DumpConfig `json:"components"`
	LocalOutputLocation      string       `json:"local_output_location"`
	S3BucketName             string       `json:"s3_bucket_name"`
	DynamoTableName          string       `json:"dynamo_table_name"`
	PostgresConnectionString string       `json:"postgres_connection_string"`
	ThirdRailContext         bool         `json:"third_rail_context"`
}

//ErrDumperValidationFailed indicates that a dumper's configuration was invalid
var ErrDumperValidationFailed = errors.New("dumper failed to build due to missing args")

//SQLOpener mocks out the database/sql.Open functino for testing
//go:generate counterfeiter . SQLOpener
type SQLOpener func(string, string) (*sql.DB, error)

//CleanupFunc is used for cleaning up persistent resources used by dumpers
type CleanupFunc func() error

//NoopCleanup is a CleanupFunc that does nothing
var NoopCleanup = func() (_ error) { return }

//NewRoundRobinCleanup executes all cleanup funcs even if some
//fail, and returns the last error.
func NewRoundRobinCleanup(comps []CleanupFunc) CleanupFunc {
	return func() (err error) {
		for _, f := range comps {
			if tmpErr := f(); tmpErr != nil {
				err = tmpErr
			}
		}
		return
	}
}

//BuildDumper builds the dumper described by the given config option.
//If no SQL-based dumpers are to be used, then `sqlOpen` is not required.
func BuildDumper(
	log *zap.Logger,
	sqlOpen SQLOpener,
	c DumpConfig,
) (dumper.Dumper, CleanupFunc, error) {
	switch c.Kind {
	case RoundRobinKind:
		componentDumpers := make([]dumper.Dumper, len(c.Components))
		componentCleanups := make([]CleanupFunc, len(c.Components))
		if len(c.Components) == 0 {
			return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no components provided: provide components using the config file, a command-line argument, or an environment variable", RoundRobinKind)
		}

		for i := range c.Components {
			var err error
			componentDumpers[i], componentCleanups[i], err = BuildDumper(log, sqlOpen, c.Components[i])
			if err != nil {
				return nil, nil, err
			}
		}

		return dumper.NewRoundRobinDumpClient(log, componentDumpers...),
			NewRoundRobinCleanup(componentCleanups), nil
	case FileDumperKind:
		if c.LocalOutputLocation == "" {
			return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no file output location provided: provide a local output location using the config file, a command-line argument, or an environment variable", FileDumperKind)
		}

		return dumper.NewLocalDumpHandler(c.LocalOutputLocation, log, afero.NewOsFs()), NoopCleanup, nil
	case DynamoDBDumperKind:
		if c.DynamoTableName == "" {
			return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no dynamo table name provided: provide a dynamo table name using the config file, a command-line argument, or an environment variable", DynamoDBDumperKind)
		}

		dynamoClient := dynamodb.New(session.Must(session.NewSession()))

		return dumper.NewDynamoDumpHandler(log, c.DynamoTableName, dynamoClient, martaapi.DigestScheduleResponse), NoopCleanup, nil
	case S3DumperKind:
		if c.S3BucketName == "" {
			return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no s3 bucket name provided: provide an s3 bucket name using the config file, a command-line argument, or an environment variable", S3DumperKind)
		}

		s3Manager := s3manager.NewUploaderWithClient(s3.New(session.Must(session.NewSession())))

		return dumper.NewS3DumpHandler(s3Manager, c.S3BucketName, log), NoopCleanup, nil
	case PostgresDumperKind:
		if c.PostgresConnectionString == "" {
			return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "dumper kind %s requested but no postgres connection string provided: provide a postgres connection string using the config file, a command-line argument, or an environment variable", PostgresDumperKind)
		}

		db, err := sqlOpen("postgres", c.PostgresConnectionString)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed connecting to postgres database")
		}

		repo := postgres.NewRepository(log, db)
		err = repo.EnsureTables(c.ThirdRailContext)
		if err != nil {
			db.Close()
			return nil, nil, errors.Wrap(err, "failed to ensure postgres tables")
		}

		// This feature won't work properly unless scrapedumper is deployed inside
		// of a third-rail database
		var aliaser alias.AliasLookup
		if c.ThirdRailContext {
			gormDB, err := gorm.Open(gpostgres.New(gpostgres.Config{
				Conn: db,
			}), &gorm.Config{})
			if err != nil {
				db.Close()
				return nil, nil, errors.Wrap(err, "failed to open gorm DB connection")
			}

			aliaser = alias.New(gormDB)
		}

		upserter := postgres.NewUpserter(repo, time.Hour, c.ThirdRailContext)
		return dumper.NewPostgresDumpHandler(log, upserter, aliaser), db.Close, nil
	default:
		return nil, nil, errors.Wrapf(ErrDumperValidationFailed, "unsupported dumper kind `%s`", string(c.Kind))
	}
}
