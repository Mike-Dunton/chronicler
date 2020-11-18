package file

import (
	"github.com/go-kit/kit/log"
	//sql driver
	_ "github.com/mattn/go-sqlite3"
)

// Storage is the interface that defines interacting with Download Records
type Storage struct {
	basePath string
	logger   log.Logger
}

// NewStorage returns a new Sql DB  storage
func NewStorage(logger *log.Logger, basePath string) (*Storage, error) {
	s := new(Storage)
	s.logger = log.With(*logger, "pkg", "file")
	s.basePath = basePath
	return s, nil
}

//GetDownloadRecord gets records.
func (s *Storage) GetDownloadInfo(downloadPath string) (*Download, error) {
	var download Download
	return &download, nil
}
