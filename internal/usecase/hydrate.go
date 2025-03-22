package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lucianocarvalho/labelify/internal/domain"
)

type HydrateUseCase struct {
	config *domain.Config
}

func NewHydrateUseCase(config *domain.Config) *HydrateUseCase {
	return &HydrateUseCase{
		config: config,
	}
}

func (h *HydrateUseCase) Execute(body []byte, originalQuery string) ([]byte, error) {
	log.Printf("Executing hydration for query: '%s'", originalQuery)

	var resp domain.PrometheusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Printf("Error unmarshaling: %v", err)
		return nil, fmt.Errorf("response incompatible with PrometheusResponse")
	}

	log.Printf("Response status: %s", resp.Status)
	log.Printf("Result type: %s", resp.Data.ResultType)
	log.Printf("Number of results: %d", len(resp.Data.Result))

	for _, rule := range h.config.Enrichment.Rules {
		log.Printf("Evaluating rule for metric: %s", rule.Match.Metric)

		var source *domain.Source
		for _, s := range h.config.Sources {
			if s.Name == rule.EnrichFrom {
				source = &s
				break
			}
		}
		if source == nil {
			log.Printf("Source %s not found for rule", rule.EnrichFrom)
			continue
		}

		for i, r := range resp.Data.Result {
			if !h.matchesMetric(r.Metric, rule.Match, originalQuery) {
				continue
			}

			// Get the label value to match against
			labelValue := r.Metric[rule.Match.Label]
			if labelValue == "" {
				continue
			}

			var matchedData *domain.SourceData
			for pattern, data := range source.Mappings {
				if pattern == labelValue {
					matchedData = &data
					break
				}
				if matched, _ := regexp.MatchString(pattern, labelValue); matched {
					matchedData = &data
					break
				}
			}

			if matchedData != nil {
				for _, label := range rule.AddLabels {
					if value, ok := matchedData.Labels[label]; ok {
						resp.Data.Result[i].Metric[label] = value
					}
				}
			} else {
				for label, value := range rule.Fallback {
					resp.Data.Result[i].Metric[label] = value
				}
			}
		}
	}

	switch resp.Data.ResultType {
	case "matrix":
		groupedMetrics := make(map[string][][]interface{})
		for _, r := range resp.Data.Result {
			// Get all labels that we want to group by
			groupKey := make(map[string]string)
			for _, label := range h.config.Enrichment.Rules[0].AddLabels {
				if value, ok := r.Metric[label]; ok {
					groupKey[label] = value
				}
			}

			// Skip if we don't have any labels to group by
			if len(groupKey) == 0 {
				continue
			}

			// Create a unique key for this group
			groupKeyStr := h.createGroupKey(groupKey)

			if values, exists := groupedMetrics[groupKeyStr]; exists {
				for i, v := range r.Values {
					if i >= len(values) {
						values = append(values, v)
					} else {
						val1, _ := strconv.Atoi(v[1].(string))
						val2, _ := strconv.Atoi(values[i][1].(string))
						values[i][1] = strconv.Itoa(val1 + val2)
					}
				}
			} else {
				groupedMetrics[groupKeyStr] = r.Values
			}
		}

		var newResult []domain.MetricData
		for groupKey, values := range groupedMetrics {
			// Parse the group key back into a map
			metric := make(map[string]string)
			for _, pair := range strings.Split(groupKey, ",") {
				if pair == "" {
					continue
				}
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					metric[parts[0]] = parts[1]
				}
			}

			newResult = append(newResult, domain.MetricData{
				Metric: metric,
				Values: values,
			})
		}
		resp.Data.Result = newResult

	case "vector":
		groupedMetrics := make(map[string][]interface{})
		for _, r := range resp.Data.Result {
			// Get all labels that we want to group by
			groupKey := make(map[string]string)
			for _, label := range h.config.Enrichment.Rules[0].AddLabels {
				if value, ok := r.Metric[label]; ok {
					groupKey[label] = value
				}
			}

			if len(groupKey) == 0 {
				continue
			}

			// Create a unique key for this group
			groupKeyStr := h.createGroupKey(groupKey)

			if value, exists := groupedMetrics[groupKeyStr]; exists {
				val1, _ := strconv.Atoi(r.Value[1].(string))
				val2, _ := strconv.Atoi(value[1].(string))
				value[1] = strconv.Itoa(val1 + val2)
			} else {
				groupedMetrics[groupKeyStr] = r.Value
			}
		}

		var newResult []domain.MetricData
		for groupKey, value := range groupedMetrics {
			metric := make(map[string]string)
			for _, pair := range strings.Split(groupKey, ",") {
				if pair == "" {
					continue
				}
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					metric[parts[0]] = parts[1]
				}
			}

			newResult = append(newResult, domain.MetricData{
				Metric: metric,
				Value:  value,
			})
		}
		resp.Data.Result = newResult
	}

	return json.Marshal(resp)
}

func (h *HydrateUseCase) matchesMetric(metric map[string]string, match domain.MatchRule, query string) bool {
	return strings.Contains(query, match.Metric)
}

func (h *HydrateUseCase) createGroupKey(groupKey map[string]string) string {
	// Ordena as chaves para garantir consistência
	keys := make([]string, 0, len(groupKey))
	for k := range groupKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, groupKey[k]))
	}
	return strings.Join(parts, ",")
}
