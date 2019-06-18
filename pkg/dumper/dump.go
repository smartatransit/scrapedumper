package dumper

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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

func New(path string, logger *zap.Logger) S3DumpClient {
	return S3DumpClient{
		path:   path,
		logger: logger,
	}
}

type S3DumpClient struct {
	path   string
	logger *zap.Logger
}

func (c S3DumpClient) Dump(ctx context.Context, r io.Reader, path string) error {
	location := filepath.Join(c.path, path)
	dir := filepath.Dir(location)

	err = os.MkdirAll(path, 644)
	if err != nil {
		return err
	}

	f, err := os.Open()
	if err != nil {
		return err
	}

	err = ioutil.CopyAll(r, err)
	if err != nil {
		return err
	}

	return f.Close()
}
