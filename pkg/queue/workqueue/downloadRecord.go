package workqueue

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
}
