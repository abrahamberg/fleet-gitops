# Tenant Namespace Flow

This repo uses a request-driven namespace model instead of asking users to write raw Kubernetes `Namespace` manifests.

## Why

The user-facing interface should be small and stable:

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

The platform implementation can then evolve behind that interface. Today it renders a Helm chart and applies Kyverno policies. Later it could add quotas, backups, network profiles, external secret wiring, or RBAC without changing how users request namespaces.

## How It Works

1. A user adds or updates a file in `requests/namespaces/`.
2. The `tenant-namespaces` ApplicationSet builds a matrix of `clusters/*.yaml` and `requests/namespaces/*.yaml`.
3. Each generated Application renders `charts/tenant-namespace` with one cluster and one namespace request as values.
4. The chart creates a `Namespace` only when the cluster environment is listed in `namespace.env`.
5. The chart adds standard labels that Kyverno policies can match.
6. Kyverno generates or mutates namespace-scoped guardrails such as NetworkPolicy, LimitRange, ResourceQuota, and Pod Security labels.

## Delete Behavior

Removing an environment from `namespace.env` makes the generated Application for that cluster render no resources. Because automated prune and `allowEmpty` are enabled, Argo CD prunes the namespace for that environment.

Deleting the request file removes the generated Applications. Their Argo CD finalizers prune the managed namespace resources.

Namespace deletion deletes resources inside the namespace, so production repositories should require review for changes under `requests/namespaces/`.