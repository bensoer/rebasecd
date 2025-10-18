package helmchart

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

type HelmChart struct {
	// Common attributes that are used when HelmChart is embedded
	ValuesFiles []string
	ReleaseName string
	Namespace   string

	// Params for getting a chart from a Helm Repository
	repoUrl   string
	version   string
	chartName string

	// Rendered state information about the chart
	renderedChart  *chart.Chart
	renderedValues map[string]interface{}
	tmpDir         string

	// Configuration on how to find the chart and where its going
	Settings     *cli.EnvSettings
	ActionConfig *action.Configuration
	ChartPath    string
}

func NewHelmChart(repoURL, chartName, version, releaseName, namespace string, valuesFiles []string) rebasecd.ChartHandler {
	return &HelmChart{
		ReleaseName: releaseName,
		ValuesFiles: valuesFiles,
		Namespace:   namespace,

		repoUrl:   repoURL,
		version:   version,
		chartName: chartName,
	}

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

func (h *HelmChart) RenderChart() error {

	// Copy chart to a unique temp directory for this reconcile
	tmpDir, err := os.MkdirTemp("", "helm-chart-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	h.tmpDir = tmpDir

	copiedChartPath := filepath.Join(tmpDir, filepath.Base(h.ChartPath))
	if err := h.copyFile(h.ChartPath, copiedChartPath); err != nil {
		return fmt.Errorf("failed to copy chart to temp dir: %w", err)
	}

	// Load the chart from the temp directory
	chart, err := loader.Load(copiedChartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	// Merge multiple value files (relative to temp chart dir)
	valueOpts := &values.Options{ValueFiles: make([]string, len(h.ValuesFiles))}
	for i, vf := range h.ValuesFiles {
		valueOpts.ValueFiles[i] = filepath.Join(tmpDir, vf)
	}

	baseVals, err := valueOpts.MergeValues(getter.All(h.Settings))
	if err != nil {
		return fmt.Errorf("failed to merge values: %w", err)
	}

	h.renderedChart = chart
	h.renderedValues = baseVals

	return nil

}

func (h *HelmChart) DeployChart() error {

	// 7. Install or upgrade the release
	upg := action.NewUpgrade(h.ActionConfig)
	upg.Namespace = h.Namespace
	upg.Install = true
	upg.Atomic = true
	upg.Wait = true

	release, err := upg.Run(h.ReleaseName, h.renderedChart, h.renderedValues)
	if err != nil {
		return fmt.Errorf("failed to install/upgrade chart: %w", err)
	}

	fmt.Printf("Deployed Chart %q:%q in Namespace %q under Release Name %q\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version, release.Namespace, release.Name)

	return nil
}

func (h *HelmChart) Cleanup() error {
	os.RemoveAll(h.tmpDir)
	h.renderedChart = nil
	h.renderedValues = nil

	return nil
}

func (h *HelmChart) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
