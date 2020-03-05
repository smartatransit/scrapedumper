package config_test

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/smartatransit/scrapedumper/pkg/config"
	"github.com/smartatransit/scrapedumper/pkg/config/configfakes"
	"github.com/smartatransit/scrapedumper/pkg/dumper"
	"github.com/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildDumper", func() {
	var (
		cfg     config.DumpConfig
		sqlOpen *configfakes.FakeSQLOpener

		result  dumper.Dumper
		callErr error
	)

	BeforeEach(func() {
		sqlOpen = &configfakes.FakeSQLOpener{}
	})

	JustBeforeEach(func() {
		result, _, callErr = config.BuildDumper(nil, sqlOpen.Spy, cfg)
	})

	When("the Kind is RoundRobinKind", func() {
		BeforeEach(func() {
			cfg = config.DumpConfig{
				Kind: config.RoundRobinKind,
				Components: []config.DumpConfig{
					config.DumpConfig{
						Kind:         config.S3DumperKind,
						S3BucketName: "my-bucket",
					},
				},
			}
		})

		It("produces a RoundRobinDumper", func() {
			_, ok := result.(dumper.RoundRobinDumpClient)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("no components are specified", func() {
			BeforeEach(func() {
				cfg.Components = []config.DumpConfig{}
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind ROUND_ROBIN requested but no components provided")))
			})
		})

		When("one of the dumpers can't be build", func() {
			BeforeEach(func() {
				cfg.Components[0].S3BucketName = ""
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
			})
		})
	})

	When("the Kind is FileDumperKind", func() {
		BeforeEach(func() {
			cfg = config.DumpConfig{
				Kind:                config.FileDumperKind,
				LocalOutputLocation: "/my/dir",
			}
		})

		It("produces a LocalDumpHandler", func() {
			_, ok := result.(dumper.LocalDumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.LocalOutputLocation = ""
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind FILE requested but no file output location provided")))
			})
		})
	})
	When("the Kind is S3DumperKind", func() {
		BeforeEach(func() {
			cfg = config.DumpConfig{
				Kind:         config.S3DumperKind,
				S3BucketName: "bucket-name",
			}
		})

		It("produces a S3DumpHandler", func() {
			_, ok := result.(dumper.S3DumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.S3BucketName = ""
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
			})
		})
	})
	When("the Kind is DynamoDBDumperKind", func() {
		BeforeEach(func() {
			cfg = config.DumpConfig{
				Kind:            config.DynamoDBDumperKind,
				DynamoTableName: "my-table",
			}
		})

		It("produces a DynamoDumpHandler", func() {
			_, ok := result.(dumper.DynamoDumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.DynamoTableName = ""
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind DYNAMODB requested but no dynamo table name provided")))
			})
		})
	})
	When("the Kind is PostgresDumperKind", func() {
		var (
			db    *sql.DB
			smock sqlmock.Sqlmock
		)

		BeforeEach(func() {
			cfg = config.DumpConfig{
				Kind:                     config.PostgresDumperKind,
				PostgresConnectionString: "postgres://host/db?option=value",
			}

			var err error
			db, smock, err = sqlmock.New()
			Expect(err).To(BeNil())

			sqlOpen.Returns(db, nil)
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.PostgresConnectionString = ""
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind POSTGRES requested but no postgres connection string provided: provide a postgres connection string using the config file, a command-line argument, or an environment variable")))
			})
		})

		When("the database connection can't be opened", func() {
			BeforeEach(func() {
				sqlOpen.Returns(nil, errors.New("open failed"))
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("failed connecting to postgres database")))
			})
		})

		When("EnsureTables is executed", func() {
			var exec *sqlmock.ExpectedExec
			BeforeEach(func() {
				exec = smock.ExpectExec(".*")
			})
			AfterEach(func() {
				Expect(smock.ExpectationsWereMet()).To(BeNil())
			})

			When("the EnsureTables call fails", func() {
				BeforeEach(func() {
					exec.WillReturnError(errors.New("CREATE TABLE failed"))
				})
				It("fails", func() {
					Expect(callErr).To(MatchError(ContainSubstring("failed to ensure postgres tables")))
				})
			})

			When("all goes well", func() {
				BeforeEach(func() {
					exec.WillReturnResult(sqlmock.NewResult(0, 0))

					//we're also expecting a second query, to create the index
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
					smock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
				})
				It("produces a PostgresDumpHandler", func() {
					Expect(callErr).To(BeNil())
					_, ok := result.(dumper.PostgresDumpHandler)
					Expect(ok).To(BeTrue())
				})
			})
		})
	})
	When("the Kind is not recognized", func() {
		BeforeEach(func() {
			cfg.Kind = ""
		})

		It("returns an error", func() {
			Expect(callErr).To(MatchError(ContainSubstring("unsupported dumper kind ``")))
		})
	})
})
