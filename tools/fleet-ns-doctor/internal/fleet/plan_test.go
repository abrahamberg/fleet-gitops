package fleet

import "testing"

func TestBuildPlanUsesRequestedEnvironments(t *testing.T) {
	request := NamespaceRequest{
		Name:        "payments",
		Env:         []string{"prod"},
		Owner:       "payments-team",
		RequestedBy: "payments-team",
		Purpose:     "payment workloads",
	}
	clusters := []Cluster{
		{Name: "dev", Environment: "dev", Region: "local"},
		{Name: "prod", Environment: "prod", Region: "local"},
	}

	plan := BuildPlan("requests/namespaces/payments.yaml", request, clusters)

	if len(plan.Entries) != 2 {
		t.Fatalf("expected 2 plan entries, got %d", len(plan.Entries))
	}
	if plan.Entries[0].Included {
		t.Fatal("expected dev to be skipped")
	}
	if !plan.Entries[1].Included {
		t.Fatal("expected prod to be included")
	}
}

func TestBuildExpectedNamespaceMatchesChartLabels(t *testing.T) {
	request := NamespaceRequest{
		Name:        "product-a",
		Env:         []string{"dev", "prod"},
		Owner:       "product-team",
		RequestedBy: "product-team",
		Labels:      []string{"product-group-a"},
		Policies:    NamespacePolicies{PodSecurity: "restricted"},
		Purpose:     "product workloads",
	}
	cluster := Cluster{Name: "dev", Environment: "dev", Region: "local"}

	expected := BuildExpectedNamespace(request, cluster)

	assertMapValue(t, expected.Labels, "app.kubernetes.io/managed-by", "argocd")
	assertMapValue(t, expected.Labels, "fleet.gitops/environment", "dev")
	assertMapValue(t, expected.Labels, "fleet.gitops/owner", "product-team")
	assertMapValue(t, expected.Labels, "fleet.gitops/type", "tenant")
	assertMapValue(t, expected.Labels, "fleet.gitops/product-group-a", "true")
	assertMapValue(t, expected.Labels, "fleet.gitops/policy-pod-security", "restricted")
	assertMapValue(t, expected.Annotations, "fleet.gitops/requested-by", "product-team")
	assertMapValue(t, expected.Annotations, "fleet.gitops/purpose", "product workloads")
	assertMapValue(t, expected.PolicyLabels, "pod-security.kubernetes.io/enforce", "restricted")
}

func TestProdExpectedNamespaceIncludesResourceQuota(t *testing.T) {
	request := NamespaceRequest{Name: "product-a", Env: []string{"prod"}}
	cluster := Cluster{Name: "prod", Environment: "prod"}

	expected := BuildExpectedNamespace(request, cluster)

	found := false
	for _, resource := range expected.GeneratedResources {
		if resource.Kind == "ResourceQuota" && resource.Name == "tenant-resource-quota" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected prod namespace to include tenant-resource-quota")
	}
}

func assertMapValue(t *testing.T, values map[string]string, key, want string) {
	t.Helper()

	got, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	if got != want {
		t.Fatalf("expected %s=%q, got %q", key, want, got)
	}
}
