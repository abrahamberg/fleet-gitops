package kubecheck

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/abrahamberg/fleet-gitops/tools/fleet-ns-doctor/internal/fleet"
)

const (
	StatusPass = "PASS"
	StatusFail = "FAIL"
)

type Result struct {
	Cluster string
	Context string
	Check   string
	Status  string
	Details string
}

type Verifier struct {
	kubeconfig string
	contexts   map[string]string
}

func NewVerifier(kubeconfig string, contexts map[string]string) Verifier {
	return Verifier{kubeconfig: kubeconfig, contexts: contexts}
}

func (v Verifier) Verify(ctx context.Context, entry fleet.PlanEntry) []Result {
	contextName := v.contextFor(entry.Cluster)
	clientset, err := v.clientset(contextName)
	if err != nil {
		return []Result{v.result(entry.Cluster, contextName, "kubeconfig", StatusFail, err.Error())}
	}

	expected := entry.Expected
	results := make([]Result, 0)

	namespace, err := clientset.CoreV1().Namespaces().Get(ctx, expected.Name, metav1.GetOptions{})
	if err != nil {
		return append(results, v.result(entry.Cluster, contextName, "Namespace/"+expected.Name, StatusFail, err.Error()))
	}
	results = append(results, v.result(entry.Cluster, contextName, "Namespace/"+expected.Name, StatusPass, "namespace exists"))

	results = append(results, v.checkMap(entry.Cluster, contextName, "label", namespace.Labels, expected.Labels)...)
	results = append(results, v.checkMap(entry.Cluster, contextName, "annotation", namespace.Annotations, expected.Annotations)...)
	results = append(results, v.checkMap(entry.Cluster, contextName, "kyverno label", namespace.Labels, expected.PolicyLabels)...)

	for _, resource := range expected.GeneratedResources {
		results = append(results, v.checkGeneratedResource(ctx, clientset, entry.Cluster, contextName, expected.Name, resource))
	}

	return results
}

func (v Verifier) contextFor(cluster fleet.Cluster) string {
	if contextName, ok := v.contexts[cluster.Name]; ok {
		return contextName
	}
	if contextName, ok := v.contexts[cluster.Environment]; ok {
		return contextName
	}
	return cluster.Name
}

func (v Verifier) clientset(contextName string) (*kubernetes.Clientset, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if v.kubeconfig != "" {
		rules.ExplicitPath = v.kubeconfig
	}

	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (v Verifier) checkMap(cluster fleet.Cluster, contextName, prefix string, actual, expected map[string]string) []Result {
	if len(expected) == 0 {
		return nil
	}

	keys := make([]string, 0, len(expected))
	for key := range expected {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	results := make([]Result, 0, len(keys))
	for _, key := range keys {
		want := expected[key]
		got, ok := actual[key]
		check := fmt.Sprintf("%s %s", prefix, key)
		if !ok {
			results = append(results, v.result(cluster, contextName, check, StatusFail, "missing"))
			continue
		}
		if got != want {
			results = append(results, v.result(cluster, contextName, check, StatusFail, fmt.Sprintf("got %q, want %q", got, want)))
			continue
		}
		results = append(results, v.result(cluster, contextName, check, StatusPass, got))
	}
	return results
}

func (v Verifier) checkGeneratedResource(ctx context.Context, clientset *kubernetes.Clientset, cluster fleet.Cluster, contextName, namespace string, resource fleet.GeneratedResource) Result {
	check := fmt.Sprintf("%s/%s", resource.Kind, resource.Name)

	var err error
	switch resource.Kind {
	case "NetworkPolicy":
		_, err = clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	case "LimitRange":
		_, err = clientset.CoreV1().LimitRanges(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	case "ResourceQuota":
		_, err = clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	default:
		return v.result(cluster, contextName, check, StatusFail, "unsupported resource kind")
	}

	if err != nil {
		return v.result(cluster, contextName, check, StatusFail, err.Error())
	}
	return v.result(cluster, contextName, check, StatusPass, "exists")
}

func (v Verifier) result(cluster fleet.Cluster, contextName, check, status, details string) Result {
	return Result{Cluster: cluster.Name, Context: contextName, Check: check, Status: status, Details: details}
}
