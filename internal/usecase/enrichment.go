package usecase

import (
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
		source := sourceConfig
		provider, err := sources.NewSource(&source)
		if err != nil {
			return nil, fmt.Errorf("error creating source %s: %w", source.Name, err)
		}
		sourcesMap[source.Name] = provider
	}

	return &EnrichmentUseCase{
		config:  config,
		sources: sourcesMap,
	}, nil
}

func (h *EnrichmentUseCase) Execute(resp *domain.QueryResponse, originalQuery string) error {
	if !h.hasApplicableRules(originalQuery, *resp) {
		return nil
	}

	log.Printf("Found applicable rules for query for query '%s': ", originalQuery)

	if err := h.enrichMetrics(resp, originalQuery); err != nil {
		return err
	}

	allLabels := h.getAllLabels(originalQuery)
	h.aggregateMetrics(resp, allLabels)

	return nil
}

func (h *EnrichmentUseCase) enrichMetrics(resp *domain.QueryResponse, originalQuery string) error {
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

		h.applyRule(resp, &rule, mappings, originalQuery)
	}
	return nil
}

func (h *EnrichmentUseCase) applyRule(resp *domain.QueryResponse, rule *domain.EnrichmentRule, mappings map[string]domain.SourceData, originalQuery string) {
	for i, r := range resp.Data.Result {
		if !h.matchesMetric(r.Metric, rule.Match, originalQuery) {
			continue
		}

		labelValue := r.Metric[rule.Match.Label]
		if labelValue == "" {
			continue
		}

		matchedData := h.findMatchingData(labelValue, mappings)
		h.applyLabels(resp, i, matchedData, rule)
	}
}

func (h *EnrichmentUseCase) findMatchingData(labelValue string, mappings map[string]domain.SourceData) *domain.SourceData {
	for pattern, data := range mappings {
		if pattern == labelValue {
			return &data
		}
		if matched, _ := regexp.MatchString(pattern, labelValue); matched {
			return &data
		}
	}
	return nil
}

func (h *EnrichmentUseCase) applyLabels(resp *domain.QueryResponse, index int, matchedData *domain.SourceData, rule *domain.EnrichmentRule) {
	if matchedData != nil {
		for _, label := range rule.AddLabels {
			if value, ok := matchedData.Labels[label]; ok {
				resp.Data.Result[index].Metric[label] = value
			}
		}
	} else {
		for label, value := range rule.Fallback {
			resp.Data.Result[index].Metric[label] = value
		}
	}
}

func (h *EnrichmentUseCase) aggregateMetrics(resp *domain.QueryResponse, allLabels []string) {
	switch resp.Data.ResultType {
	case "matrix":
		h.aggregateMatrixMetrics(resp, allLabels)
	case "vector":
		h.aggregateVectorMetrics(resp, allLabels)
	}
}

func (h *EnrichmentUseCase) aggregateMatrixMetrics(resp *domain.QueryResponse, allLabels []string) {
	groupedMetrics := make(map[string][][]interface{})
	for _, r := range resp.Data.Result {
		groupKey := h.buildGroupKey(r.Metric, allLabels)
		if groupKey == "" {
			continue
		}

		if values, exists := groupedMetrics[groupKey]; exists {
			h.mergeMatrixValues(values, r.Values)
		} else {
			groupedMetrics[groupKey] = r.Values
		}
	}

	resp.Data.Result = h.createMatrixResult(groupedMetrics)
}

func (h *EnrichmentUseCase) aggregateVectorMetrics(resp *domain.QueryResponse, allLabels []string) {
	groupedMetrics := make(map[string][]interface{})
	for _, r := range resp.Data.Result {
		groupKey := h.buildGroupKey(r.Metric, allLabels)
		if groupKey == "" {
			continue
		}

		if value, exists := groupedMetrics[groupKey]; exists {
			h.mergeVectorValues(value, r.Value)
		} else {
			groupedMetrics[groupKey] = r.Value
		}
	}

	resp.Data.Result = h.createVectorResult(groupedMetrics)
}

func (h *EnrichmentUseCase) buildGroupKey(metric map[string]string, labels []string) string {
	groupKey := make(map[string]string)
	for _, label := range labels {
		if value, ok := metric[label]; ok {
			groupKey[label] = value
		}
	}

	if len(groupKey) == 0 {
		return ""
	}

	return h.createGroupKey(groupKey)
}

func (h *EnrichmentUseCase) mergeMatrixValues(existing, values [][]interface{}) {
	for i, v := range values {
		if i >= len(existing) {
			existing = append(existing, v)
		} else {
			val1, _ := strconv.Atoi(v[1].(string))
			val2, _ := strconv.Atoi(existing[i][1].(string))
			existing[i][1] = strconv.Itoa(val1 + val2)
		}
	}
}

func (h *EnrichmentUseCase) mergeVectorValues(existing, values []interface{}) {
	val1, _ := strconv.Atoi(values[1].(string))
	val2, _ := strconv.Atoi(existing[1].(string))
	existing[1] = strconv.Itoa(val1 + val2)
}

func (h *EnrichmentUseCase) createMatrixResult(groupedMetrics map[string][][]interface{}) []domain.MetricData {
	result := make([]domain.MetricData, 0, len(groupedMetrics))
	for groupKey, values := range groupedMetrics {
		result = append(result, domain.MetricData{
			Metric: h.parseGroupKey(groupKey),
			Values: values,
		})
	}
	return result
}

func (h *EnrichmentUseCase) createVectorResult(groupedMetrics map[string][]interface{}) []domain.MetricData {
	result := make([]domain.MetricData, 0, len(groupedMetrics))
	for groupKey, value := range groupedMetrics {
		result = append(result, domain.MetricData{
			Metric: h.parseGroupKey(groupKey),
			Value:  value,
		})
	}
	return result
}

func (h *EnrichmentUseCase) parseGroupKey(groupKey string) map[string]string {
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
	return metric
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

func (h *EnrichmentUseCase) getAllLabels(query string) []string {
	labelSet := make(map[string]bool)
	for _, rule := range h.config.Enrichment.Rules {
		if strings.Contains(query, rule.Match.Metric) {
			for _, label := range rule.AddLabels {
				labelSet[label] = true
			}
		}
	}

	labels := make([]string, 0, len(labelSet))
	for label := range labelSet {
		labels = append(labels, label)
	}
	return labels
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

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, groupKey[k]))
	}
	return strings.Join(parts, ",")
}
