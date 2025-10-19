package rebasecd

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChartHandler interface {
	GetChart() error
	RenderChartObjects() ([]client.Object, error)
	DeployChartObjects(objects []client.Object, ctx context.Context, k8sClient client.Client) error
	Cleanup()

	ChartName() string
	ChartVersion() string
}
