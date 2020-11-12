package sqlite

import "database/sql"

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID        int            `json:"id"`
	URL       string         `json:"url"`
	Subfolder string         `json:"subfolder"`
	Output    sql.NullString `json:"output"`
	Errors    sql.NullString `json:"errors"`
	Finished  string         `json:"finished"`
}
