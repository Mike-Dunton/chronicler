package adding

// Service provides DownloadRecord adding operations.
type Service interface {
	AddDownloadRecord(...DownloadRecord) (*[]int64, error)
	ValidateDownloadRecord(...DownloadRecord) bool
}

// Repository provides access to DownloadRecord repository.
type Repository interface {
	// AddDownloadRecord saves a given Download Record to the repository.
	AddDownloadRecord(*DownloadRecord) (int64, error)
}

// Queue provides access to the Worker Queue
type Queue interface {
	// EnqueueDownload queues the download for downloading
	EnqueueDownload(*DownloadRecord, int64) error
}

// Validator provides access to the Download Request Validator.
type Validator interface {
	ValidateDownloadRequest(*DownloadRecord) bool
}

type service struct {
	r Repository
	q Queue
	v Validator
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository, q Queue, v Validator) Service {
	return &service{r, q, v}
}

// AddDownloadRecord persists the given DownloadRecord(s) to storage
func (s *service) AddDownloadRecord(dr ...DownloadRecord) (*[]int64, error) {
	var downloadRecords []int64
	for _, downloadRecord := range dr {
		downloadRecordID, err := s.r.AddDownloadRecord(&downloadRecord)
		if err != nil {
			return &downloadRecords, err
		}
		err = s.q.EnqueueDownload(&downloadRecord, downloadRecordID)
		if err != nil {
			return &downloadRecords, err
		}
		downloadRecords = append(downloadRecords, downloadRecordID)
	}

	return &downloadRecords, nil
}

func (s *service) ValidateDownloadRecord(dr ...DownloadRecord) bool {
	for _, downloadRecord := range dr {
		isValid := s.v.ValidateDownloadRequest(&downloadRecord)
		if !isValid {
			return false
		}
	}

	return true
}
