# Kyverno Tenant Policies

This directory contains guardrails for tenant namespaces. Kyverno watches namespace labels and then generates or mutates namespace-scoped resources.

Read this page when you want to understand how a small namespace request turns into default security and resource policy.

## How It Works Here

The tenant namespace chart adds labels to every generated namespace. The policies in this directory match those labels.

Important labels come from [../../charts/tenant-namespace/templates/namespace.yaml](../../charts/tenant-namespace/templates/namespace.yaml):

- `fleet.gitops/type: tenant`
- `fleet.gitops/environment: dev` or `prod`
- `fleet.gitops/policy-pod-security: baseline` or `restricted`, when requested

[../../applicationsets/kyverno-policies-appset.yaml](../../applicationsets/kyverno-policies-appset.yaml) reads each cluster file and deploys `overlays/{{environment}}` to that cluster. A `dev` cluster receives the dev overlay, and a `prod` cluster receives the prod overlay.

## Policy Layout

| Area | File | What it does |
| --- | --- | --- |
| Base Kustomize package | [base/kustomization.yaml](base/kustomization.yaml) | Includes policies shared by every environment. |
| Default network policy | [base/tenant-default-network-policy.yaml](base/tenant-default-network-policy.yaml) | Generates `default-deny-ingress` in every tenant namespace. |
| Pod Security labels | [base/tenant-pod-security-profile.yaml](base/tenant-pod-security-profile.yaml) | Applies Kubernetes Pod Security Admission labels for `baseline` or `restricted`. |
| Dev overlay | [overlays/dev/kustomization.yaml](overlays/dev/kustomization.yaml) | Includes base policies and dev resource defaults. |
| Dev limits | [overlays/dev/tenant-dev-limit-range.yaml](overlays/dev/tenant-dev-limit-range.yaml) | Generates a relaxed `LimitRange` for dev namespaces. |
| Prod overlay | [overlays/prod/kustomization.yaml](overlays/prod/kustomization.yaml) | Includes base policies and stricter prod controls. |
| Prod limits | [overlays/prod/tenant-prod-limit-range.yaml](overlays/prod/tenant-prod-limit-range.yaml) | Generates a stricter `LimitRange` for prod namespaces. |
| Prod quota | [overlays/prod/tenant-prod-resource-quota.yaml](overlays/prod/tenant-prod-resource-quota.yaml) | Generates a `ResourceQuota` for prod namespaces. |

## Request To Policy Example

[../../requests/namespaces/product-a.yaml](../../requests/namespaces/product-a.yaml) requests this policy setting:

```yaml
policies:
	podSecurity: restricted
```

The namespace chart converts it into this label:

```yaml
fleet.gitops/policy-pod-security: restricted
```

Kyverno then matches that label and adds these Kubernetes Pod Security Admission labels:

```yaml
pod-security.kubernetes.io/enforce: restricted
pod-security.kubernetes.io/audit: restricted
pod-security.kubernetes.io/warn: restricted
```

## What To Change

| Goal | Where to edit | Notes |
| --- | --- | --- |
| Add a policy for every tenant namespace | Add a policy under [base/](base/) and include it from [base/kustomization.yaml](base/kustomization.yaml). | Match `fleet.gitops/type: tenant`. |
| Add a dev-only policy | Add a policy under [overlays/dev/](overlays/dev/) and include it from [overlays/dev/kustomization.yaml](overlays/dev/kustomization.yaml). | Match `fleet.gitops/environment: dev` if the policy is namespace-specific. |
| Add a prod-only policy | Add a policy under [overlays/prod/](overlays/prod/) and include it from [overlays/prod/kustomization.yaml](overlays/prod/kustomization.yaml). | Keep prod defaults stricter than dev unless the platform intentionally changes that rule. |
| Add a new request option | Update [../../requests/namespaces/schema.json](../../requests/namespaces/schema.json), [../../applicationsets/tenant-namespaces-appset.yaml](../../applicationsets/tenant-namespaces-appset.yaml), the Helm chart, and the matching policy. | Keep the request file small; hide platform implementation details behind labels. |

## Ordering Note

Kyverno is installed by [../../applicationsets/platform-helm-addons-appset.yaml](../../applicationsets/platform-helm-addons-appset.yaml). These policies are intended to sync after Kyverno and before tenant namespaces, so new namespaces immediately receive generated guardrails.