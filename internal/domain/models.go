package domain

type RuleSet struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

type Rule struct {
	Name   string     `json:"name" yaml:"name"`
	Mutate MutateRule `json:"mutate" yaml:"mutate"`
}

type MutateRule struct {
	Type         string    `json:"type" yaml:"type"`
	Matchers     []Matcher `json:"matchers" yaml:"matchers"`
	TargetLabel  string    `json:"targetLabel" yaml:"target_label"`
	DefaultValue string    `json:"defaultValue" yaml:"default_value"`
}

type Matcher struct {
	MatchLabels map[string]string `json:"match_labels" yaml:"match_labels"`
	Value       string            `json:"value" yaml:"value"`
}

type PrometheusResponse struct {
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
