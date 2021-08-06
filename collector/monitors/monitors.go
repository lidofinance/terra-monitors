package monitors

import (
	"context"
)

type Monitor interface {
	Name() string
	InitMetrics()
	// Handler fetches the data to inner storage
	Handler(ctx context.Context) error
	// GetMetrics - provides single value metrics fetched by Handler method
	GetMetrics() map[MetricName]float64
	// GetMetricVectors - provides set of a labeled values fetched by Handler method
	GetMetricVectors() map[MetricName]MetricVector
}

type MetricName string

type MetricVector map[string]float64
