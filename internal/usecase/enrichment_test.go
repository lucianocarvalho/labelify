package usecase

import (
	"reflect"
	"testing"

	"github.com/lucianocarvalho/labelify/internal/domain"
)

func TestEnrichmentUseCase_Execute(t *testing.T) {
	// As a proxy, it modifies the original response as pointer
	// so we need to always create a new one for each test.
	createResponse := func() domain.QueryResponse {
		return domain.QueryResponse{
			Status: "success",
			Data: domain.QueryData{
				ResultType: "vector",
				Result: []domain.MetricData{
					{
						Metric: map[string]string{"deployment": "coredns"},
						Value: []interface{}{
							float64(182778586.0),
							"1",
						},
					},
					{
						Metric: map[string]string{"deployment": "microservice-1"},
						Value: []interface{}{
							float64(182778586.0),
							"2",
						},
					},
					{
						Metric: map[string]string{"deployment": "microservice-2"},
						Value: []interface{}{
							float64(182778586.0),
							"3",
						},
					},
				},
			},
		}
	}

	t.Run("given an instant vector returning a list of deployments", func(t *testing.T) {
		query := "sum(kube_deployment_spec_replicas) by (deployment)"

		enrichment := domain.Enrichment{
			Rules: []domain.EnrichmentRule{
				{
					Match:      domain.MatchRule{Metric: "kube_deployment_spec_replicas", Label: "deployment"},
					EnrichFrom: "just-a-random-source",
					AddLabels:  []string{"team"},
				},
			},
		}

		t.Run("when calling enrichment having a source only to coredns matcher", func(t *testing.T) {
			response := createResponse()
			sources := []domain.Source{
				{
					Name: "just-a-random-source",
					Type: "yaml",
					Mappings: map[string]domain.SourceData{
						"coredns": {
							Labels: map[string]string{
								"team": "networking",
							},
						},
					},
				},
			}

			config := &domain.Config{
				Sources:    sources,
				Enrichment: enrichment,
			}

			uc, err := NewEnrichmentUseCase(config)
			if err != nil {
				t.Fatal(err)
			}

			err = uc.Execute(&response, query)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("then it should return only team networking as a result", func(t *testing.T) {
				expected := []domain.MetricData{
					{
						Metric: map[string]string{
							"team": "networking",
						},
						Value: []interface{}{
							float64(182778586.0),
							"1",
						},
					},
				}

				if !reflect.DeepEqual(response.Data.Result, expected) {
					t.Fatalf("expected %+v, got %+v", expected, response.Data.Result)
				}
			})

			t.Run("then it should return one result", func(t *testing.T) {
				if len(response.Data.Result) != 1 {
					t.Fatalf("expected 1 result, got %d", len(response.Data.Result))
				}
			})
		})

		t.Run("when calling enrichment having a source regex to microservice", func(t *testing.T) {
			response := createResponse()
			sources := []domain.Source{
				{
					Name: "just-a-random-source",
					Type: "yaml",
					Mappings: map[string]domain.SourceData{
						"microservice-.*": {
							Labels: map[string]string{
								"team": "engineering",
							},
						},
					},
				},
			}

			config := &domain.Config{
				Sources:    sources,
				Enrichment: enrichment,
			}

			uc, err := NewEnrichmentUseCase(config)
			if err != nil {
				t.Fatal(err)
			}

			err = uc.Execute(&response, query)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("then it should return only team engineering, but aggregating the value", func(t *testing.T) {
				expected := []domain.MetricData{
					{
						Metric: map[string]string{
							"team": "engineering",
						},
						Value: []interface{}{
							float64(182778586.0),
							"5", // 3 + 2
						},
					},
				}

				if !reflect.DeepEqual(response.Data.Result, expected) {
					t.Fatalf("expected %+v, got %+v", expected, response.Data.Result)
				}
			})

			t.Run("then it should return one result", func(t *testing.T) {
				if len(response.Data.Result) != 1 {
					t.Fatalf("expected 1 result, got %d", len(response.Data.Result))
				}
			})
		})

		t.Run("when calling enrichment having both regex and static sources", func(t *testing.T) {
			response := createResponse()
			sources := []domain.Source{
				{
					Name: "just-a-random-source",
					Type: "yaml",
					Mappings: map[string]domain.SourceData{
						"microservice-.*": {
							Labels: map[string]string{
								"team": "engineering",
							},
						},
						"coredns": {
							Labels: map[string]string{
								"team": "networking",
							},
						},
					},
				},
			}

			config := &domain.Config{
				Sources:    sources,
				Enrichment: enrichment,
			}

			uc, err := NewEnrichmentUseCase(config)
			if err != nil {
				t.Fatal(err)
			}

			err = uc.Execute(&response, query)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("then it should return both results", func(t *testing.T) {
				expected := []domain.MetricData{
					{
						Metric: map[string]string{
							"team": "networking",
						},
						Value: []interface{}{
							float64(182778586.0),
							"1",
						},
					},
					{
						Metric: map[string]string{
							"team": "engineering",
						},
						Value: []interface{}{
							float64(182778586.0),
							"5", // 3 + 2
						},
					},
				}

				if !reflect.DeepEqual(response.Data.Result, expected) {
					t.Fatalf("expected %+v, got %+v", expected, response.Data.Result)
				}
			})

			t.Run("then it should return one result", func(t *testing.T) {
				if len(response.Data.Result) != 2 {
					t.Fatalf("expected 2 result, got %d", len(response.Data.Result))
				}
			})
		})

		t.Run("when calling enrichment having multiple tags", func(t *testing.T) {
			response := createResponse()
			sources := []domain.Source{
				{
					Name: "just-a-random-source",
					Type: "yaml",
					Mappings: map[string]domain.SourceData{
						"coredns": {
							Labels: map[string]string{
								"team":          "networking",
								"business_unit": "foundation",
							},
						},
					},
				},
			}

			config := &domain.Config{
				Sources:    sources,
				Enrichment: enrichment,
			}

			uc, err := NewEnrichmentUseCase(config)
			if err != nil {
				t.Fatal(err)
			}

			err = uc.Execute(&response, query)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("then it should return only team networking as a result", func(t *testing.T) {
				expected := []domain.MetricData{
					{
						Metric: map[string]string{
							"team": "networking",
						},
						Value: []interface{}{
							float64(182778586.0),
							"1",
						},
					},
				}

				if !reflect.DeepEqual(response.Data.Result, expected) {
					t.Fatalf("expected %+v, got %+v", expected, response.Data.Result)
				}
			})

			t.Run("then it should return one result", func(t *testing.T) {
				if len(response.Data.Result) != 1 {
					t.Fatalf("expected 1 result, got %d", len(response.Data.Result))
				}
			})
		})
	})

	t.Run("given an instant vector but using a fallback config", func(t *testing.T) {
		query := "sum(kube_deployment_spec_replicas) by (deployment)"

		enrichment := domain.Enrichment{
			Rules: []domain.EnrichmentRule{
				{
					Match:      domain.MatchRule{Metric: "kube_deployment_spec_replicas", Label: "deployment"},
					EnrichFrom: "just-a-random-source",
					AddLabels:  []string{"team"},
					// Fallback is used when the source does not have a match.
					Fallback: map[string]string{
						"team": "unknown",
					},
				},
			},
		}

		t.Run("when calling enrichment having a source only to coredns matcher", func(t *testing.T) {
			response := createResponse()
			sources := []domain.Source{
				{
					Name: "just-a-random-source",
					Type: "yaml",
					Mappings: map[string]domain.SourceData{
						"coredns": {
							Labels: map[string]string{
								"team": "networking",
							},
						},
					},
				},
			}

			config := &domain.Config{
				Sources:    sources,
				Enrichment: enrichment,
			}

			uc, err := NewEnrichmentUseCase(config)
			if err != nil {
				t.Fatal(err)
			}

			err = uc.Execute(&response, query)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("then it should return the coredns and fallback results", func(t *testing.T) {
				expected := []domain.MetricData{
					{
						Metric: map[string]string{
							"team": "networking",
						},
						Value: []interface{}{
							float64(182778586.0),
							"1",
						},
					},
					{
						Metric: map[string]string{
							"team": "unknown",
						},
						Value: []interface{}{
							float64(182778586.0),
							"5",
						},
					},
				}

				if !reflect.DeepEqual(response.Data.Result, expected) {
					t.Fatalf("expected %+v, got %+v", expected, response.Data.Result)
				}
			})

			t.Run("then it should return both results", func(t *testing.T) {
				if len(response.Data.Result) != 2 {
					t.Fatalf("expected 2 result, got %d", len(response.Data.Result))
				}
			})
		})
	})
}
