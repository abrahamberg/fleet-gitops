package kubecheck

import (
	"fmt"
	"strings"
)

func ParseContextAssignments(values []string) (map[string]string, error) {
	assignments := make(map[string]string, len(values))
	for _, value := range values {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid --context value %q, expected cluster=context", value)
		}
		assignments[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return assignments, nil
}
