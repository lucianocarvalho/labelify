package domain

type Config struct {
	Config     ServerConfig `json:"config" yaml:"config"`
	Sources    []Source     `json:"sources" yaml:"sources"`
	Enrichment Enrichment   `json:"enrichment" yaml:"enrichment"`
}

type ServerConfig struct {
	Prometheus PrometheusConfig `json:"prometheus" yaml:"prometheus"`
	Server     PortConfig       `json:"server" yaml:"server"`
}

type PrometheusConfig struct {
	URL string `json:"url" yaml:"url"`
}

type PortConfig struct {
	Port int `json:"port" yaml:"port"`
}

type Source struct {
	Name     string                `json:"name" yaml:"name"`
	Type     string                `json:"type" yaml:"type"`
	Mappings map[string]SourceData `json:"mappings" yaml:"mappings"`
}

type SourceData struct {
	Labels map[string]string `json:"labels" yaml:"labels"`
}

type Enrichment struct {
	Rules []EnrichmentRule `json:"rules" yaml:"rules"`
}

type EnrichmentRule struct {
	Match      MatchRule         `json:"match" yaml:"match"`
	EnrichFrom string            `json:"enrich_from" yaml:"enrich_from"`
	AddLabels  []string          `json:"add_labels" yaml:"add_labels"`
	Fallback   map[string]string `json:"fallback" yaml:"fallback"`
}

type MatchRule struct {
	Metric string `json:"metric" yaml:"metric"`
	Label  string `json:"label" yaml:"label"`
}

type QueryResponse struct {
	Status string `json:"status" yaml:"status"`
	Data   struct {
		ResultType string       `json:"resultType" yaml:"result_type"`
		Result     []MetricData `json:"result" yaml:"result"`
	} `json:"data" yaml:"data"`
}

type MetricData struct {
	Metric map[string]string `json:"metric" yaml:"metric"`
	Values [][]interface{}   `json:"values,omitempty" yaml:"values,omitempty"`
	Value  []interface{}     `json:"value,omitempty" yaml:"value,omitempty"`
}
