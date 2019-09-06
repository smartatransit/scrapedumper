package config_test

import (
	"github.com/bipol/scrapedumper/pkg/config"
	"github.com/bipol/scrapedumper/pkg/dumper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildDumper", func() {
	var (
		cfg config.DumpConfig

		result  dumper.Dumper
		callErr error
	)

	JustBeforeEach(func() {
		result, callErr = config.BuildDumper(nil, cfg)
	})

	When("the Kind is RoundRobinKind", func() {
		BeforeEach(func() {
			opts := make(map[string]string)
			opts["s3_bucket_name"] = "bucket-name"

			cfg = config.DumpConfig{
				Kind: config.RoundRobinKind,
				Components: []config.DumpConfig{
					config.DumpConfig{
						Kind:    config.S3DumperKind,
						Options: opts,
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
				cfg.Components[0].Options = nil
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
			})
		})
	})

	When("the Kind is FileDumperKind", func() {
		BeforeEach(func() {
			opts := make(map[string]string)
			opts["local_output_location"] = "/my/dir"
			cfg = config.DumpConfig{
				Kind:    config.FileDumperKind,
				Options: opts,
			}
		})

		It("produces a LocalDumpHandler", func() {
			_, ok := result.(dumper.LocalDumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.Options = nil
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind FILE requested but no file output location provided")))
			})
		})
	})
	When("the Kind is S3DumperKind", func() {
		BeforeEach(func() {
			opts := make(map[string]string)
			opts["s3_bucket_name"] = "bucket-name"
			cfg = config.DumpConfig{
				Kind:    config.S3DumperKind,
				Options: opts,
			}
		})

		It("produces a S3DumpHandler", func() {
			_, ok := result.(dumper.S3DumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.Options = nil
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind S3 requested but no s3 bucket name provided")))
			})
		})
	})
	When("the Kind is DynamoDBDumperKind", func() {
		BeforeEach(func() {
			opts := make(map[string]string)
			opts["dynamo_table_name"] = "my-table"

			cfg = config.DumpConfig{
				Kind:    config.DynamoDBDumperKind,
				Options: opts,
			}
		})

		It("produces a DynamoDumpHandler", func() {
			_, ok := result.(dumper.DynamoDumpHandler)
			Expect(ok).To(BeTrue())
			Expect(callErr).To(BeNil())
		})

		When("the required configs are missing", func() {
			BeforeEach(func() {
				cfg.Options = nil
			})
			It("fails", func() {
				Expect(callErr).To(MatchError(ContainSubstring("dumper kind DYNAMODB requested but no dynamo table name provided")))
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
