# Platform Add-ons

This directory is an educational reference for bootstrapping common platform capabilities across multiple Kubernetes clusters with Argo CD ApplicationSets.

The namespace add-on uses Kustomize overlays per environment. The fleet Helm add-ons are generated from `applicationsets/platform-helm-addons-appset.yaml` for every cluster declared in `clusters/*.yaml`. Central observability services for the platform cluster are defined in `applicationsets/platform-observability.yaml`.

Included add-ons:

- `cert-manager` for certificate lifecycle automation.
- `ingress-nginx` for inbound HTTP and HTTPS traffic.
- `external-secrets` for syncing secrets from external secret stores.
- `metrics-server` for Kubernetes resource metrics.
- `kube-prometheus-stack` for observability with Prometheus, Alertmanager, and Grafana.
- `kyverno` for policy as code.

Platform-cluster observability:

- `loki` runs in single-binary mode as a simple central log store for the reference environment.
- `grafana` runs on the platform cluster and includes a preconfigured Loki datasource.

These chart versions are pinned so the example is reproducible. For a production fleet, review each chart's values, security posture, resource requests, and upgrade notes before enabling automated sync.