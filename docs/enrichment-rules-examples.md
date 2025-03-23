# Examples and Use Cases

This document shows real-world examples of how to use enrichment rules in Labelify to enhance your Prometheus queries.

## Default rewriting

Use it when you want to return **only** metrics that match the matchers you configured.

**config.yaml:**
```yaml
sources:
  - name: static_map                                 # <-- Source name
    type: yaml
    mappings:
      coredns:                                       # <-- Matcher
        labels:
          team: networking                           # <-- Rewrite

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"     # <-- Metric name
        label: "deployment"                         # <-- Label that will be overwritten
      enrich_from: static_map                       # <-- Source
      add_labels:
        - team
```

**Default query/response:**
```
promql> sum(kube_deployment_spec_replicas) by (deployment)

{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**

Since the source only has `coredns` as a mapping, all remaining deployments **are ignored**.

```
{team="networking"}                                         1
```

## Wildcard

Use when you have multiple deployments with the same pattern that will have the same rewrite rule.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      prometheus-.*:              # <-- Wildcard
        labels:
          team: observability     # <-- Label that will be overwritten

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - team
```

**Default query/response:**
```
promql> sum(kube_deployment_spec_replicas) by (deployment)

# Normal response
{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**

It will respect the wildcard: everything that starts with `prometheus-` will be aggregated into the observability team. The other results **are ignored**.

```
{team="observability"}                                      3
```

## Fallback

Use when you want to return other data aggregated in a fallback group.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      prometheus-.*:            # <-- Wildcard
        labels:
          team: observability   # <-- Label that will be overwritten

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - team
      fallback:                 # <-- Fallback
        team: "unknown"         # <-- Default value
```

**Default query/response:**
```
promql> sum(kube_deployment_spec_replicas) by (deployment)

# Normal response
{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**

All deployments that don't have matchers will have their values aggregated within the label configured in `fallback`.

```
{team="observability"}                                      3
{team="unknown"}                                            2
```

## Selecting specific label

Use when you want to have more information for each mapping, but want to return a different result for each query.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      prometheus-.*:                              # <-- Wildcard
        labels:
          team: observability                     # <-- Label `team`
          business_unit: foundation               # <-- Label `business_unit`

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"   # <-- Query 1
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - team                                    # <-- Using `team` label

    - match:
        metric: "kube_deployment_created"         # <-- Query 2
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - business_unit                           # <-- Using `business_unit` label
```

**Labelify response:**

The source map is the same. It will add the label `team` when the query is `kube_deployment_spec_replicas`, and it will add the label `business_unit` when the query is `kube_deployment_created`.

```
promql> kube_deployment_spec_replicas
{team="observability"}                                      3

promql> kube_deployment_created
{business_unit="foundation"}                                5227825434
```

## Dynamic sources

You can also have dynamic label configurations.

**config.yaml:**
```yaml
sources:
  - name: dynamic_map
    type: http
    config:
      url: https://run.mocky.io/v3/ba325f0c-f98e-4584-a4ec-966cecd3a773
      method: GET
      refresh_interval: 60s

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - team
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
