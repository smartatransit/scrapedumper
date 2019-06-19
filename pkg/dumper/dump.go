package dumper

import (
	"fmt"
	"context"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Dumper
type Dumper interface {
	Dump(ctx context.Context, r io.Reader, path string) error
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

	fmt.Println(dir)
	err := os.MkdirAll(dir, 644)
	if err != nil {
		fmt.Println(1)
		return err
	}

		fmt.Println(2)
	f, err := os.Open(location)
	if err != nil {
		fmt.Println(3)
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		fmt.Println(4)
		return err
	}

		fmt.Println(5)
	return f.Close()
}
