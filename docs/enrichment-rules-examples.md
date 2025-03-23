# Examples and Use Cases

This document shows real-world examples of how to use enrichment rules in Labelify to enhance your Prometheus queries.

## Default rewriting

Use it when you want to return **only** metrics that match the matchers you configured.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      coredns:                  # <-- Matcher
        labels:
          team: networking      # <-- Rewrite

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
```promql
$ sum(kube_deployment_spec_replicas) by (deployment)

{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**
```
{team="networking"}                                         1
```

**Why?**

Since the source only has `coredns` as a source, all remaining deployments are ignored.

## Wildcard

Use when you have multiple deployments with the same pattern that will have the same rewrite rule.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      prometheus-.*:            # <-- Wildcard
        labels:
          team: observability

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
```promql
$ sum(kube_deployment_spec_replicas) by (deployment)

# Normal response
{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**
```
{team="observability"}                                      3
```

**Why?**

It will respect the wildcard: everything that starts with `prometheus-` will be aggregated into the observability team. The other deployments are ignored.

## Fallback

Use when you want to return other data aggregated in another subgroup.

**config.yaml:**
```yaml
sources:
  - name: static_map
    type: yaml
    mappings:
      prometheus-.*:
        labels:
          team: observability

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
```promql
$ sum(kube_deployment_spec_replicas) by (deployment)

# Normal response
{deployment="prometheus-grafana"}                           1
{deployment="prometheus-kube-prometheus-operator"}          1
{deployment="prometheus-kube-state-metrics"}                1
{deployment="coredns"}                                      1
{deployment="local-path-provisioner"}                       1
```

**Labelify response:**
```
{team="observability"}                                      3
{team="unknown"}                                            2
```

**Why?**

All deployments that don't have matchers will have their values aggregated within the label configured in `fallback`.
