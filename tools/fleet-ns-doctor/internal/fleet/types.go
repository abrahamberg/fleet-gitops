package fleet

type NamespaceRequestFile struct {
	Namespace NamespaceRequest `json:"namespace" yaml:"namespace"`
}

type NamespaceRequest struct {
	Name        string            `json:"name" yaml:"name"`
	Env         []string          `json:"env" yaml:"env"`
	Owner       string            `json:"owner" yaml:"owner"`
	RequestedBy string            `json:"requestedBy" yaml:"requestedBy"`
	Labels      []string          `json:"labels,omitempty" yaml:"labels,omitempty"`
	Policies    NamespacePolicies `json:"policies,omitempty" yaml:"policies,omitempty"`
	Purpose     string            `json:"purpose" yaml:"purpose"`
}

type NamespacePolicies struct {
	PodSecurity string `json:"podSecurity,omitempty" yaml:"podSecurity,omitempty"`
}

type Cluster struct {
	Name         string `json:"name" yaml:"name"`
	Server       string `json:"server" yaml:"server"`
	Environment  string `json:"environment" yaml:"environment"`
	UpgradeWave  string `json:"upgradeWave" yaml:"upgradeWave"`
	Region       string `json:"region" yaml:"region"`
	LokiEndpoint string `json:"lokiEndpoint" yaml:"lokiEndpoint"`
}
