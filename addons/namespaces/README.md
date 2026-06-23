# Namespace Requests

Platform namespaces are managed in this add-on. User and product namespaces are managed separately through `requests/namespaces`.

The `tenant-namespaces` ApplicationSet reads request files from `requests/namespaces/*.yaml` and renders `charts/tenant-namespace` for each cluster. When a user namespace request is merged to `main`, Argo CD creates or updates it on the environments listed in `namespace.env`. When the request file is deleted from `main`, Argo CD prunes the generated namespace resources.

Use this pattern:

- Add `dev` to `namespace.env` when the team needs dev access.
- Add `prod` to `namespace.env` when the team needs prod access.
- Keep shared platform namespaces in `base/platform-namespaces.yaml`.

Example: `requests/namespaces/product-a.yaml` declares the `product-a` namespace for dev and prod. To make it dev-only, remove `prod` from `namespace.env`.

For real teams, protect the requests folder with pull request reviews because namespace deletion also deletes the resources inside the namespace.