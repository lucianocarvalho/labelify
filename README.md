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

Let's suppose you have a series of replicas running on your cluster:

```
$ sum(kube_deployment_spec_replicas) by (deployment)

{deployment="microservice-1"}                             1
{deployment="microservice-2"}                             1
{deployment="microservice-3"}                             1
{deployment="prometheus-grafana"}                         1
{deployment="prometheus-kube-prometheus-operator"}        1
{deployment="coredns"}                                    1
```

And you intend to write Labelify rules to group deployments by team:

```yml
sources:
  - name: static_map
    type: yaml
    mappings:
      # Using exact match for coredns
      coredns:
        labels:
          team: networking
          business_unit: platform

      # Using wildcard to match all prometheus deployments
      prometheus-.*:
        labels:
          team: observability
          business_unit: platform

enrichment:
  rules:
    - match:
        metric: "kube_deployment_spec_replicas"
        label: "deployment"
      enrich_from: static_map
      add_labels:
        - team
        - business_unit
      fallback:
        team: "unknown"
        business_unit: "n/a"
```

Enriched response from Labelify:

```
{team="observability", business_unit="platform"}        2
{team="networking", business_unit="platform"}           1
{team="unknown", business_unit="n/a"}                   3
```

Now your dashboards and alerts can group deployments by responsible team, without needing to change how metrics are collected or creating label replace rules.

We currently support both [instant vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors) and [range vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#range-vectors).

If no rule matches the executed query, seamlessly falls back to acting as a transparent Prometheus-agnostic proxy, — forwarding any query without interfering in your results.

## Features

**What Labelify can do (and what’s coming soon):**

- Create new labels into your query results
- Aggregate results dynamically based in your current labels
- Creating conditions using expressions and templates (coming soon)

**Supported data for rules:**

- Static mappings (defined in config.yaml)
- External APIs (coming soon)
- Other prometheus queries (coming soon)

## License

This library is licensed under the [MIT License](LICENSE).