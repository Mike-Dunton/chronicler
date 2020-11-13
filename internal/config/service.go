package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Service provides DownloadRecord adding operations.
type Service interface {
	LoadConfig() (*AppConfig, error)
}

type service struct {
	configFile   string
	loadedConfig *AppConfig
}

// NewService creates an adding service with the necessary dependencies
func NewService(cf string) Service {
	return &service{cf, new(AppConfig)}
}

func (s *service) LoadConfig() (*AppConfig, error) {
	jsonFile, err := os.Open(s.configFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(byteValue, s.loadedConfig)
	if err != nil {
		return nil, err
	}
	return s.loadedConfig, nil
}
