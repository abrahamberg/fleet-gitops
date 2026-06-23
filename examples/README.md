# Workload Examples

This directory contains sample product workloads. They are deployed by Argo CD, but they are not platform add-ons.

Use this directory when you want to see how an application team might consume the fleet platform.

## How Examples Fit The Fleet

ApplicationSets still live in [../applicationsets/](../applicationsets/) because Argo CD reads that folder from the root app. The workload manifests live here so platform code and product code stay visually separate.

| Example | What it shows | Fleet generator |
| --- | --- | --- |
| [guestbook-rollout/](guestbook-rollout/) | A simple HTTP workload using the Argo Rollouts `Rollout` resource. | [../applicationsets/guestbook-appset.yaml](../applicationsets/guestbook-appset.yaml) |
| Wave demo | The same guestbook workload deployed only to clusters with `upgradeWave: "0"`. | [../applicationsets/wave-demo-appset.yaml](../applicationsets/wave-demo-appset.yaml) |

## What To Notice

- [../applicationsets/guestbook-appset.yaml](../applicationsets/guestbook-appset.yaml) reads every file in [../clusters/](../clusters/) and deploys the example to each cluster.
- [../applicationsets/wave-demo-appset.yaml](../applicationsets/wave-demo-appset.yaml) uses a selector so only matching cluster inventory files receive the workload.
- The workload depends on the Argo Rollouts controller installed by [../applicationsets/platform-helm-addons-appset.yaml](../applicationsets/platform-helm-addons-appset.yaml).

## Add Another Example

1. Create a new folder under this directory with Kubernetes manifests and a `kustomization.yaml`.
2. Add an ApplicationSet under [../applicationsets/](../applicationsets/) that points to the new folder.
3. Decide whether the workload should deploy to every cluster or only clusters matching labels such as `environment`, `region`, or `upgradeWave`.