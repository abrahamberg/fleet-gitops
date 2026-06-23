package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/abrahamberg/fleet-gitops/tools/fleet-ns-doctor/internal/fleet"
	"github.com/abrahamberg/fleet-gitops/tools/fleet-ns-doctor/internal/kubecheck"
	"github.com/abrahamberg/fleet-gitops/tools/fleet-ns-doctor/internal/output"
)

type options struct {
	repoRoot string
	out      io.Writer
	errOut   io.Writer
}

func Execute(out, errOut io.Writer) error {
	return NewRootCommand(out, errOut).Execute()
}

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	opts := &options{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:           "fleet-ns-doctor",
		Short:         "Validate, plan, and verify tenant namespace requests for the fleet GitOps repo",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.repoRoot, "repo-root", "", "path to the fleet-gitops repository root")
	cmd.AddCommand(newValidateCommand(opts))
	cmd.AddCommand(newPlanCommand(opts))
	cmd.AddCommand(newVerifyCommand(opts))

	return cmd
}

func newValidateCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [request-file ...]",
		Short: "Validate tenant namespace request files against the repo schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, files, err := resolveRepoAndRequests(opts.repoRoot, args)
			if err != nil {
				return err
			}

			validator, err := fleet.NewValidator(filepath.Join(repoRoot, fleet.SchemaRelativePath))
			if err != nil {
				return err
			}

			rows := make([][]string, 0, len(files))
			failures := 0
			for _, file := range files {
				details := "valid namespace request"
				status := "OK"
				if err := validator.ValidateFile(file); err != nil {
					failures++
					status = "FAIL"
					details = compactError(err)
				}
				rows = append(rows, []string{displayPath(repoRoot, file), status, details})
			}

			if err := output.WriteTable(opts.out, []string{"REQUEST", "STATUS", "DETAILS"}, rows); err != nil {
				return err
			}

			if failures > 0 {
				return fmt.Errorf("%d request file(s) failed validation", failures)
			}
			return nil
		},
	}
}

func newPlanCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "plan [request-file ...]",
		Short: "Preview which clusters will render each tenant namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, files, err := resolveRepoAndRequests(opts.repoRoot, args)
			if err != nil {
				return err
			}

			validator, err := fleet.NewValidator(filepath.Join(repoRoot, fleet.SchemaRelativePath))
			if err != nil {
				return err
			}

			clusters, err := fleet.LoadClusters(repoRoot)
			if err != nil {
				return err
			}

			for index, file := range files {
				if index > 0 {
					fmt.Fprintln(opts.out)
				}

				if err := validator.ValidateFile(file); err != nil {
					return fmt.Errorf("validate %s: %w", displayPath(repoRoot, file), err)
				}

				document, err := fleet.LoadNamespaceRequest(file)
				if err != nil {
					return err
				}

				plan := fleet.BuildPlan(file, document.Namespace, clusters)
				if err := printPlan(opts.out, repoRoot, plan); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func newVerifyCommand(opts *options) *cobra.Command {
	var contextAssignments []string
	var kubeconfig string

	cmd := &cobra.Command{
		Use:   "verify [request-file ...]",
		Short: "Verify live Kubernetes namespaces and generated guardrails",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, files, err := resolveRepoAndRequests(opts.repoRoot, args)
			if err != nil {
				return err
			}

			validator, err := fleet.NewValidator(filepath.Join(repoRoot, fleet.SchemaRelativePath))
			if err != nil {
				return err
			}

			clusters, err := fleet.LoadClusters(repoRoot)
			if err != nil {
				return err
			}

			contextMap, err := kubecheck.ParseContextAssignments(contextAssignments)
			if err != nil {
				return err
			}

			verifier := kubecheck.NewVerifier(kubeconfig, contextMap)
			rows := make([][]string, 0)
			failures := 0

			for _, file := range files {
				if err := validator.ValidateFile(file); err != nil {
					return fmt.Errorf("validate %s: %w", displayPath(repoRoot, file), err)
				}

				document, err := fleet.LoadNamespaceRequest(file)
				if err != nil {
					return err
				}

				plan := fleet.BuildPlan(file, document.Namespace, clusters)
				for _, entry := range plan.IncludedEntries() {
					results := verifier.Verify(context.Background(), entry)
					for _, result := range results {
						if result.Status == kubecheck.StatusFail {
							failures++
						}
						rows = append(rows, []string{
							displayPath(repoRoot, file),
							result.Cluster,
							result.Context,
							result.Check,
							result.Status,
							result.Details,
						})
					}
				}
			}

			if err := output.WriteTable(opts.out, []string{"REQUEST", "CLUSTER", "CONTEXT", "CHECK", "STATUS", "DETAILS"}, rows); err != nil {
				return err
			}

			if failures > 0 {
				return fmt.Errorf("%d verification check(s) failed", failures)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&contextAssignments, "context", nil, "map a cluster or environment to a kubeconfig context, for example dev=kind-dev")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to a kubeconfig file")

	return cmd
}

func resolveRepoAndRequests(repoRoot string, args []string) (string, []string, error) {
	resolvedRoot, err := fleet.ResolveRepoRoot(repoRoot)
	if err != nil {
		return "", nil, err
	}

	if len(args) == 0 {
		files, err := fleet.NamespaceRequestFiles(resolvedRoot)
		return resolvedRoot, files, err
	}

	files := make([]string, 0, len(args))
	for _, arg := range args {
		file, err := fleet.ResolvePath(resolvedRoot, arg)
		if err != nil {
			return "", nil, err
		}
		files = append(files, file)
	}
	return resolvedRoot, files, nil
}

func printPlan(w io.Writer, repoRoot string, plan fleet.Plan) error {
	fmt.Fprintf(w, "Request: %s\n", displayPath(repoRoot, plan.RequestPath))
	fmt.Fprintf(w, "Namespace: %s\n", plan.Request.Name)
	fmt.Fprintf(w, "Owner: %s\n", plan.Request.Owner)
	fmt.Fprintf(w, "Requested environments: %s\n\n", strings.Join(plan.Request.Env, ", "))

	clusterRows := make([][]string, 0, len(plan.Entries))
	for _, entry := range plan.Entries {
		action := "render"
		detail := fmt.Sprintf("Namespace/%s", plan.Request.Name)
		if !entry.Included {
			action = "skip"
			detail = entry.Reason
		}
		clusterRows = append(clusterRows, []string{entry.Cluster.Name, entry.Cluster.Environment, entry.Cluster.Region, action, detail})
	}
	if err := output.WriteTable(w, []string{"CLUSTER", "ENV", "REGION", "ACTION", "DETAILS"}, clusterRows); err != nil {
		return err
	}

	for _, entry := range plan.IncludedEntries() {
		fmt.Fprintf(w, "\nExpected metadata on cluster %s:\n", entry.Cluster.Name)

		metadataRows := make([][]string, 0)
		for _, pair := range output.SortedPairs(entry.Expected.Labels) {
			metadataRows = append(metadataRows, []string{"label", pair.Key, pair.Value})
		}
		for _, pair := range output.SortedPairs(entry.Expected.Annotations) {
			metadataRows = append(metadataRows, []string{"annotation", pair.Key, pair.Value})
		}
		for _, pair := range output.SortedPairs(entry.Expected.PolicyLabels) {
			metadataRows = append(metadataRows, []string{"kyverno label", pair.Key, pair.Value})
		}
		if err := output.WriteTable(w, []string{"TYPE", "KEY", "VALUE"}, metadataRows); err != nil {
			return err
		}

		resourceRows := make([][]string, 0, len(entry.Expected.GeneratedResources))
		for _, resource := range entry.Expected.GeneratedResources {
			resourceRows = append(resourceRows, []string{resource.Kind, resource.Name})
		}
		fmt.Fprintln(w, "Expected generated resources:")
		if err := output.WriteTable(w, []string{"KIND", "NAME"}, resourceRows); err != nil {
			return err
		}
	}

	return nil
}

func displayPath(repoRoot, path string) string {
	relative, err := filepath.Rel(repoRoot, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func compactError(err error) string {
	message := strings.ReplaceAll(err.Error(), "\n", "; ")
	if len(message) <= 120 {
		return message
	}
	return message[:117] + "..."
}
