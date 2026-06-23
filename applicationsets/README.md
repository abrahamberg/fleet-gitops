# Argo CD ApplicationSets

This folder is the control layer of the repo. The root Application points Argo CD here, and these manifests generate the Applications that manage the fleet.

If you understand this folder, the rest of the repo becomes much easier to read.

## Bootstrap Relationship

[../root-app.yaml](../root-app.yaml) is a normal Argo CD Application. It targets this folder:

```yaml
source:
  path: applicationsets
```

After the root app syncs, Argo CD sees the AppProject, Applications, and ApplicationSets in this directory.

## Files

| File | Kind | What it creates |
| --- | --- | --- |
| [fleet-platform-project.yaml](fleet-platform-project.yaml) | AppProject | Permission boundary for generated Applications. |
| [platform-addons-appset.yaml](platform-addons-appset.yaml) | ApplicationSet | Platform namespace Applications, one per cluster. |
| [platform-helm-addons-appset.yaml](platform-helm-addons-appset.yaml) | ApplicationSet | Helm add-on Applications, one per cluster and add-on. |
| [kyverno-policies-appset.yaml](kyverno-policies-appset.yaml) | ApplicationSet | Kyverno policy Applications, one per cluster. |
| [tenant-namespaces-appset.yaml](tenant-namespaces-appset.yaml) | ApplicationSet | Tenant namespace Applications, one per cluster and request file. |
| [observability-agents-appset.yaml](observability-agents-appset.yaml) | ApplicationSet | Grafana Alloy Applications, one per cluster. |
| [platform-observability.yaml](platform-observability.yaml) | Application | Platform Loki and Grafana on the management cluster. |
| [guestbook-appset.yaml](guestbook-appset.yaml) | ApplicationSet | Guestbook example on every cluster. |
| [wave-demo-appset.yaml](wave-demo-appset.yaml) | ApplicationSet | Guestbook example only on clusters in upgrade wave `0`. |

## Generator Patterns

| Pattern | Example | Why it is used |
| --- | --- | --- |
| Git file generator | [platform-addons-appset.yaml](platform-addons-appset.yaml) | Reads every cluster file and creates one Application per cluster. |
| Matrix generator | [platform-helm-addons-appset.yaml](platform-helm-addons-appset.yaml) | Combines clusters with a list of add-ons. |
| Matrix generator with two Git sources | [tenant-namespaces-appset.yaml](tenant-namespaces-appset.yaml) | Combines every cluster with every namespace request. |
| Selector | [wave-demo-appset.yaml](wave-demo-appset.yaml) | Deploys only to inventory files with matching metadata. |
| Plain Application | [platform-observability.yaml](platform-observability.yaml) | Installs central services only once on the platform cluster. |

## Main Flow

1. Cluster files in [../clusters/](../clusters/) provide `name`, `server`, `environment`, `region`, `upgradeWave`, and `lokiEndpoint`.
2. ApplicationSets render those values into Application names, labels, Helm values, source paths, and destinations.
3. Argo CD syncs the generated Applications to the matching clusters.

## Safe Change Points

| Goal | Edit | Watch for |
| --- | --- | --- |
| Add a new cluster-wide add-on | [platform-helm-addons-appset.yaml](platform-helm-addons-appset.yaml) | Pin the chart version and choose a sync wave. |
| Add an environment overlay | [platform-addons-appset.yaml](platform-addons-appset.yaml) or [kyverno-policies-appset.yaml](kyverno-policies-appset.yaml) | The overlay path must exist for each cluster `environment`. |
| Add a namespace request field | [tenant-namespaces-appset.yaml](tenant-namespaces-appset.yaml) | Keep the generated Helm `values` YAML valid after templating. |
| Deploy only to selected clusters | Add or reuse inventory metadata and a selector. | Confirm selector labels match fields from [../clusters/](../clusters/). |

## Sync Waves

Sync waves document the intended order: project first, platform basics next, Kyverno before policies, policies before tenant namespaces, and observability agents after that. For production, combine waves with health checks and review gates instead of relying on ordering alone.