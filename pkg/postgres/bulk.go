package postgres

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/bipol/scrapedumper/pkg/martaapi"
)

//BulkLoader loads a file into the Postgres database
type BulkLoader interface {
	Load(io.Reader) error
	LoadDir(dir string) error
}

//BulkLoaderAgent implements BulkLoader
type BulkLoaderAgent struct {
	Upserter Upserter
}

//LoadDir loads all files in the specified directory
func (a *BulkLoaderAgent) LoadDir(dir string) error {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		file, err := os.Open(info.Name())
		if err != nil {
			return err
		}

		err = a.Load(file)
		if err != nil {
			return err
		}
	}

	return nil
}

//Load loads a file into the Postgres database
func (a *BulkLoaderAgent) Load(r io.Reader) error {
	var records []martaapi.Schedule
	err := json.NewDecoder(r).Decode(records)
	if err != nil {
		return err
	}

	for _, rec := range records {
		err := a.Upserter.AddRecordToDatabase(rec)
		if err != nil {
			return err
		}
	}

	return nil
}
