package updating

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID       int64  `json:"id"`
	Output   string `json:"output"`
	Errors   string `json:"errors"`
	Finished string `json:"finished"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
}
