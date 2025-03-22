# Labelify

**Labelify** is a lightweight Prometheus proxy that injects labels into PromQL query results — enabling more insightful dashboards, smarter alerts, and clearer operational context, without modifying your original metrics or ingestion configs.

## What does Labelify do?

Let's suppose you have a series of replicas running on your cluster:

```
$ sum(kube_deployment_spec_replicas) by (deployment)

{deployment="microservice-1"}           1
{deployment="microservice-2"}           1
{deployment="microservice-3"}           1
{deployment="prometheus"}               1
{deployment="coredns"}                  1
```

And you intend to write Labelify rules to group deployments by team:

```yml
rules:
  - mutate:
      type: "static"
      target_label: "team"
      default_value: "engineering-team"
      matchers:
        - match:
            deployment: "prometheus"
          replace: "observability-team"
        - match:
            deployment: "coredns"
          replace: "networking-team"
```

Enriched response from Labelify:

```
{team="engineering-team"}       3
{team="observability-team"}     1
{team="networking-team"}        1
```

Now your dashboards and alerts can group deployments by responsible team, without needing to change how metrics are collected or creating label replace rules.

We currently support both result type: [instant vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors) and [range vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#range-vectors).