package file

// Download defines the stored Data on disk
type Download struct {
	DownloadName string                 `json:"basePath"`
	Subfolder    string                 `json:"subfolder"`
	Description  string                 `json:"description"`
	Info         map[string]interface{} `json:"info"`
}
