package helmchart

import (
	"fmt"
	"path"
	"strings"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"
)

type OCIHelmChart struct {
	OCIRepositoryURL string
	Version          string
	*HelmChart
}

func NewOCIHelmChart(ociRepositoryURL, version, releaseName, namespace string, credentials rebasecd.CredentialsHandler, valuesFiles []string) rebasecd.ChartHandler {
	protocolStrippedURL := stripOCIPrefix(ociRepositoryURL)
	return &OCIHelmChart{
		OCIRepositoryURL: path.Join("oci://", protocolStrippedURL),
		Version:          version,
		HelmChart: &HelmChart{
			ReleaseName: releaseName,
			ValuesFiles: valuesFiles,
			Namespace:   namespace,
			Credentials: credentials,
		},
	}
}

func (ohc *OCIHelmChart) ChartName() string {
	return ohc.extractChartName(ohc.OCIRepositoryURL)
}

func (ohc *OCIHelmChart) ChartVersion() string {
	return ohc.Version
}

func (ohc *OCIHelmChart) GetChart() error {

	if ohc.Credentials.HasCredentials() {
		// Set up authentication for Helm repository access
		username := ohc.Credentials.GetUsername()
		password := ohc.Credentials.GetPassword()

		client, err := registry.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create registry client: %w", err)
		}

		if err := client.Login(
			ohc.OCIRepositoryURL,
			registry.LoginOptBasicAuth(username, password),
		); err != nil {
			return fmt.Errorf("failed to login to registry: %w", err)
		}

	}

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

func (ohc *OCIHelmChart) extractChartName(ociURL string) string {
	// Remove oci:// prefix if present
	ociURL = strings.TrimPrefix(ociURL, "oci://")

	// Remove tag or digest if present (:1.2.3 or @sha256:...)
	if idx := strings.IndexAny(ociURL, ":@"); idx != -1 {
		ociURL = ociURL[:idx]
	}

	// Return the last segment of the path as the chart name
	return path.Base(ociURL)
}
