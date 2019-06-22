package dumper

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Dumper
type Dumper interface {
	Dump(ctx context.Context, r io.Reader, path string) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Uploader
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type RoundRobinDumpClient struct {
	logger  *zap.Logger
	clients []Dumper
}

func NewRoundRobinDumpClient(logger *zap.Logger, clients ...Dumper) RoundRobinDumpClient {
	return RoundRobinDumpClient{
		logger,
		clients,
	}
}

func (c RoundRobinDumpClient) Dump(ctx context.Context, r io.Reader, path string) error {
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

type LocalDumpClient struct {
	path   string
	logger *zap.Logger
	fs     afero.Fs
}

func NewLocalDumpClient(path string, logger *zap.Logger, fs afero.Fs) LocalDumpClient {
	return LocalDumpClient{
		path,
		logger,
		fs,
	}
}

func (c LocalDumpClient) Dump(ctx context.Context, r io.Reader, path string) error {
	location := filepath.Join(c.path, path)

	f, err := c.fs.Create(location)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return f.Close()
}

func NewS3DumpClient(uploader Uploader, bucket string, logger *zap.Logger) S3DumpClient {
	return S3DumpClient{
		uploader,
		bucket,
		logger,
	}
}

type S3DumpClient struct {
	uploader Uploader
	bucket   string
	logger   *zap.Logger
}

func (c S3DumpClient) Dump(ctx context.Context, r io.Reader, path string) error {
	_, err := c.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
		Body:   r,
	})
	return err
}
