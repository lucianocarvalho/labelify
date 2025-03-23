package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/lucianocarvalho/labelify/internal/domain"
)

type HTTPSource struct {
	name     string
	config   domain.SourceConfig
	client   *http.Client
	mappings map[string]domain.SourceData
	mu       sync.RWMutex
}

func NewHTTPSource(name string, config domain.SourceConfig) *HTTPSource {
	source := &HTTPSource{
		name:   name,
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Fazer a primeira carga
	if err := source.refresh(); err != nil {
		// Log do erro, mas continua com mappings vazio
		fmt.Printf("Error loading initial mappings for source %s: %v\n", name, err)
	}

	// Se tiver refresh interval configurado, iniciar o refresh periódico
	if config.RefreshInterval != "" {
		duration, err := time.ParseDuration(config.RefreshInterval)
		if err != nil {
			fmt.Printf("Invalid refresh interval for source %s: %v\n", name, err)
			return source
		}

		go source.startRefreshLoop(duration)
	}

	return source
}

func (s *HTTPSource) GetMappings() (map[string]domain.SourceData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mappings, nil
}

func (s *HTTPSource) Name() string {
	return s.name
}

func (s *HTTPSource) refresh() error {
	req, err := http.NewRequest(s.config.Method, s.config.URL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var newMappings map[string]domain.SourceData
	if err := json.Unmarshal(body, &newMappings); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	s.mu.Lock()
	s.mappings = newMappings
	s.mu.Unlock()

	return nil
}

func (s *HTTPSource) startRefreshLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := s.refresh(); err != nil {
			fmt.Printf("Error refreshing mappings for source %s: %v\n", s.name, err)
		}
	}
}
