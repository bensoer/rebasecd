package utils

import (
	"fmt"
	"sync"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// chartLocks holds a map of mutexes for each chart+version
var chartLocks = make(map[string]*sync.Mutex)
var chartLocksMu sync.Mutex // protects the map itself

// GetMutexForChart returns a mutex for the given chart+version combination.
// It creates a new mutex if one does not already exist.
func GetMutexForChart(chartName, version string) *sync.Mutex {
	key := fmt.Sprintf("%s:%s", chartName, version)

	// Lock the map itself
	chartLocksMu.Lock()
	defer chartLocksMu.Unlock()

	// If a mutex already exists for this chart/version, return it
	if m, exists := chartLocks[key]; exists {
		return m
	}

	// Otherwise, create a new mutex and store it
	m := &sync.Mutex{}
	chartLocks[key] = m
	return m
}

func NewActionConfigAndSettings(namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", func(format string, v ...interface{}) {
		fmt.Printf(format, v...)
	}); err != nil {
		actionConfig = nil
		return nil, nil, fmt.Errorf("failed to init helm config: %w", err)
	}

	return actionConfig, settings, nil
}
