# Fleet Architecture

This page explains the whole sample at the level of a platform developer reading the repo for the first time.

## What Runs Where

| Place | Role | Main files |
| --- | --- | --- |
| Platform cluster | Runs Argo CD, plus central Grafana and Loki. | [../root-app.yaml](../root-app.yaml), [platform-observability.yaml](../applicationsets/platform-observability.yaml) |
| Workload clusters | Receive platform add-ons, tenant namespaces, policies, observability agents, and sample workloads. | [../clusters/dev.yaml](../clusters/dev.yaml), [../clusters/prod.yaml](../clusters/prod.yaml) |
| Git repository | Source of truth for cluster inventory, add-on definitions, request files, and examples. | [../applicationsets/](../applicationsets/), [../addons/](../addons/), [../requests/namespaces/](../requests/namespaces/) |

## Sync Flow

1. Apply [../root-app.yaml](../root-app.yaml) to the Argo CD namespace on the platform cluster.
2. The root Application reads every manifest in [../applicationsets/](../applicationsets/).
3. [fleet-platform-project.yaml](../applicationsets/fleet-platform-project.yaml) creates the AppProject used by the generated Applications.
4. ApplicationSets read [../clusters/dev.yaml](../clusters/dev.yaml) and [../clusters/prod.yaml](../clusters/prod.yaml).
5. Generated Applications install platform namespaces, Helm add-ons, Kyverno policies, tenant namespaces, observability agents, and examples.
6. Platform-only Applications install Loki and Grafana on the management cluster.

## Sync Order

Argo CD sync waves keep the demo readable. Lower waves apply first.

| Wave | File | Result |
| --- | --- | --- |
| `-10` | [fleet-platform-project.yaml](../applicationsets/fleet-platform-project.yaml) | Creates the AppProject boundary. |
| `0` | [platform-addons-appset.yaml](../applicationsets/platform-addons-appset.yaml) | Creates platform namespaces with Kustomize. |
| `1` to `6` | [platform-helm-addons-appset.yaml](../applicationsets/platform-helm-addons-appset.yaml) | Installs Helm add-ons such as cert-manager, ingress-nginx, Rollouts, metrics-server, Prometheus stack, and Kyverno. |
| `7` | [kyverno-policies-appset.yaml](../applicationsets/kyverno-policies-appset.yaml) | Applies policy overlays after Kyverno exists. |
| `8` | [tenant-namespaces-appset.yaml](../applicationsets/tenant-namespaces-appset.yaml) | Renders tenant namespace requests through the local Helm chart. |
| `9` | [observability-agents-appset.yaml](../applicationsets/observability-agents-appset.yaml) | Runs Grafana Alloy on workload clusters. |
| `10` and `11` | [platform-observability.yaml](../applicationsets/platform-observability.yaml) | Installs Loki and Grafana on the platform cluster. |

## Main Design Choice

The repo keeps user-facing inputs small and moves platform behavior behind them.

Example: a team edits [../requests/namespaces/product-a.yaml](../requests/namespaces/product-a.yaml). Argo CD renders [../charts/tenant-namespace/templates/namespace.yaml](../charts/tenant-namespace/templates/namespace.yaml), labels the Namespace, and Kyverno policies generate the default guardrails.

That lets the platform change quotas, policies, labels, or add-ons later without asking app teams to learn every platform manifest.

## Where To Go Next

- Read [../applicationsets/README.md](../applicationsets/README.md) to understand the generators.
- Read [kustomize-overlays.md](kustomize-overlays.md) to understand the base/overlay layout.
- Read [tenant-namespace-flow.md](tenant-namespace-flow.md) to follow the most complete end-to-end workflow.
- Read [observability-topology.md](observability-topology.md) for the Grafana, Loki, Prometheus, and Alloy split.