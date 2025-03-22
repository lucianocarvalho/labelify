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

	for i, source := range config.Sources {
		log.Printf("Source %d: Name=%s, Type=%s, Mappings=%d",
			i+1, source.Name, source.Type, len(source.Mappings))
	}

	for i, rule := range config.Enrichment.Rules {
		log.Printf("Rule %d: Metric=%s, Label=%s, EnrichFrom=%s",
			i+1, rule.Match.Metric, rule.Match.Label, rule.EnrichFrom)
	}

	return &config, nil
}
