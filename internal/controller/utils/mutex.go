package utils

import (
	"fmt"
	"sync"
)

type chartMutex struct {
	mu       sync.Mutex
	refCount int
}

// chartLocks holds the mutexes
var chartLocks = make(map[string]*chartMutex)
var chartLocksMu sync.Mutex

// AcquireMutex returns the mutex for a chart+version and increments refCount
func GetMutexForChart(chartName, version string) *sync.Mutex {
	key := fmt.Sprintf("%s:%s", chartName, version)

	chartLocksMu.Lock()
	defer chartLocksMu.Unlock()

	cm, exists := chartLocks[key]
	if !exists {
		cm = &chartMutex{refCount: 1}
		chartLocks[key] = cm
	} else {
		cm.refCount++
	}
	return &cm.mu
}

// ReleaseMutex decrements refCount and removes the mutex if no one is using it
func ReleaseMutexForChart(chartName, version string) {
	key := fmt.Sprintf("%s:%s", chartName, version)

	chartLocksMu.Lock()
	defer chartLocksMu.Unlock()

	if cm, exists := chartLocks[key]; exists {
		cm.refCount--
		if cm.refCount <= 0 {
			delete(chartLocks, key)
		}
	}
}
