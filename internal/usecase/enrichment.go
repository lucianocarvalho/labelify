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
	"github.com/lucianocarvalho/labelify/internal/infrastructure/sources"
)

type EnrichmentUseCase struct {
	config  *domain.Config
	sources map[string]domain.SourceProvider
}

func NewEnrichmentUseCase(config *domain.Config) (*EnrichmentUseCase, error) {
	sourcesMap := make(map[string]domain.SourceProvider)

	for _, sourceConfig := range config.Sources {
		source, err := sources.NewSource(sourceConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating source %s: %w", sourceConfig.Name, err)
		}
		sourcesMap[sourceConfig.Name] = source
	}

	return &EnrichmentUseCase{
		config:  config,
		sources: sourcesMap,
	}, nil
}

func (h *EnrichmentUseCase) hasApplicableRules(query string, resp domain.QueryResponse) bool {
	for _, rule := range h.config.Enrichment.Rules {
		if !strings.Contains(query, rule.Match.Metric) {
			continue
		}

		for _, r := range resp.Data.Result {
			if _, ok := r.Metric[rule.Match.Label]; ok {
				return true
			}
		}
	}
	return false
}

// Nova função auxiliar para coletar todas as labels
func (h *EnrichmentUseCase) getAllLabels(query string) []string {
	labelSet := make(map[string]bool)
	for _, rule := range h.config.Enrichment.Rules {
		if strings.Contains(query, rule.Match.Metric) {
			for _, label := range rule.AddLabels {
				labelSet[label] = true
			}
		}
	}

	var labels []string
	for label := range labelSet {
		labels = append(labels, label)
	}
	return labels
}

func (h *EnrichmentUseCase) Execute(body []byte, originalQuery string) ([]byte, error) {
	var resp domain.QueryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// Probably not a query response, just returning
		return body, nil
	}

	if !h.hasApplicableRules(originalQuery, resp) {
		// Nothing to do, no applicable rules found, just returning
		return body, nil
	}

	log.Printf("Found applicable rules for query for query '%s': ", originalQuery)

	for _, rule := range h.config.Enrichment.Rules {
		log.Printf("Evaluating rule for metric: %s", rule.Match.Metric)

		source, ok := h.sources[rule.EnrichFrom]
		if !ok {
			log.Printf("Source %s not found for rule", rule.EnrichFrom)
			continue
		}

		mappings, err := source.GetMappings()
		if err != nil {
			log.Printf("Error getting mappings from source %s: %v", rule.EnrichFrom, err)
			continue
		}

		for i, r := range resp.Data.Result {
			if !h.matchesMetric(r.Metric, rule.Match, originalQuery) {
				continue
			}

			labelValue := r.Metric[rule.Match.Label]
			if labelValue == "" {
				continue
			}

			var matchedData *domain.SourceData
			for pattern, data := range mappings {
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

	allLabels := h.getAllLabels(originalQuery)

	switch resp.Data.ResultType {
	case "matrix":
		groupedMetrics := make(map[string][][]interface{})
		for _, r := range resp.Data.Result {
			groupKey := make(map[string]string)
			for _, label := range allLabels { // Usar allLabels ao invés de h.config.Enrichment.Rules[0].AddLabels
				if value, ok := r.Metric[label]; ok {
					groupKey[label] = value
				}
			}

			if len(groupKey) == 0 {
				continue
			}

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
			groupKey := make(map[string]string)
			for _, label := range allLabels {
				if value, ok := r.Metric[label]; ok {
					groupKey[label] = value
				}
			}

			if len(groupKey) == 0 {
				continue
			}

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

func (h *EnrichmentUseCase) matchesMetric(metric map[string]string, match domain.MatchRule, query string) bool {
	return strings.Contains(query, match.Metric)
}

func (h *EnrichmentUseCase) createGroupKey(groupKey map[string]string) string {
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
