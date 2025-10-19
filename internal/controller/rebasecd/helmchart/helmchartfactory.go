package helmchart

import (
	"fmt"
	"strings"

	operatorv1alpha1 "github.com/bensoer/rebasecd/api/v1alpha1"
	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/rebasecd/credentials"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateChartHandlerFromSpec(spec *operatorv1alpha1.ApplicationSpec, namespace string, client client.Client) (rebasecd.ChartHandler, error) {

	credentials, err := credentials.CreateCredentialsFromSpec(&spec.HelmChartCredentials, namespace, client)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(spec.Url, "oci://") {
		// OCI Helm Chart
		return NewOCIHelmChart(
			spec.Url,
			spec.Version,
			spec.ReleaseName,
			namespace,
			credentials,
			spec.ValuesFiles,
		), nil
	} else if (strings.HasPrefix(spec.Url, "http://") || strings.HasPrefix(spec.Url, "https://")) && !strings.HasSuffix(spec.Url, ".git") {
		// Standard Helm Chart Repository
		return NewHelmChart(
			spec.Url,
			spec.ChartName,
			spec.Version,
			spec.ReleaseName,
			namespace,
			credentials,
			spec.ValuesFiles,
		), nil
	} else if strings.HasSuffix(spec.Url, ".git") {
		// Git Repository Helm Chart
		return NewGitHelmChart(
			spec.Url,
			spec.Path,
			spec.ReleaseName,
			namespace,
			credentials,
			spec.ValuesFiles,
		), nil
	}

	// could not determine type
	return nil, fmt.Errorf("could not determine chart type for URL: %s", spec.Url)
}
