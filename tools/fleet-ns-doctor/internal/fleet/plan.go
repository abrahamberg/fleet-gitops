package fleet

import "fmt"

type Plan struct {
	RequestPath string
	Request     NamespaceRequest
	Entries     []PlanEntry
}

type PlanEntry struct {
	Cluster  Cluster
	Included bool
	Reason   string
	Expected ExpectedNamespace
}

type ExpectedNamespace struct {
	Name               string
	Labels             map[string]string
	Annotations        map[string]string
	PolicyLabels       map[string]string
	GeneratedResources []GeneratedResource
}

type GeneratedResource struct {
	Kind string
	Name string
}

func BuildPlan(requestPath string, request NamespaceRequest, clusters []Cluster) Plan {
	entries := make([]PlanEntry, 0, len(clusters))
	for _, cluster := range clusters {
		included := contains(request.Env, cluster.Environment)
		entry := PlanEntry{
			Cluster:  cluster,
			Included: included,
		}

		if included {
			entry.Expected = BuildExpectedNamespace(request, cluster)
		} else {
			entry.Reason = fmt.Sprintf("environment %q not requested", cluster.Environment)
		}

		entries = append(entries, entry)
	}

	return Plan{RequestPath: requestPath, Request: request, Entries: entries}
}

func (p Plan) IncludedEntries() []PlanEntry {
	entries := make([]PlanEntry, 0, len(p.Entries))
	for _, entry := range p.Entries {
		if entry.Included {
			entries = append(entries, entry)
		}
	}
	return entries
}

func BuildExpectedNamespace(request NamespaceRequest, cluster Cluster) ExpectedNamespace {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "argocd",
		"fleet.gitops/environment":     cluster.Environment,
		"fleet.gitops/owner":           request.Owner,
		"fleet.gitops/type":            "tenant",
	}

	if request.Policies.PodSecurity != "" {
		labels["fleet.gitops/policy-pod-security"] = request.Policies.PodSecurity
	}

	for _, label := range request.Labels {
		labels["fleet.gitops/"+label] = "true"
	}

	annotations := map[string]string{
		"fleet.gitops/requested-by": request.RequestedBy,
		"fleet.gitops/purpose":      request.Purpose,
	}

	policyLabels := map[string]string{}
	if request.Policies.PodSecurity != "" {
		policyLabels["pod-security.kubernetes.io/audit"] = request.Policies.PodSecurity
		policyLabels["pod-security.kubernetes.io/enforce"] = request.Policies.PodSecurity
		policyLabels["pod-security.kubernetes.io/warn"] = request.Policies.PodSecurity
	}

	resources := []GeneratedResource{
		{Kind: "NetworkPolicy", Name: "default-deny-ingress"},
		{Kind: "LimitRange", Name: "tenant-defaults"},
	}

	if cluster.Environment == "prod" {
		resources = append(resources, GeneratedResource{Kind: "ResourceQuota", Name: "tenant-resource-quota"})
	}

	return ExpectedNamespace{
		Name:               request.Name,
		Labels:             labels,
		Annotations:        annotations,
		PolicyLabels:       policyLabels,
		GeneratedResources: resources,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
