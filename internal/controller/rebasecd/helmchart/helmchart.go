package helmchart

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HelmChart struct {
	// Common attributes that are used when HelmChart is embedded
	ValuesFiles []string
	ReleaseName string
	Namespace   string
	Credentials rebasecd.CredentialsHandler

	// Params for getting a chart from a Helm Repository
	repoUrl   string
	version   string
	chartName string

	// Rendered state information about the chart
	tmpDir string

	// Configuration on how to find the chart and where its going
	Settings     *cli.EnvSettings
	ActionConfig *action.Configuration
	ChartPath    string
}

func NewHelmChart(repoURL, chartName, version, releaseName, namespace string, credentials rebasecd.CredentialsHandler, valuesFiles []string) rebasecd.ChartHandler {
	return &HelmChart{
		ReleaseName: releaseName,
		ValuesFiles: valuesFiles,
		Namespace:   namespace,
		Credentials: credentials,

		repoUrl:   repoURL,
		version:   version,
		chartName: chartName,
	}

}

func (h *HelmChart) ChartName() string {
	return h.chartName
}

func (h *HelmChart) ChartVersion() string {
	return h.version
}

func (h *HelmChart) GetChart() error {

	actionConfig, settings, err := utils.NewActionConfigAndSettings(h.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create action config: %w", err)
	}

	chartPathOpts := action.ChartPathOptions{
		RepoURL: h.repoUrl,
		Version: h.version,
	}

	if h.Credentials.HasCredentials() {
		// Set up authentication for Helm repository access
		username := h.Credentials.GetUsername()
		password := h.Credentials.GetPassword()

		chartPathOpts.Username = username
		chartPathOpts.Password = password
	}

	chartPath, err := chartPathOpts.LocateChart(h.chartName, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	// Configure HelmChart with this information
	h.ActionConfig = actionConfig
	h.Settings = settings
	h.ChartPath = chartPath

	return nil
}

func (h *HelmChart) RenderChartObjects() ([]client.Object, error) {

	// Copy chart to a unique temp directory for this reconcile
	tmpDir, err := os.MkdirTemp("", "helm-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	h.tmpDir = tmpDir

	copiedChartPath := filepath.Join(tmpDir, filepath.Base(h.ChartPath))
	if err := utils.CopyFile(h.ChartPath, copiedChartPath); err != nil {
		return nil, fmt.Errorf("failed to copy chart to temp dir: %w", err)
	}

	// Load the chart from the temp directory
	chart, err := loader.Load(copiedChartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Merge multiple value files (relative to temp chart dir)
	valueOpts := &values.Options{ValueFiles: make([]string, len(h.ValuesFiles))}
	for i, vf := range h.ValuesFiles {
		valueOpts.ValueFiles[i] = filepath.Join(tmpDir, vf)
	}

	baseVals, err := valueOpts.MergeValues(getter.All(h.Settings))
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	// Now dryrun so that we can grab out all of the resources in this helm chart

	dryUpgrade := action.NewUpgrade(h.ActionConfig)
	dryUpgrade.Namespace = h.Namespace
	dryUpgrade.Install = true
	dryUpgrade.Atomic = true
	dryUpgrade.Wait = true
	dryUpgrade.DryRun = true

	release, err := dryUpgrade.Run(h.ReleaseName, chart, baseVals)
	if err != nil {
		return nil, fmt.Errorf("failed to install/upgrade chart: %w", err)
	}

	objs, err := utils.DecodeYAMLManifest(release.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML manifest: %w", err)
	}

	return objs, nil

}

func (h *HelmChart) DeployChartObjects(objects []client.Object, ctx context.Context, k8sClient client.Client) error {

	for _, obj := range objects {

		// Create or update
		key := client.ObjectKeyFromObject(obj)
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

		if err := k8sClient.Get(ctx, key, existing); err != nil {
			if apierrors.IsNotFound(err) {
				_ = k8sClient.Create(ctx, obj)
			} else {
				return err
			}
		} else {
			if !utils.UnstructuredObjectsAreEqual(existing, obj) {
				obj.SetResourceVersion(existing.GetResourceVersion())
				_ = k8sClient.Update(ctx, obj)
			}
		}
	}

	//fmt.Printf("Deployed Chart %q:%q in Namespace %q under Release Name %q\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version, release.Namespace, release.Name)

	return nil
}

func (h *HelmChart) Cleanup() {
	os.RemoveAll(h.tmpDir)
}
