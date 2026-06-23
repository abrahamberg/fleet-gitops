# Guestbook Rollout Example

This is a small product workload used to demonstrate progressive delivery across the fleet.

It assumes the Argo Rollouts controller has already been installed by the platform add-ons.

## Files

| File | Purpose |
| --- | --- |
| [kustomization.yaml](kustomization.yaml) | Tells Kustomize to include the Rollout and Service. |
| [rollout.yaml](rollout.yaml) | Defines the Argo Rollouts `Rollout` with a canary strategy. |
| [service.yaml](service.yaml) | Exposes the workload inside the cluster as a ClusterIP Service. |
| [../../applicationsets/guestbook-appset.yaml](../../applicationsets/guestbook-appset.yaml) | Deploys this example to every cluster in the inventory. |
| [../../applicationsets/wave-demo-appset.yaml](../../applicationsets/wave-demo-appset.yaml) | Deploys the same example only to clusters in upgrade wave `0`. |

## How The Rollout Works

[rollout.yaml](rollout.yaml) starts three replicas of `argoproj/rollouts-demo:blue` and uses this canary sequence:

1. Send 20 percent of traffic to the new ReplicaSet.
2. Pause for 30 seconds.
3. Send 50 percent of traffic to the new ReplicaSet.
4. Pause for 1 minute.

This is intentionally small. The point is to show where a product workload lives and how it can use a platform controller installed by [../../applicationsets/platform-helm-addons-appset.yaml](../../applicationsets/platform-helm-addons-appset.yaml).

## Try A Change

Change the image tag in [rollout.yaml](rollout.yaml) from `blue` to another demo color such as `yellow`. After Argo CD syncs, Argo Rollouts creates a new ReplicaSet and steps through the canary strategy.

## Things To Extend

- Add ingress once the ingress controller is available.
- Add analysis templates if you want automated rollout checks.
- Deploy the example only to selected clusters by adding labels to [../../clusters/dev.yaml](../../clusters/dev.yaml) or [../../clusters/prod.yaml](../../clusters/prod.yaml) and using an ApplicationSet selector.