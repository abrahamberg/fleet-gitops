# fleet-ns-doctor

`fleet-ns-doctor` is a small Go CLI for this GitOps fleet repo. It validates tenant namespace request files, previews which clusters will render a Namespace, and can verify the live Kubernetes objects created by Argo CD and Kyverno.

It is intentionally focused on the tenant namespace flow:

```text
request YAML -> cluster inventory -> Helm-rendered Namespace -> Kyverno-generated guardrails
```

## Commands

Validate all namespace requests:

```bash
go run ./cmd/fleet-ns-doctor validate
```

Validate one request:

```bash
go run ./cmd/fleet-ns-doctor validate ../../requests/namespaces/product-a.yaml
```

Preview the fleet plan:

```bash
go run ./cmd/fleet-ns-doctor plan ../../requests/namespaces/product-a.yaml
```

Verify live clusters using kubeconfig contexts:

```bash
go run ./cmd/fleet-ns-doctor verify ../../requests/namespaces/product-a.yaml \
  --context dev=kind-dev \
  --context prod=kind-prod
```

If `--context` is omitted, the command tries to use the cluster name from `clusters/*.yaml` as the kubeconfig context.

## Build And Test

```bash
make test
make build
```

The binary is created in this folder as `fleet-ns-doctor`.

## What It Checks

`validate` checks request YAML against `requests/namespaces/schema.json`.

`plan` reads `clusters/*.yaml`, applies the same environment filter used by the tenant namespace Helm chart, and prints expected namespace labels, annotations, and Kyverno-generated resources.

`verify` connects to Kubernetes and checks:

- The Namespace exists.
- Chart labels and annotations match the request.
- Pod Security Admission labels were applied when requested.
- The default NetworkPolicy exists.
- The tenant LimitRange exists.
- The prod ResourceQuota exists for prod namespaces.
