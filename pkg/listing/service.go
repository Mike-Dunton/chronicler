package listing

// Service provides downloadRecords listing operations.
type Service interface {
	GetDownloadRecord(int64) (*DownloadRecord, error)
	GetDownloadRecords() (*[]DownloadRecord, error)
}

// Repository provides access to DownloadRecord repository.
type Repository interface {
	GetDownloadRecord(int64) (*DownloadRecord, error)
	AllDownloadRecords() (*[]DownloadRecord, error)
}

type service struct {
	r Repository
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository) Service {
	return &service{r}
}

// GetDownloadRecords returns all DownloadRecords
func (s *service) GetDownloadRecords() (*[]DownloadRecord, error) {
	return s.r.AllDownloadRecords()
}

// GetDownloadRecord returns a DownloadRecord
func (s *service) GetDownloadRecord(id int64) (*DownloadRecord, error) {
	return s.r.GetDownloadRecord(id)
}
