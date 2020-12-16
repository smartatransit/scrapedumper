package dumper_test

import (
	"context"
	"errors"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/smartatransit/scrapedumper/pkg/dumper"
	"github.com/smartatransit/scrapedumper/pkg/dumper/dumperfakes"
	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/postgres/postgresfakes"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var _ = Describe("Dump", func() {
	Context("RoundRobinDump", func() {
		var (
			fake1  *dumperfakes.FakeDumper
			fake2  *dumperfakes.FakeDumper
			logger *zap.Logger
			client dumper.Dumper
			r      io.Reader
			err    error
		)
		BeforeEach(func() {
			fake1 = &dumperfakes.FakeDumper{}
			fake2 = &dumperfakes.FakeDumper{}
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewRoundRobinDumpClient(logger, fake1, fake2)
			err = client.Dump(context.Background(), r, "some path")
		})
		When("one of the clients error", func() {
			BeforeEach(func() {
				fake1.DumpReturns(errors.New("an error"))
			})
			It("does err", func() {
				Expect(err).To(MatchError("an error"))
			})
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("calls both clients", func() {
				Expect(fake1.DumpCallCount()).To(Equal(1))
				Expect(fake2.DumpCallCount()).To(Equal(1))
			})
		})
	})
	Context("S3DumpHandler", func() {
		var (
			uploader *dumperfakes.FakeUploader
			logger   *zap.Logger
			client   dumper.Dumper
			r        io.Reader
			err      error
		)
		BeforeEach(func() {
			uploader = &dumperfakes.FakeUploader{}
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewS3DumpHandler(uploader, "bucket", logger)
			err = client.Dump(context.Background(), r, "some path")
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("gives the correct upload input", func() {
				inp, _ := uploader.UploadArgsForCall(0)
				Expect(inp).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"Bucket": PointTo(Equal("bucket")),
					"Key":    PointTo(Equal("some path")),
				})))
			})
		})
	})
	Context("LocalDumpHandler", func() {
		var (
			fs     afero.Fs
			logger *zap.Logger
			client dumper.Dumper
			r      io.Reader
			err    error
		)
		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			logger = zap.NewNop()
			r = strings.NewReader("ahhhhh")
			err = nil
		})
		JustBeforeEach(func() {
			client = dumper.NewLocalDumpHandler("path", logger, fs)
			err = client.Dump(context.Background(), r, "somepath")
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("writes to the filesystem", func() {
				_, err := fs.Stat("path/somepath")
				Expect(err).To(BeNil())
			})
		})
		When("it fails to copy", func() {
			BeforeEach(func() {
				var w *io.PipeWriter
				r, w = io.Pipe()
				w.CloseWithError(errors.New("broken pipe")) //nolint:errcheck
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("DynamoDumpHandler", func() {
		var (
			table  string
			logger *zap.Logger
			dp     *dumperfakes.FakeDynamoPuter
			dh     dumper.DynamoDumpHandler
			err    error
		)
		BeforeEach(func() {
			table = "table"
			logger = zap.NewNop()
			dp = &dumperfakes.FakeDynamoPuter{}
			err = nil
		})
		JustBeforeEach(func() {
			dh = dumper.NewDynamoDumpHandler(
				logger,
				table,
				dp,
				dumper.NoOpMarshaller,
			)
			err = dh.Dump(context.Background(), strings.NewReader(""), "somepath")
		})
		When("it dumps", func() {
			It("does not err", func() {
				Expect(err).To(BeNil())
			})
			It("calls batch write item", func() {
				Expect(dp.BatchWriteItemWithContextCallCount()).To(Equal(1))
			})
		})
	})
	Context("PostgresDumpHandler", func() {
		var (
			logger   *zap.Logger
			dh       dumper.PostgresDumpHandler
			upserter *postgresfakes.FakeUpserter
			err      error
			r        io.Reader
		)
		BeforeEach(func() {
			logger = zap.NewNop()
			upserter = &postgresfakes.FakeUpserter{}
			r = strings.NewReader(`[
				{
					"DIRECTION": "N",
					"LINE": "BLUE",
					"TRAIN_ID": "206401"
				},
				{
					"DIRECTION": "N",
					"LINE": "BLUE",
					"TRAIN_ID": "206402"
				},
				{
					"DIRECTION": "N",
					"LINE": "BLUE",
					"TRAIN_ID": "206402"
				},
				{
					"DIRECTION": "N",
					"LINE": "BLUE",
					"TRAIN_ID": "206401"
				}
			]`)
			err = nil
		})
		JustBeforeEach(func() {
			dh = dumper.NewPostgresDumpHandler(
				logger,
				upserter,
				nil,
			)
			err = dh.Dump(context.Background(), r, "somepath")
		})
		When("the JSON is invalid", func() {
			BeforeEach(func() {
				r = strings.NewReader("{")
			})
			It("fails", func() {
				Expect(err).To(MatchError("unexpected EOF"))
				Expect(upserter.AddRecordToDatabaseCallCount()).To(Equal(0))
			})
		})
		When("an upsert fails", func() {
			BeforeEach(func() {
				upserter.AddRecordToDatabaseReturnsOnCall(0, errors.New("upsert failed"))
			})
			It("logs and moves on", func() {
				Expect(err).To(BeNil())
				Expect(upserter.AddRecordToDatabaseCallCount()).To(Equal(4))
			})
		})
		When("all goes well", func() {
			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(upserter.AddRecordToDatabaseCallCount()).To(Equal(4))

				_, l, d, _, _, _ := upserter.AddRecordToDatabaseArgsForCall(0)
				Expect(l).To(Equal(martaapi.Blue))
				Expect(d).To(Equal(martaapi.North))
			})
		})
	})
})
