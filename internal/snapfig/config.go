package snapfig

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Locations         []Location        `mapstructure:"locations"`
	WarningThresholds WarningThresholds `mapstructure:"warning_thresholds"`
}

type Location struct {
	Path      string   `mapstructure:"path"`
	Name      string   `mapstructure:"name"`
	Frequency string   `mapstructure:"frequency"`
	Branch    string   `mapstructure:"branch"`
	Include   []string `mapstructure:"include,omitempty"`
}

type WarningThresholds struct {
	CommitSizeMB int `mapstructure:"commit_size_mb"`
	FileCount    int `mapstructure:"file_count"`
}

func LoadConfig(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
		return nil, err
	}

	return &config, nil
}
