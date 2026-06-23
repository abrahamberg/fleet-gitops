# Tenant Namespace Chart

This local Helm chart turns a namespace request into a Kubernetes Namespace.

It is intentionally small. The point is not to show advanced Helm. The point is to hide platform implementation details behind the request format in [../../requests/namespaces/](../../requests/namespaces/).

## Files

| File | Purpose |
| --- | --- |
| [Chart.yaml](Chart.yaml) | Chart metadata. |
| [values.yaml](values.yaml) | Example default values for local rendering. |
| [templates/namespace.yaml](templates/namespace.yaml) | The only template; it conditionally renders the Namespace. |

## How It Is Called

[../../applicationsets/tenant-namespaces-appset.yaml](../../applicationsets/tenant-namespaces-appset.yaml) builds Helm values from two sources:

- Cluster metadata from [../../clusters/](../../clusters/).
- Namespace request data from [../../requests/namespaces/](../../requests/namespaces/).

The generated values look like this shape:

```yaml
cluster:
  environment: dev
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

## Conditional Rendering

[templates/namespace.yaml](templates/namespace.yaml) starts with this condition:

```gotemplate
{{- if has .Values.cluster.environment .Values.namespace.env }}
```

That means the chart creates a Namespace only when the current cluster's environment is listed in the request.

Example: if a request includes `dev` and `prod`, both clusters get the namespace. If it includes only `dev`, the prod Application renders no resources and Argo CD can prune the prod namespace.

## Labels And Annotations

The chart writes labels that other platform components can match:

| Output | Used by |
| --- | --- |
| `fleet.gitops/type: tenant` | Kyverno policies that apply to every tenant namespace. |
| `fleet.gitops/environment` | Environment-specific Kyverno policies. |
| `fleet.gitops/owner` | Ownership visibility. |
| `fleet.gitops/policy-pod-security` | Pod Security Admission policy selection. |
| `fleet.gitops/<custom-label>: "true"` | Simple team or grouping labels from the request. |

Annotations record who requested the namespace and why.

## Render Locally

From the repo root:

```bash
helm template tenant-product-a charts/tenant-namespace \
  --set cluster.environment=dev \
  --set namespace.name=product-a \
  --set namespace.owner=product-team \
  --set namespace.requestedBy=product-team \
  --set namespace.purpose="product workloads" \
  --set namespace.env='{dev,prod}' \
  --set namespace.policies.podSecurity=restricted
```

The full end-to-end explanation is in [../../docs/tenant-namespace-flow.md](../../docs/tenant-namespace-flow.md).