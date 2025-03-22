package config

import (
	"log"
	"os"

	"github.com/lucianocarvalho/labelify/internal/domain"
	"gopkg.in/yaml.v3"
)

func LoadLabelifyConfig(path string) (*domain.Config, error) {
	log.Printf("Loading Labelify config from file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return nil, err
	}
	defer f.Close()

	var config domain.Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Error decoding config: %v", err)
		return nil, err
	}

	log.Printf("Config loaded successfully. Total sources: %d, Total rules: %d",
		len(config.Sources), len(config.Enrichment.Rules))

	return &config, nil
}
