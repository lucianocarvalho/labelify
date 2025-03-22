package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/lucianocarvalho/labelify/internal/domain"
)

type HydrateUseCase struct {
	rules *domain.RuleSet
}

func NewHydrateUseCase(rules *domain.RuleSet) *HydrateUseCase {
	return &HydrateUseCase{
		rules: rules,
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

	for _, rule := range h.rules.Rules {
		log.Printf("Evaluating rule: %s", rule.Name)
		log.Printf("Mutation type: %s", rule.Mutate.Type)

		for i, r := range resp.Data.Result {
			log.Printf("Processing metric %d: %v", i+1, r.Metric)
			found := false

			for _, matcher := range rule.Mutate.Matchers {
				log.Printf("Trying to match with: %v", matcher.MatchLabels)
				if h.matchEnrichment(r.Metric, matcher.MatchLabels) {
					log.Printf("Match found! Adding label %s=%s",
						rule.Mutate.TargetLabel, matcher.Value)
					resp.Data.Result[i].Metric[rule.Mutate.TargetLabel] = matcher.Value
					found = true
					break
				}
			}

			if !found && rule.Mutate.DefaultValue != "" {
				log.Printf("No match found, using default value: %s",
					rule.Mutate.DefaultValue)
				resp.Data.Result[i].Metric[rule.Mutate.TargetLabel] = rule.Mutate.DefaultValue
			}
		}
	}

	switch resp.Data.ResultType {
	case "matrix":
		groupedMetrics := make(map[string][][]interface{})
		for _, r := range resp.Data.Result {
			team, ok := r.Metric["team"]
			if !ok {
				continue
			}

			if values, exists := groupedMetrics[team]; exists {
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
				groupedMetrics[team] = r.Values
			}
		}

		var newResult []domain.MetricData
		for team, values := range groupedMetrics {
			newResult = append(newResult, domain.MetricData{
				Metric: map[string]string{
					"team": team,
				},
				Values: values,
			})
		}
		resp.Data.Result = newResult

	case "vector":
		groupedMetrics := make(map[string][]interface{})
		for _, r := range resp.Data.Result {
			team, ok := r.Metric["team"]
			if !ok {
				continue
			}

			if value, exists := groupedMetrics[team]; exists {
				val1, _ := strconv.Atoi(r.Value[1].(string))
				val2, _ := strconv.Atoi(value[1].(string))
				value[1] = strconv.Itoa(val1 + val2)
			} else {
				groupedMetrics[team] = r.Value
			}
		}

		var newResult []domain.MetricData
		for team, value := range groupedMetrics {
			newResult = append(newResult, domain.MetricData{
				Metric: map[string]string{
					"team": team,
				},
				Value: value,
			})
		}
		resp.Data.Result = newResult
	}

	return json.Marshal(resp)
}

func (h *HydrateUseCase) matchEnrichment(metric, matchLabels map[string]string) bool {
	for k, v := range matchLabels {
		val, ok := metric[k]
		if !ok || val != v {
			log.Printf("No match: %s=%s (expected: %s)", k, val, v)
			return false
		}
		log.Printf("Match: %s=%s", k, v)
	}
	return true
}
