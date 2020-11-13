package updating

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID       int    `json:"id"`
	Output   string `json:"output"`
	Errors   string `json:"errors"`
	Finished string `json:"finished"`
}
