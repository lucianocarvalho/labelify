package config

import (
	"log"
	"os"

	"github.com/lucianocarvalho/labelify/internal/domain"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PrometheusURL string          `yaml:"prometheus_url"`
	Port          string          `yaml:"port"`
	RulesPath     string          `yaml:"rules_path"`
	Rules         *domain.RuleSet `yaml:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadRules(path string) (*domain.RuleSet, error) {
	log.Printf("Loading rules from file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening rules file: %v", err)
		return nil, err
	}
	defer f.Close()

	var rs domain.RuleSet
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&rs)
	if err != nil {
		log.Printf("Error decoding rules: %v", err)
		return nil, err
	}

	log.Printf("Rules loaded successfully. Total rules: %d", len(rs.Rules))
	for i, rule := range rs.Rules {
		log.Printf("Rule %d: Name=%s, Type=%s, Label=%s",
			i+1, rule.Name, rule.Mutate.Type, rule.Mutate.TargetLabel)
	}

	return &rs, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
