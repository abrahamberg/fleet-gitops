# Tenant Namespace Flow

This page explains the most important pattern in the repo: a small user request becomes a real Kubernetes namespace with policy attached.

## Mental Model

Developers should not need to write raw Namespace manifests, Kyverno labels, or Helm values. They write a request file. The platform translates that request into implementation details.

```text
request YAML -> ApplicationSet matrix -> Helm values -> Namespace labels -> Kyverno guardrails
```

## User-Facing Request

The request file lives in [../requests/namespaces/product-a.yaml](../requests/namespaces/product-a.yaml):

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

The shape is intentionally small. [../requests/namespaces/schema.json](../requests/namespaces/schema.json) defines the contract.

## End-To-End Flow

| Step | File | What happens |
| --- | --- | --- |
| 1 | [../requests/namespaces/product-a.yaml](../requests/namespaces/product-a.yaml) | A user requests a namespace and target environments. |
| 2 | [../clusters/dev.yaml](../clusters/dev.yaml), [../clusters/prod.yaml](../clusters/prod.yaml) | Cluster files provide `name`, `server`, `environment`, `region`, `upgradeWave`, and `lokiEndpoint`. |
| 3 | [../applicationsets/tenant-namespaces-appset.yaml](../applicationsets/tenant-namespaces-appset.yaml) | The matrix generator pairs every cluster with every namespace request. |
| 4 | [../charts/tenant-namespace/](../charts/tenant-namespace/) | The generated Application renders the local Helm chart. |
| 5 | [../charts/tenant-namespace/templates/namespace.yaml](../charts/tenant-namespace/templates/namespace.yaml) | The chart creates a Namespace only if `cluster.environment` is listed in `namespace.env`. |
| 6 | [../addons/kyverno-policies/README.md](../addons/kyverno-policies/README.md) | Kyverno matches namespace labels and generates NetworkPolicy, LimitRange, ResourceQuota, or Pod Security labels. |

## Why Helm Is Used Here

The chart does one small job: render a namespace from a request. It also hides conditional logic from the user.

This line in [../charts/tenant-namespace/templates/namespace.yaml](../charts/tenant-namespace/templates/namespace.yaml) is the key behavior:

```gotemplate
{{- if has .Values.cluster.environment .Values.namespace.env }}
```

If the cluster is `dev` and the request contains `dev`, the namespace renders. If not, the chart renders nothing for that cluster.

## Labels As The Policy Contract

The chart writes stable labels that policy can match:

| Label | Why it exists |
| --- | --- |
| `fleet.gitops/type: tenant` | Marks the namespace as tenant-owned. |
| `fleet.gitops/environment` | Lets policies distinguish dev from prod. |
| `fleet.gitops/owner` | Records the owning team. |
| `fleet.gitops/policy-pod-security` | Lets Kyverno choose `baseline` or `restricted` Pod Security labels. |
| `fleet.gitops/<custom-label>: "true"` | Carries simple request labels into the namespace. |

This keeps the boundary clear: requests describe intent, labels become the platform contract, and Kyverno owns enforcement.

## Delete Behavior

| User change | Argo CD behavior | Risk |
| --- | --- | --- |
| Remove `dev` or `prod` from `namespace.env` | The generated Application for that cluster renders no resources. With `allowEmpty` and prune enabled, Argo CD prunes the namespace. | Workloads in that namespace are deleted. |
| Delete the request file | The generated Applications disappear. Their Argo CD finalizers prune managed namespace resources. | Workloads in that namespace are deleted in all requested environments. |

Production repos should protect [../requests/namespaces/](../requests/namespaces/) with review rules and make namespace deletion an explicit approval.

## Extension Points

| Need | Where to extend |
| --- | --- |
| Add another request field | Update [../requests/namespaces/schema.json](../requests/namespaces/schema.json) and [../applicationsets/tenant-namespaces-appset.yaml](../applicationsets/tenant-namespaces-appset.yaml). |
| Add a namespace label or annotation | Update [../charts/tenant-namespace/templates/namespace.yaml](../charts/tenant-namespace/templates/namespace.yaml). |
| Add a policy that reacts to a request | Add a Kyverno policy under [../addons/kyverno-policies/](../addons/kyverno-policies/) and match the chart label. |
| Add RBAC, quotas, or secret wiring | Keep the request stable, then implement behind the chart and policy layer. |