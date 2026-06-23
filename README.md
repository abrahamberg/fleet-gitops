# Fleet GitOps Reference Architecture

This repository is an educational multi-cluster GitOps reference built around Argo CD ApplicationSets.

It shows how to manage:

- Cluster inventory from `clusters/*.yaml`.
- Platform namespaces and platform add-ons across every cluster.
- Central platform-cluster observability with Grafana and Loki.
- Per-cluster observability agents for workload log collection.
- Tenant namespace requests from a simple user-facing YAML format.
- Kyverno policies that react to namespace request settings.
- Product examples generated from the same cluster inventory.

## Repository Layout

- `root-app.yaml` bootstraps the app-of-apps entry point.
- `applicationsets/` contains Argo CD AppProjects, Applications, and ApplicationSets.
- `clusters/` defines target cluster metadata such as `name`, `server`, `environment`, `region`, `upgradeWave`, and telemetry endpoints.
- `addons/` contains Kustomize add-ons and policy overlays.
- `charts/` contains local Helm charts used by Argo CD.
- `examples/` contains product or workload examples, separate from the platform layer.
- `requests/namespaces/` is the user-facing tenant namespace catalog.
- `docs/` contains architecture notes for the reference patterns.

## Sync Flow

1. Apply `root-app.yaml` to the Argo CD management cluster.
2. The root app syncs everything in `applicationsets/`.
3. `fleet-platform-project` creates the AppProject boundary.
4. Platform namespace scaffolding syncs to each cluster.
5. Helm add-ons install shared services such as cert-manager, ingress-nginx, Argo Rollouts, external-secrets, metrics-server, kube-prometheus-stack, and Kyverno on each workload cluster. The fleet kube-prometheus-stack disables Grafana.
6. Kyverno policy overlays sync after Kyverno is installed.
7. Tenant namespace Applications render `charts/tenant-namespace` from files in `requests/namespaces/*.yaml`.
8. Observability agents run on workload clusters and ship pod logs to platform Loki.
9. Product examples under `examples/` can be deployed by ApplicationSets without treating those workloads as platform components.

## Product Examples

`examples/guestbook-rollout/` is a simple product workload that uses Argo Rollouts for canary-style progressive delivery. The ApplicationSets for the examples still live in `applicationsets/` because Argo CD reads them as part of the root app, but the workload manifests live under `examples/`.

## Observability Topology

Grafana and Loki are installed only on the platform cluster by `applicationsets/platform-observability.yaml`. In this reference, the platform cluster is the Argo CD management cluster, addressed by `https://kubernetes.default.svc`.

Workload clusters receive per-cluster observability components from `platform-helm-addons`, but they do not run Grafana. The `kube-prometheus-stack` value `grafana.enabled: false` prevents duplicate Grafana instances in dev and prod.

## Tenant Namespace Requests

Users request a namespace by adding one file under `requests/namespaces/`:

```yaml
namespace:
  name: product-a
  env:
    - dev
    - prod
  owner: product-team
  requestedBy: product-team
  labels:
    - product-group-a
  policies:
    podSecurity: restricted
  purpose: product workloads
```

The tenant namespace ApplicationSet combines each request with each cluster. The local Helm chart only renders a namespace when the cluster environment appears in `namespace.env`.

## Policy Model

The namespace chart adds standard labels such as:

- `fleet.gitops/type: tenant`
- `fleet.gitops/environment: dev` or `prod`
- `fleet.gitops/policy-pod-security: restricted`

Kyverno policies match those labels and generate or mutate namespace-scoped resources. Dev and prod use different overlays under `addons/kyverno-policies/overlays/`.

## Production Notes

This repo is intentionally readable for learning. Before using the pattern in production, add branch protection and CODEOWNERS for `requests/namespaces/`, validate request files in CI, review chart values for each platform add-on, and make namespace deletion an explicit reviewed operation.

See `docs/tenant-namespace-flow.md` for the namespace request design rationale.

See `docs/telemetry-collection.md` for the logs and metrics collection model.