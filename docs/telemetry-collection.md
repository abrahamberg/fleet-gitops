# Telemetry Collection

Central Grafana and Loki are not enough by themselves. Each workload cluster also needs an agent that discovers workloads and ships telemetry.

This reference uses Grafana Alloy as the per-cluster observability agent. Alloy can run Loki, Prometheus, and OpenTelemetry-style pipelines, which makes it a good fit when the platform uses Grafana and Loki.

## Logs

`applicationsets/observability-agents-appset.yaml` installs Alloy as a DaemonSet in every cluster from `clusters/*.yaml`.

The agent discovers Kubernetes pods and sends pod logs to the platform Loki gateway with labels for:

- `cluster`
- `environment`
- `namespace`
- `pod`
- `container`

Each cluster declares the Loki push endpoint in `clusters/*.yaml`:

```yaml
lokiEndpoint: http://loki-gateway.observability.svc.cluster.local/loki/api/v1/push
```

For real separate clusters, expose Loki through an internal ingress, private load balancer, service mesh, VPN, or another reachable endpoint. Workload clusters cannot normally resolve the platform cluster's `cluster.local` DNS name.

## Metrics

Metrics are split into two layers:

- `kube-prometheus-stack` runs per workload cluster and collects Kubernetes metrics, Prometheus rules, ServiceMonitors, and PodMonitors.
- Central Grafana runs only on the platform cluster.

To view all workload metrics centrally, choose one of these patterns:

- Add each workload cluster Prometheus as a Grafana datasource if the platform cluster can reach it.
- Add Mimir on the platform cluster and remote-write metrics from each workload cluster.
- Extend Alloy or OpenTelemetry Collector pipelines to scrape and remote-write metrics to a central backend.

For a top-tier fleet architecture, Mimir plus per-cluster Alloy remote write is the most scalable central metrics pattern.

## Do We Need OpenTelemetry Collector?

Use OpenTelemetry Collector when you want vendor-neutral OTLP pipelines, especially for traces and application telemetry.

Use Grafana Alloy when your platform already centers on Grafana, Loki, Prometheus, and future Mimir. Alloy can still receive and forward OTLP data, but it integrates naturally with Loki and Prometheus-style discovery.