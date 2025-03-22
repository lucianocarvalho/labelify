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

	// Verifica se alguma regra dá match com a query
	hasMatchingRule := false
	for _, rule := range h.config.Enrichment.Rules {
		if h.matchesMetric(resp.Data.Result[0].Metric, rule.Match, originalQuery) {
			hasMatchingRule = true
			break
		}
	}

	// Se não houver regra que dê match, retorna a resposta original
	if !hasMatchingRule {
		log.Printf("No matching rules found for query: '%s', returning original response", originalQuery)
		return body, nil
	}

	for _, rule := range h.config.Enrichment.Rules {
		log.Printf("Evaluating rule for metric: %s", rule.Match.Metric)

		// Find the source for this rule
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
			// Check if this metric matches our rule
			if !h.matchesMetric(r.Metric, rule.Match, originalQuery) {
				continue
			}

			// Get the label value to match against
			labelValue := r.Metric[rule.Match.Label]
			if labelValue == "" {
				continue
			}

			// Try to find a matching mapping
			var matchedData *domain.SourceData
			for pattern, data := range source.Mappings {
				if pattern == labelValue {
					matchedData = &data
					break
				}
				// Try wildcard match
				if matched, _ := regexp.MatchString(pattern, labelValue); matched {
					matchedData = &data
					break
				}
			}

			// Apply the matched data or fallback
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

			// Skip if we don't have any labels to group by
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
				Value:  value,
			})
		}
		resp.Data.Result = newResult
	}

	return json.Marshal(resp)
}

func (h *HydrateUseCase) matchesMetric(metric map[string]string, match domain.MatchRule, query string) bool {
	// Verifica se a métrica está presente na query original
	return strings.Contains(query, match.Metric)
}

func (h *HydrateUseCase) createGroupKey(groupKey map[string]string) string {
	// Ordena as chaves para garantir consistência
	keys := make([]string, 0, len(groupKey))
	for k := range groupKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Cria a chave ordenada
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, groupKey[k]))
	}
	return strings.Join(parts, ",")
}
