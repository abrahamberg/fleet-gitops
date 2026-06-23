# Guestbook Rollout Example

This example represents a product workload rather than a platform component.

It uses the Argo Rollouts `Rollout` resource to demonstrate canary-style progressive delivery after the `argo-rollouts` platform addon has installed the controller and CRDs on each workload cluster.

The rollout starts with the `argoproj/rollouts-demo:blue` image. Change the image tag to another demo color, such as `yellow`, to watch Argo Rollouts step through the canary strategy.