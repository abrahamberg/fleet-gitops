# Tenant Namespace Requests

This folder is the user-facing namespace catalog. A product team requests a namespace by adding or changing one YAML file here.

Use this model when you want developers to request namespaces without learning the platform's Helm, Kyverno, and ApplicationSet internals.

## Request Format

Use one file per namespace. The example request is [product-a.yaml](product-a.yaml):

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

[schema.json](schema.json) describes the expected shape. In a production repo, validate it in CI before merging.

## Field Reference

| Field | Meaning | Example |
| --- | --- | --- |
| `namespace.name` | Kubernetes namespace name to create. | `product-a` |
| `namespace.env` | Environments where the namespace should exist. | `dev`, `prod` |
| `namespace.owner` | Team or group responsible for the namespace. | `product-team` |
| `namespace.requestedBy` | Person, team, or system that requested it. | `product-team` |
| `namespace.labels` | Extra tenant labels. The chart writes them as `fleet.gitops/<label>: "true"`. | `product-group-a` |
| `namespace.policies.podSecurity` | Requested Kubernetes Pod Security profile. | `baseline` or `restricted` |
| `namespace.purpose` | Short human reason for the namespace. | `product workloads` |

## How It Becomes A Namespace

1. [../../applicationsets/tenant-namespaces-appset.yaml](../../applicationsets/tenant-namespaces-appset.yaml) reads every cluster file and every namespace request file.
2. The ApplicationSet matrix creates one generated Application per cluster and request.
3. Each generated Application renders [../../charts/tenant-namespace](../../charts/tenant-namespace/).
4. [../../charts/tenant-namespace/templates/namespace.yaml](../../charts/tenant-namespace/templates/namespace.yaml) creates a Namespace only when the cluster environment appears in `namespace.env`.
5. The chart adds labels such as `fleet.gitops/type: tenant`, `fleet.gitops/environment`, and `fleet.gitops/policy-pod-security`.
6. Kyverno policies under [../../addons/kyverno-policies/](../../addons/kyverno-policies/) match those labels and generate guardrails.

For the full flow, read [../../docs/tenant-namespace-flow.md](../../docs/tenant-namespace-flow.md).

## Change Behavior

| Change | Result |
| --- | --- |
| Add `dev` to `namespace.env` | The namespace is created on dev clusters. |
| Add `prod` to `namespace.env` | The namespace is created on prod clusters. |
| Remove an environment from `namespace.env` | Argo CD prunes the generated namespace resources from that environment. |
| Delete the request file | The generated Applications are removed, and their finalizers prune the managed namespace resources. |

## Deletion Warning

Deleting a namespace deletes workloads and resources inside that namespace. Protect this folder with pull request reviews, CODEOWNERS, and branch protection in any real team setup.