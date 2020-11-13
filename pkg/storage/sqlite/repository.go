package sqlite

import (
	"database/sql"

	//sql driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/updating"
)

// Storage is the interface that defines interacting with Download Records
type Storage struct {
	db *sql.DB
}

// NewStorage returns a new Sql DB  storage
func NewStorage(sqlDir string) (*Storage, error) {
	var err error

	s := new(Storage)

	s.db, err = sql.Open("sqlite3", sqlDir)
	if err != nil {
		return nil, err
	}
	prepDatabase(s.db)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func prepDatabase(db *sql.DB) error {
	downloads := `
	CREATE TABLE IF NOT EXISTS downloads(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT, 
		subfolder TEXT, 
		output TEXT, 
		error TEXT, 
		finished TEXT NOT NULL
		CHECK( typeof("finished") = "text" AND
			"finished" IN ("true","false")
		)
	);
	`
	_, err := db.Exec(downloads)
	if err != nil {
		return err
	}
	return nil
}

//AllDownloadRecords gets records.
func (s *Storage) AllDownloadRecords() (*[]listing.DownloadRecord, error) {
	sql := "SELECT id, url, subfolder, output, error, finished FROM downloads"
	rows, err := s.db.Query(sql)
	// Exit if the SQL doesn't work for some reason
	if err != nil {
		return nil, err
	}
	// make sure to cleanup when the program exits
	defer rows.Close()

	var resultDownloads []listing.DownloadRecord
	for rows.Next() {
		var download DownloadRecord
		err = rows.Scan(&download.ID, &download.URL, &download.Subfolder, &download.Output, &download.Errors, &download.Finished)
		// Exit if we get an error
		if err != nil {
			return nil, err
		}
		var resultDownload listing.DownloadRecord
		resultDownload.ID = download.ID
		resultDownload.Errors = download.Errors.String
		resultDownload.Finished = download.Finished
		resultDownload.Output = download.Output.String
		resultDownload.Subfolder = download.Subfolder
		resultDownload.URL = download.URL

		resultDownloads = append(resultDownloads, resultDownload)
	}
	return &resultDownloads, nil
}

//GetDownloadRecord gets records.
func (s *Storage) GetDownloadRecord(id int64) (*listing.DownloadRecord, error) {
	sql := "SELECT id, url, subfolder, output, error, finished FROM downloads WHERE id = ?"
	statement, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	row := statement.QueryRow(id)

	var download DownloadRecord
	err = row.Scan(&download.ID, &download.URL, &download.Subfolder, &download.Output, &download.Errors, &download.Finished)
	// Exit if we get an error
	if err != nil {
		return nil, err
	}
	var resultDownload listing.DownloadRecord
	resultDownload.ID = download.ID
	resultDownload.Errors = download.Errors.String
	resultDownload.Finished = download.Finished
	resultDownload.Output = download.Output.String
	resultDownload.Subfolder = download.Subfolder
	resultDownload.URL = download.URL

	return &resultDownload, nil
}

//AddDownloadRecord Puts the records
func (s *Storage) AddDownloadRecord(dr *adding.DownloadRecord) (int64, error) {
	sql := "INSERT INTO downloads(url, subfolder, finished) VALUES(?,?,?)"
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, sqlExecError := stmt.Exec(dr.URL, dr.Subfolder, "false")
	if sqlExecError != nil {
		return 0, sqlExecError
	}
	return result.LastInsertId()
}

//UpdateDownloadRecord Puts the records
func (s *Storage) UpdateDownloadRecord(dr *updating.DownloadRecord) error {
	sql := `
	UPDATE downloads
	SET output = $2, error = $3, finished = $4
	WHERE id = $1;
	`
	_, err := s.db.Exec(sql, dr.ID, dr.Output, dr.Errors, dr.Finished)
	if err != nil {
		return err
	}
	return nil
}
