# Platform Namespaces

This add-on creates namespaces used by platform components. It is separate from tenant namespace requests.

Use this directory for namespaces such as `observability`, `ingress-system`, `cert-manager`, and `kyverno`. Use [../../requests/namespaces/README.md](../../requests/namespaces/README.md) for user or product namespaces.

## How It Works Here

[../../applicationsets/platform-addons-appset.yaml](../../applicationsets/platform-addons-appset.yaml) reads each file in [../../clusters/](../../clusters/) and points Argo CD at `addons/namespaces/overlays/{{environment}}`.

Kustomize then builds the right overlay for each cluster:

| File | Role |
| --- | --- |
| [base/kustomization.yaml](base/kustomization.yaml) | Includes the shared platform namespace manifest. |
| [base/platform-namespaces.yaml](base/platform-namespaces.yaml) | Lists platform-owned namespaces. |
| [overlays/dev/kustomization.yaml](overlays/dev/kustomization.yaml) | Includes the base and adds dev labels. |
| [overlays/prod/kustomization.yaml](overlays/prod/kustomization.yaml) | Includes the base and adds prod labels. |

Kustomize is helpful here because most namespace YAML is identical, while labels can still differ by environment.

## Platform Vs Tenant Namespaces

| Namespace type | Owner | Managed by | Example |
| --- | --- | --- | --- |
| Platform namespace | Platform team | This add-on | `observability`, `kyverno`, `cert-manager` |
| Tenant namespace | Product or application team | Namespace request catalog and Helm chart | `product-a` |

Keep the two models separate. Platform namespaces are infrastructure. Tenant namespaces are user-facing requests and have deletion risks for workloads inside them.

## What To Change

| Goal | Edit | Result |
| --- | --- | --- |
| Add a platform namespace everywhere | [base/platform-namespaces.yaml](base/platform-namespaces.yaml) | Every environment overlay includes it. |
| Add environment-specific labels | [overlays/dev/kustomization.yaml](overlays/dev/kustomization.yaml) or [overlays/prod/kustomization.yaml](overlays/prod/kustomization.yaml) | Kustomize applies the labels while building that overlay. |
| Add an environment-specific namespace | Add a manifest in the target overlay and include it from that overlay's `kustomization.yaml`. | Only clusters for that environment receive it. |

## Deletion Warning

Deleting a namespace deletes resources inside it. Treat platform namespace removal as a reviewed infrastructure change, even in this reference repo.