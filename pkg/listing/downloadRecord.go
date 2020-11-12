package listing

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Output    string `json:"output"`
	Errors    string `json:"errors"`
	Finished  string `json:"finished"`
}
