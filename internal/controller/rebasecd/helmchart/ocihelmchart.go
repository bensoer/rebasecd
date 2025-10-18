package helmchart

import (
	"fmt"
	"path"
	"strings"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/utils"
	"helm.sh/helm/v3/pkg/action"
)

type OCIHelmChart struct {
	OCIRepositoryURL string
	Version          string
	*HelmChart
}

func NewOCIHelmChart(ociRepositoryURL, version, releaseName, namespace string, valuesFiles []string) rebasecd.ChartHandler {
	protocolStrippedURL := stripOCIPrefix(ociRepositoryURL)
	return &OCIHelmChart{
		OCIRepositoryURL: path.Join("oci://", protocolStrippedURL),
		Version:          version,
		HelmChart: &HelmChart{
			ReleaseName: releaseName,
			ValuesFiles: valuesFiles,
			Namespace:   namespace,
		},
	}
}

func (ohc *OCIHelmChart) GetChart() error {

	actionConfig, settings, err := utils.NewActionConfigAndSettings(ohc.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create action config: %w", err)
	}

	chartPathOpts := action.ChartPathOptions{}

	var finalOciURL string
	if ohc.Version != "" {
		finalOciURL = fmt.Sprintf("%s:%s", ohc.OCIRepositoryURL, ohc.Version)
	} else {
		finalOciURL = ohc.OCIRepositoryURL
	}

	chartPath, err := chartPathOpts.LocateChart(finalOciURL, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	// Configure HelmChart with this information
	ohc.HelmChart.ActionConfig = actionConfig
	ohc.HelmChart.Settings = settings
	ohc.HelmChart.ChartPath = chartPath

	return nil
}

func (ohc *OCIHelmChart) Cleanup() {
	ohc.HelmChart.Cleanup()
}

func stripOCIPrefix(repoURL string) string {
	return strings.TrimPrefix(repoURL, "oci://")
}
