# Cluster Inventory

This folder is the fleet inventory. Each YAML file describes one target Kubernetes cluster that Argo CD can deploy to.

ApplicationSets read these files with the Git file generator. That makes Git, not Argo CD cluster secrets, the readable source of truth for which clusters are in the fleet and what metadata they carry.

## Files

| File | Environment | Purpose |
| --- | --- | --- |
| [dev.yaml](dev.yaml) | `dev` | Local development workload cluster. |
| [prod.yaml](prod.yaml) | `prod` | Local production-like workload cluster. |

## Field Reference

| Field | Used by | Meaning |
| --- | --- | --- |
| `name` | Application names, labels, Alloy log labels | Stable cluster name such as `dev` or `prod`. |
| `server` | Argo CD `destination.server` | Kubernetes API server URL. This must match an Argo CD registered cluster. |
| `environment` | Overlay paths, labels, tenant namespace filtering | Environment name used by Kustomize overlays and namespace requests. |
| `upgradeWave` | ApplicationSet selectors | Simple rollout grouping for wave-based deployment examples. |
| `region` | Application labels | Region metadata for filtering and visibility. |
| `lokiEndpoint` | Grafana Alloy values | Loki push endpoint used by workload-cluster log agents. |

## How The Inventory Is Used

| Consumer | File | Result |
| --- | --- | --- |
| Platform namespaces | [../applicationsets/platform-addons-appset.yaml](../applicationsets/platform-addons-appset.yaml) | Points each cluster at `addons/namespaces/overlays/{{environment}}`. |
| Helm add-ons | [../applicationsets/platform-helm-addons-appset.yaml](../applicationsets/platform-helm-addons-appset.yaml) | Installs each listed Helm chart once per cluster. |
| Kyverno policies | [../applicationsets/kyverno-policies-appset.yaml](../applicationsets/kyverno-policies-appset.yaml) | Applies the policy overlay matching `environment`. |
| Tenant namespaces | [../applicationsets/tenant-namespaces-appset.yaml](../applicationsets/tenant-namespaces-appset.yaml) | Combines every cluster with every namespace request. |
| Observability agents | [../applicationsets/observability-agents-appset.yaml](../applicationsets/observability-agents-appset.yaml) | Configures Alloy with cluster labels and `lokiEndpoint`. |
| Workload examples | [../applicationsets/guestbook-appset.yaml](../applicationsets/guestbook-appset.yaml) | Deploys examples to each cluster. |

## Add A Cluster

1. Copy [dev.yaml](dev.yaml) or [prod.yaml](prod.yaml) to a new file.
2. Set a unique `name`.
3. Set `server` to the Kubernetes API server URL Argo CD will use.
4. Set `environment` to an overlay that exists in [../addons/namespaces/overlays/](../addons/namespaces/overlays/) and [../addons/kyverno-policies/overlays/](../addons/kyverno-policies/overlays/).
5. Register the cluster in Argo CD with the same server URL.
6. Commit and push the inventory file.

The server URL matters: if the cluster file says one URL and Argo CD registered another, generated Applications will target an unknown cluster.