package utils 

import "io"

type Response struct {
	StatusCode:	int			`json:"statuscode"`
	Success  bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ServiceConfig struct {
	Services struct {
		Auth struct {
			Port      int    `yaml:"port"`
			BaseURL   string `yaml:"base_url"`
		} `yaml:"auth"`
		Sentinel struct {
			Port      int    `yaml:"port"`
			BaseURL   string `yaml:"base_url"`
		} `yaml:"sentinel"`
	} `yaml:"services"`
}

type HTTPRequestInfo struct {
	Header      map[string][]string
	IPAddress   string
	Scheme		string
	Host        string
	UserAgent   string
	Method      string
	OriginalURL string
	Body        io.Reader
}


type Request struct {
	Header	map[string][]string
	Method string
	URL string
	Body io.Reader
}