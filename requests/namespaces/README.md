# Tenant Namespace GitOps

Users create, update, and delete tenant namespaces by opening a pull request against this folder. Use one file per namespace.

Request files should follow `schema.json`. In a production repository, validate this schema in CI before merging namespace changes.

Example:

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

The `tenant-namespaces` ApplicationSet reads these files, combines them with `clusters/*.yaml`, and renders the local `charts/tenant-namespace` Helm chart. A namespace is created only on clusters whose `environment` appears in `namespace.env`.

Custom policy settings go under `namespace.policies`. The tenant namespace Helm chart turns those settings into labels, and Kyverno policies match those labels. For example, `podSecurity: restricted` adds `fleet.gitops/policy-pod-security: restricted` to the namespace, then Kyverno adds the matching Kubernetes Pod Security Admission labels.

When a request file is merged, Argo CD creates or updates the namespace in the matching clusters. Removing an environment from `namespace.env` prunes the namespace from that environment. When the request file is removed from `main`, the generated Argo CD Applications are removed and their finalizers prune the namespace resources.

Deleting a namespace deletes the workloads and resources inside it. In production, protect this folder with pull request reviews and branch protection.