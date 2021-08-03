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
	GetMetrics() map[Metric]float64
	// GetMetricVectors - provides set of a labeled values fetched by Handler method
	GetMetricVectors() map[Metric]map[string]float64
}

type Metric string
