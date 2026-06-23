# Observability Topology

Grafana and Loki are platform-cluster services in this reference architecture.

## Platform Cluster

`applicationsets/platform-observability.yaml` creates two Argo CD Applications that target the Argo CD management cluster:

```yaml
destination:
  server: https://kubernetes.default.svc
  namespace: observability
```

That means these services are not generated once per workload cluster:

- `platform-loki`
- `platform-grafana`

## Workload Clusters

`applicationsets/platform-helm-addons-appset.yaml` installs `kube-prometheus-stack` on each cluster from `clusters/*.yaml`, but disables the chart's Grafana subchart:

```yaml
grafana:
  enabled: false
```

This gives each workload cluster local metrics components while keeping the dashboard UI centralized on the platform cluster.

`applicationsets/observability-agents-appset.yaml` also installs Grafana Alloy on each workload cluster to discover pod logs and ship them to platform Loki.

## Future Extension

For a larger fleet, workload clusters can remote-write metrics or ship logs to the platform cluster. The same topology can later add Mimir, Alloy, Promtail, OpenTelemetry Collector, or remote-write configuration without changing where Grafana lives.