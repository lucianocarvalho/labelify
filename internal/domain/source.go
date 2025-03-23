package domain

type SourceProvider interface {
	GetMappings() (map[string]SourceData, error)
	Name() string
}

type SourceType string

const (
	SourceTypeYAML SourceType = "yaml"
	SourceTypeHTTP SourceType = "http"
)
