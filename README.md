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

> If you want to see all practical and possible enrichment rule examples, [click here to check out `enrichment-rules-examples.md`](./docs/enrichment-rules-examples.md).

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

      # Using wildcard to match all prometheus deployments
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
      fallback:
        team: "unknown"
        business_unit: "n/a"
```

Enriched response from Labelify:

```
{team="observability"}              2        # `grafana` and `operator`
{team="networking"}                 1        # `coredns`
{team="unknown"}                    3        # `microservices-(1|2|3)`
```

Now your dashboards and alerts can group deployments by responsible team, without needing to change how metrics are collected or creating label replace rules.

You can send all promql-compatible queries to Labelify, whether they have rules or not. If no rule matches the executed query (`match.metric`), seamlessly falls back to acting as a transparent Prometheus-agnostic proxy - forwarding any query without interfering in your results. 

We currently support both [instant vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors) and [range vectors](https://prometheus.io/docs/prometheus/latest/querying/api/#range-vectors).

## ✨ Features

**What Labelify can do (and what’s coming soon):**

- Rewrite new labels into your query results
- Aggregate results dynamically based in your current labels
- Creating conditions using expressions and templates (coming soon)

**Supported data for rules:**

- Static mappings (defined in config.yaml)
- External APIs (coming soon)
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
  lucianoajunior/labelify:latest
```
> **⚠️ Important:** You need to create your own config.yaml with the enrichment rules and label mappings. The default configuration in this example is just proxying queries to http://prometheus:9090/.

### Using Kubernetes

Simply run:

```bash
curl -s https://raw.githubusercontent.com/lucianocarvalho/labelify/main/k8s/manifest.yaml | kubectl apply -f -
```

You should see output like this:
```  
namespace/labelify created
configmap/labelify-config created
deployment.apps/labelify created
service/labelify created
```

Feel free to browse the resources:

```bash
kubectl get all -n labelify
```

> **⚠️ Important:** Don't forget to configure your prometheus url inside the `configmap/labelify-config`. The default configuration in this example is just proxying queries to http://prometheus.monitoring.svc.cluster.local:9090/.

After configuring it correctly you can test it by running a command like this:

```bash
# Port-forwarding
kubectl port-forward service/labelify -n labelify 8080:80

# Testing the proxy
curl -XGET http://localhost:8080/api/v1/query?query=time()
```

### Running locally

```bash
# Start by cloning the repository:
git clone https://github.com/lucianocarvalho/labelify.git
cd labelify

# Running main.go
go run cmd/api/main.go --config.file="$(PWD)/examples/config.yaml"
```

## License

This library is licensed under the [MIT License](LICENSE).