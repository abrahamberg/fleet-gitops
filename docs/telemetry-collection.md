# Telemetry Collection

This page explains how workload telemetry leaves each cluster and reaches the platform observability stack.

Read [observability-topology.md](observability-topology.md) first if you want the platform-vs-workload cluster picture.

## Logs

[../applicationsets/observability-agents-appset.yaml](../applicationsets/observability-agents-appset.yaml) installs Grafana Alloy as a DaemonSet in every cluster from [../clusters/](../clusters/).

Alloy runs on each node, discovers pods, reads container logs, and sends them to Loki.

```text
workload pod logs -> Alloy DaemonSet -> Loki push endpoint -> platform Loki -> Grafana
```

The Alloy config adds labels that make logs usable in a fleet:

| Label | Source |
| --- | --- |
| `cluster` | Cluster inventory file, such as [../clusters/dev.yaml](../clusters/dev.yaml). |
| `environment` | Cluster inventory file. |
| `namespace` | Kubernetes pod metadata. |
| `pod` | Kubernetes pod metadata. |
| `container` | Kubernetes pod metadata. |

Each cluster declares where Alloy should push logs:

```yaml
lokiEndpoint: http://loki-gateway.observability.svc.cluster.local/loki/api/v1/push
```

That endpoint is fine for the local reference shape. In real separate clusters, expose Loki through a reachable private path such as internal ingress, private load balancer, service mesh, VPN, or a managed endpoint. A workload cluster usually cannot resolve another cluster's `cluster.local` DNS name.

## Metrics

Metrics are split between local collection and central viewing.

| Component | Where it runs | Purpose |
| --- | --- | --- |
| kube-prometheus-stack | Every workload cluster, from [../applicationsets/platform-helm-addons-appset.yaml](../applicationsets/platform-helm-addons-appset.yaml) | Collects Kubernetes metrics, Prometheus rules, ServiceMonitors, and PodMonitors. |
| Grafana | Management cluster, from [../applicationsets/platform-observability.yaml](../applicationsets/platform-observability.yaml) | Central UI. |
| Loki | Management cluster, from [../applicationsets/platform-observability.yaml](../applicationsets/platform-observability.yaml) | Central log store for this reference. |

The fleet `kube-prometheus-stack` values disable Grafana on workload clusters:

```yaml
grafana:
	enabled: false
```

That avoids one Grafana per cluster while keeping local Prometheus components.

## Central Metrics Options

This repo stops at local Prometheus plus central Grafana. For a larger fleet, choose one of these patterns:

| Pattern | When it fits |
| --- | --- |
| Add each workload Prometheus as a Grafana datasource | Small fleet where the platform cluster can reach each Prometheus. |
| Add Grafana Mimir and remote-write metrics | Larger fleet with central metrics storage. |
| Extend Alloy or OpenTelemetry Collector pipelines | Platform wants one agent path for logs, metrics, and application telemetry. |

## Alloy Or OpenTelemetry Collector

Use Grafana Alloy when the platform is centered on Grafana, Loki, Prometheus, and possibly Mimir. It integrates naturally with Kubernetes discovery and Loki log pipelines.

Use OpenTelemetry Collector when you need vendor-neutral OTLP pipelines, especially for traces and application telemetry. Alloy can still receive and forward OTLP data, but OpenTelemetry Collector is the common neutral choice for trace-heavy platforms.