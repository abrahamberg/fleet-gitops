# Kyverno Tenant Policies

Kyverno policies in this directory apply guardrails to namespaces created from `requests/namespaces/*.yaml`.

The `kyverno-policies` ApplicationSet deploys `overlays/{{environment}}` to each cluster. Policies match tenant namespaces by labels added by the tenant namespace Helm chart:

- `fleet.gitops/type: tenant`
- `fleet.gitops/environment: dev` or `prod`

Included policies:

- Base: generate a `default-deny-ingress` NetworkPolicy in every tenant namespace.
- Base: apply Pod Security Admission labels when a namespace request sets `namespace.policies.podSecurity` to `baseline` or `restricted`.
- Dev: generate a relaxed `LimitRange` for development workloads.
- Prod: generate a stricter `LimitRange` and `ResourceQuota` for production workloads.

The policy Application runs after Kyverno is installed and before tenant namespace Applications sync. That lets Kyverno generate namespace-scoped resources when the tenant namespace is created.