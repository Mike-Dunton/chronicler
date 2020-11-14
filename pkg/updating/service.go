package updating

import "fmt"

// Service provides DownloadRecord updating operations.
type Service interface {
	UpdateDownloadRecord(...DownloadRecord) []error
}

// Repository provides access to DownloadRecord repository.
type Repository interface {
	// AddDownloadRecord saves a given Download Record to the repository.
	UpdateDownloadRecord(*DownloadRecord) error
}

type service struct {
	r Repository
}

// NewService creates an updating service with the necessary dependencies
func NewService(r Repository) Service {
	return &service{r}
}

// AddDownloadRecord persists the given DownloadRecord(s) to storage
func (s *service) UpdateDownloadRecord(dr ...DownloadRecord) (errorList []error) {
	fmt.Printf("%v many records to update\n", len(dr))
	for _, downloadRecord := range dr {
		err := s.r.UpdateDownloadRecord(&downloadRecord)
		if err != nil {
			errorList = append(errorList, err)
		}
	}
	return errorList
}
