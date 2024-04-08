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