package monitors

import (
	"context"
)

type Monitor interface {
	Name() string
	InitMetrics()
	// Handler fetches the data to inner storage
	Handler(ctx context.Context) error
	// GetMetrics - provides metrics fetched by Handler method
	GetMetrics() map[Metric]float64
}

type Metric string
