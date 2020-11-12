package validating

import (
	"github.com/mike-dunton/chronicler/pkg/adding"
)

// Validator is the interface that defines validation functions
type Validator struct {
	subfolders []string
}

// NewValidator returns a Validator
func NewValidator(subfolders []string) (*Validator, error) {
	v := new(Validator)

	v.subfolders = subfolders
	return v, nil
}

// ValidateDownloadRequest validates the download request
func (v *Validator) ValidateDownloadRequest(dr *adding.DownloadRecord) bool {
	if Contains(v.subfolders, dr.Subfolder) {
		return true
	}
	return false
}

// Contains takes a slice and looks for an element in it.
func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
