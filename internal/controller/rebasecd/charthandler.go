package rebasecd

type ChartHandler interface {
	GetChart() error
	RenderChart() error
	DeployChart() error
}
