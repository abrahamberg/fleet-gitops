package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

const SchemaRelativePath = "requests/namespaces/schema.json"

func ResolveRepoRoot(configured string) (string, error) {
	if configured != "" {
		return filepath.Abs(configured)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if fileExists(filepath.Join(current, SchemaRelativePath)) && dirExists(filepath.Join(current, "clusters")) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("could not find fleet-gitops repo root; pass --repo-root")
}

func ResolvePath(repoRoot, value string) (string, error) {
	if filepath.IsAbs(value) {
		return value, nil
	}

	if _, err := os.Stat(value); err == nil {
		return filepath.Abs(value)
	}

	candidate := filepath.Join(repoRoot, value)
	if _, err := os.Stat(candidate); err != nil {
		return "", fmt.Errorf("resolve %s: %w", value, err)
	}
	return filepath.Abs(candidate)
}

func NamespaceRequestFiles(repoRoot string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "requests", "namespaces", "*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no namespace request files found under requests/namespaces")
	}
	sort.Strings(files)
	return files, nil
}

func LoadNamespaceRequest(path string) (NamespaceRequestFile, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return NamespaceRequestFile{}, err
	}

	var document NamespaceRequestFile
	if err := yaml.UnmarshalStrict(contents, &document); err != nil {
		return NamespaceRequestFile{}, fmt.Errorf("parse namespace request %s: %w", path, err)
	}
	return document, nil
}

func LoadClusters(repoRoot string) ([]Cluster, error) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "clusters", "*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no cluster inventory files found under clusters")
	}
	sort.Strings(files)

	clusters := make([]Cluster, 0, len(files))
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var cluster Cluster
		if err := yaml.UnmarshalStrict(contents, &cluster); err != nil {
			return nil, fmt.Errorf("parse cluster inventory %s: %w", file, err)
		}
		if strings.TrimSpace(cluster.Name) == "" || strings.TrimSpace(cluster.Environment) == "" {
			return nil, fmt.Errorf("cluster inventory %s must set name and environment", file)
		}
		clusters = append(clusters, cluster)
	}

	sort.Slice(clusters, func(left, right int) bool {
		return clusters[left].Name < clusters[right].Name
	})
	return clusters, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
