# Fleet GitOps Reference

This repository is a compact Kubernetes fleet example for developers who already know the basics: clusters, Namespaces, Pods, Services, Helm, Kustomize, and the idea of an Argo CD Application.

The goal is to show how one Git repo can describe a small multi-cluster platform without making every team write raw platform manifests.

## What Fleet Means Here

A fleet is a set of Kubernetes clusters managed from one control point. In this example:

- The platform cluster runs Argo CD, Grafana, and Loki.
- The workload clusters are `dev` and `prod`.
- Cluster metadata lives in [clusters/](clusters/).
- Argo CD ApplicationSets in [applicationsets/](applicationsets/) generate one or more Argo CD Applications per cluster.
- Platform add-ons, policies, tenant namespaces, observability agents, and sample workloads are all driven from the same cluster inventory.

## Architecture At A Glance

```text
Git repository
	|
	| root-app.yaml
	v
Argo CD on the platform cluster
	|
	| reads applicationsets/
	v
Generated Argo CD Applications
	|-- dev cluster: namespaces, Helm add-ons, Kyverno policies, tenants, examples
	|-- prod cluster: namespaces, Helm add-ons, Kyverno policies, tenants, examples
	|-- platform cluster: Grafana and Loki
```

The important idea: add or change fleet inputs in Git, then Argo CD generates the per-cluster Applications.

## Fast Reading Path

Read these in order if you want the whole repo quickly:

1. [docs/README.md](docs/README.md) for the documentation map.
2. [docs/fleet-architecture.md](docs/fleet-architecture.md) for the sync flow and platform/workload split.
3. [clusters/README.md](clusters/README.md) for the fleet inventory fields.
4. [applicationsets/README.md](applicationsets/README.md) for how Applications are generated.
5. [addons/README.md](addons/README.md) for platform add-ons.
6. [requests/namespaces/README.md](requests/namespaces/README.md) for the user-facing namespace request format.
7. [docs/kustomize-overlays.md](docs/kustomize-overlays.md) for how the `kustomization.yaml` files work in this sample.

## Key Files

| Area | Start here | What it teaches |
| --- | --- | --- |
| Bootstrap | [root-app.yaml](root-app.yaml) | App-of-apps entry point applied to the Argo CD cluster. |
| Fleet inventory | [clusters/dev.yaml](clusters/dev.yaml), [clusters/prod.yaml](clusters/prod.yaml) | Cluster name, API server, environment, region, upgrade wave, and Loki endpoint. |
| Application generation | [applicationsets/README.md](applicationsets/README.md) | Git, list, matrix, and selector generators. |
| Platform add-ons | [addons/README.md](addons/README.md) | Namespaces, Helm add-ons, Kyverno, and observability agents. |
| Tenant namespace requests | [requests/namespaces/product-a.yaml](requests/namespaces/product-a.yaml) | Small YAML interface for teams requesting namespaces. |
| Tenant namespace chart | [charts/tenant-namespace/README.md](charts/tenant-namespace/README.md) | How request YAML becomes a Kubernetes Namespace. |
| Policy model | [addons/kyverno-policies/README.md](addons/kyverno-policies/README.md) | How Kyverno reacts to namespace labels. |
| Workload example | [examples/guestbook-rollout/README.md](examples/guestbook-rollout/README.md) | A sample app deployed across the fleet. |

## Run A Local Multi-Cluster Demo With Kind

For the cleanest local run, use a fork of this repo. Argo CD syncs from Git, so local-only edits are not enough unless your Argo CD instance can read them from a Git remote.

Prerequisites: Docker, `kind`, `kubectl`, `argocd`, `git`, and `perl`.

