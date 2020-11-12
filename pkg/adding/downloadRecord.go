package adding

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
}
