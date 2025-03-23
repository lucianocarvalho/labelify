package sources

import (
	"fmt"

	"github.com/lucianocarvalho/labelify/internal/domain"
)

func NewSource(source *domain.Source) (domain.SourceProvider, error) {
	switch domain.SourceType(source.Type) {
	case domain.SourceTypeYAML:
		return NewYAMLSource(source.Name, source.Mappings), nil
	case domain.SourceTypeHTTP:
		return NewHTTPSource(source.Name, source.Config), nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", source.Type)
	}
}
