package config

//AppConfig struct
type AppConfig struct {
	Subfolders []string `json:"subfolders"`
	Redis      struct {
		Host      string `json:"host"`
		Port      string `json:"port"`
		Namespace string `json:"namespace"`
		DebugPort string `json:"httpPort"`
	} `json:"redis"`
	Database struct {
		File string `json:"file"`
	} `json:"database"`
	Web struct {
		Port string `json:"port"`
	} `json:"web"`
	Downloads struct {
		PathPrefix string `json:"pathPrefix"`
	} `json:"downloads"`
}
