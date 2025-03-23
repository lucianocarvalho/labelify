package sources

import "github.com/lucianocarvalho/labelify/internal/domain"

type YAMLSource struct {
	name     string
	mappings map[string]domain.SourceData
}

func NewYAMLSource(name string, mappings map[string]domain.SourceData) *YAMLSource {
	return &YAMLSource{
		name:     name,
		mappings: mappings,
	}
}

func (s *YAMLSource) GetMappings() (map[string]domain.SourceData, error) {
	return s.mappings, nil
}

func (s *YAMLSource) Name() string {
	return s.name
}
