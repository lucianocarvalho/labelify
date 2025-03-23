<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./.github/assets/labelify-logo-dark-mode.svg">
    <source media="(prefers-color-scheme: light)" srcset="./.github/assets/labelify-logo-light-mode.svg">
    <img alt="Logo" src="./.github/assets/labelify-logo-light-mode.svg" width="300">
  </picture>
</p>

# Labelify

Labelify is a lightweight, Prometheus-compatible proxy that enhances your PromQL query results using dynamic, rule-based label enrichment, enabling more insightful dashboards, smarter alerts, and clearer operational context - without modifying your original metrics or creating ingestion configs.

## What does Labelify do?

Let's suppose you have a series of deployments running replicas on your cluster:

```
promql> sum(kube_deployment_spec_replicas) by (deployment)

{deployment="microservice-1"}                             1
{deployment="microservice-2"}                             1
{deployment="microservice-3"}                             1
{deployment="prometheus-grafana"}                         1
{deployment="prometheus-kube-prometheus-operator"}        1
```

And instead of listing the deployments directly, you might want to define an aggregation where:
- All deployments starting with `prometheus-.*` belongs to `team="observability"`
- All deployments starting with `microservices-.*` belongs to `team="engineering"`

Instead of adding these labels directly to the ingestion pipeline using [relabel config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config), you can just create a Labelify's `mapping` specifying these rules:

```yml
sources:
  - name: awesome-static-labels     # <-- Source name (allowing having multiple sources)
    type: yaml                      # <-- `yaml` means a static yaml
    mappings:                       # <-- List of mappings (matchers)
      microservices-.*:             # <----- Wildcard for microservices-.*
        labels:                     # <-------- Set of labels that can be used later
          team: engineering         # <----------- Team responsible for microservices
      prometheus-.*:                # <----- Wildcard for prometheus-.*
        labels:                     # <-------- Set of labels that can be used later
          team: observability       # <----------- Team responsible for prometheus
```

You can have different sources (static and dynamic ones). These sources are responsible for just creating labels given a pattern (`mappings` indexes), which will later can be used in queries. Feel free to create as many labels as you want to represent the specified index (eg: `team`, `business_unit`, `cost_center`).

With the mappings registered, we can now attach the sources to the queries:

```yml
enrichment:
  rules:                                           # <-- List of rules
    - match:                                       # <---- Match config
        metric: "kube_deployment_spec_replicas"    # <------- Enrich this metric
        label: "deployment"                        # <------- Rewriting this label
      enrich_from: awesome-static-labels           # <---- Using this source name
      add_labels:                                  
        - team                                     # <---- To this label
```

This means that every time we have the `deployment` label as a response when running a query on metric `kube_deployment_spec_replicas`, Labelify is gonna replace the `deployment` label with the previously created `team` label:

```
promql> sum(kube_deployment_spec_replicas) by (deployment)

{team="engineering"}                3
{team="observability"}              2
```

Labelify also supports **dynamic sources** 🎉. This means that you can add labels to your queries at runtime, allowing you to use labels using your catalog sources dynamically (IDP, catalog-info.yaml, GitHub repository).

```yml
sources:
  - name: awesome-catalog-service
    type: http
    config:
      url: https://run.mocky.io/v3/ba325f0c-f98e-4584-a4ec-966cecd3a773
      method: GET
      refresh_interval: 60s
```

Just like in yaml, Labelify expects the response from this endpoint to look something like this:
```json
{
  "microservice-.*": {
    "labels": {
      "team": "engineering"
    }
  },
  "prometheus-.*": {
    "labels": {
      "team": "observability"
    }
  }
}
```

> If you want step by step practical examples of how it works [click here to check out `enrichment-rules-examples.md`](./docs/enrichment-rules-examples.md). 

You can **always** send promql-compatible queries to Labelify, whether they have rules or not. If no rule matches the executed query, seamlessly falls back to acting as a transparent Prometheus-agnostic proxy - forwarding any query without interfering in your results. 

We currently support both [instant vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors) and [range vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#range-vectors).

## ✨ Features

**What Labelify can do (and what’s coming soon):**

- Rewrite new labels into your query results
- Aggregate results dynamically based in your current labels
- Creating conditions using expressions and templates (coming soon)

**Supported sources for rules:**

- Static yaml mappings
- External APIs
- Other prometheus queries (coming soon)

# 🚀 Installation

There are various ways of installing Labelify.

### Using Docker

You can quickly get started using Docker:

```bash
# Run the container as a proxy
docker run -d \
  -p 8080:8080 \
  -v ./examples/config.yaml:/etc/labelify/config.yaml \
  ghcr.io/lucianocarvalho/labelify:latest
```
> **⚠️ Important:** You need to create your own config.yaml with the enrichment rules and label mappings. The default configuration in this example is just proxying queries to http://prometheus:9090/.

### Using Kubernetes

Simply run:

```bash
curl -s https://raw.githubusercontent.com/lucianocarvalho/labelify/main/k8s/manifest.yaml | kubectl apply -f -
```

You should get an output like this:
```  
namespace/labelify created
configmap/labelify-config created
deployment.apps/labelify created
service/labelify created
horizontalpodautoscaler.autoscaling/labelify created
```

> **⚠️ Important:** Don't forget to configure your prometheus url inside the `configmap/labelify-config`. The default configuration in this example is just proxying queries to http://prometheus.monitoring.svc.cluster.local:9090/.

### Running locally

```bash
# Start by cloning the repository
git clone https://github.com/lucianocarvalho/labelify.git
cd labelify

# Running main.go
go run cmd/api/main.go --config.file="$(PWD)/examples/config.yaml"
```

## License

This library is licensed under the [MIT License](LICENSE).