```bash
# 1. Create one platform cluster and two workload clusters.
kind create cluster --name platform
kind create cluster --name dev
kind create cluster --name prod

# 2. Install Argo CD on the platform cluster.
kubectl --context kind-platform create namespace argocd
kubectl --context kind-platform apply -n argocd \
	-f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl --context kind-platform -n argocd rollout status deploy/argocd-server
kubectl --context kind-platform -n argocd rollout status deploy/argocd-repo-server
kubectl --context kind-platform -n argocd rollout status deploy/argocd-applicationset-controller
kubectl --context kind-platform -n argocd rollout status statefulset/argocd-application-controller

# 3. Point the repo at your fork, then push the change.
export REPO_URL="https://github.com/YOUR_ORG/fleet-gitops.git"
perl -pi -e "s#https://github.com/abrahamberg/fleet-gitops.git#$REPO_URL#g" root-app.yaml applicationsets/*.yaml

# 4. Point the cluster inventory at the local kind API servers.
DEV_API="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' dev-control-plane)"
PROD_API="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' prod-control-plane)"
sed -i "s#^server:.*#server: https://${DEV_API}:6443#" clusters/dev.yaml
sed -i "s#^server:.*#server: https://${PROD_API}:6443#" clusters/prod.yaml
git add root-app.yaml applicationsets clusters/dev.yaml clusters/prod.yaml
git commit -m "Configure local kind fleet"
git push
```

In another terminal, port-forward Argo CD and register the workload clusters with the same server URLs used in [clusters/dev.yaml](clusters/dev.yaml) and [clusters/prod.yaml](clusters/prod.yaml):

```bash
kubectl --context kind-platform -n argocd port-forward svc/argocd-server 8080:443
```

```bash
ARGOCD_PASSWORD="$(kubectl --context kind-platform -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)"
argocd login localhost:8080 --username admin --password "$ARGOCD_PASSWORD" --insecure

kind get kubeconfig --name dev > /tmp/kind-dev.kubeconfig
kind get kubeconfig --name prod > /tmp/kind-prod.kubeconfig
DEV_API="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' dev-control-plane)"
PROD_API="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' prod-control-plane)"
kubectl --kubeconfig /tmp/kind-dev.kubeconfig config set-cluster kind-dev --server="https://${DEV_API}:6443"
kubectl --kubeconfig /tmp/kind-prod.kubeconfig config set-cluster kind-prod --server="https://${PROD_API}:6443"
KUBECONFIG=/tmp/kind-dev.kubeconfig argocd cluster add kind-dev --name dev --yes
KUBECONFIG=/tmp/kind-prod.kubeconfig argocd cluster add kind-prod --name prod --yes

kubectl --context kind-platform apply -f root-app.yaml
kubectl --context kind-platform -n argocd get applicationsets
kubectl --context kind-platform -n argocd get applications
```

Detailed setup notes and cleanup commands are in [docs/local-kind-setup.md](docs/local-kind-setup.md).

## Safety Notes

- The sample uses automated sync and prune so the behavior is visible. Deleting a namespace request can delete the namespace and the resources inside it.
- Local kind API server IPs can change when clusters are recreated. Update [clusters/dev.yaml](clusters/dev.yaml), [clusters/prod.yaml](clusters/prod.yaml), and the Argo CD cluster registrations together.
- The observability endpoint in each cluster file is intentionally simple. Real clusters need a reachable Loki endpoint across cluster boundaries.
- Before using this pattern for real teams, add branch protection, CODEOWNERS, request schema validation, chart value review, and a reviewed namespace deletion process.

## Common Changes

| Change | File to edit | What happens |
| --- | --- | --- |
| Add a cluster | Add a file under [clusters/](clusters/) and register that cluster in Argo CD. | ApplicationSets generate the platform add-ons, policies, agents, and examples for the new cluster. |
| Add a tenant namespace | Add a request under [requests/namespaces/](requests/namespaces/). | The tenant namespace chart creates the namespace only in the requested environments. |
| Add a platform namespace | Edit [addons/namespaces/base/platform-namespaces.yaml](addons/namespaces/base/platform-namespaces.yaml). | The namespace appears on every cluster through the environment overlays. |
| Add a Helm add-on | Add an element in [applicationsets/platform-helm-addons-appset.yaml](applicationsets/platform-helm-addons-appset.yaml). | The matrix generator installs that chart once per cluster. |
| Add tenant policy | Add a Kyverno policy and include it from a Kustomize overlay under [addons/kyverno-policies/](addons/kyverno-policies/). | Kyverno matches tenant namespace labels and generates or mutates resources. |
| Add a workload example | Add manifests under [examples/](examples/) and an ApplicationSet under [applicationsets/](applicationsets/). | The workload can deploy across the same cluster inventory. |
