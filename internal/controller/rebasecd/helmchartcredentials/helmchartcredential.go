package helmchartcredentials

type HelmChartCredentialHandler interface {
	GetUsername() string
	GetPassword() string
	IsType() string
	HasCredentials() bool
}

type HelmChartCredential struct {
}

func (hcc *HelmChartCredential) HasCredentials() {
	// shared implementation
}
