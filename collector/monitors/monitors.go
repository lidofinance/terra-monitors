package monitors

import (
	"context"
	"fmt"
)

type Monitor interface {
	Name() string
	// Handler fetches the data to inner storage
	Handler(ctx context.Context) error
	ProvidedMetrics() []Metric
	// Get - provides metric fetched by Handler method
	// In case metric does not exist on the monitor, you MUST return MetricDoesNotExistError type error
	Get(m Metric) (float64, error)
}

type Metric string

type MetricDoesNotExistError struct {
	metricName Metric
}

func (m *MetricDoesNotExistError) Error() string {
	return fmt.Sprintf("metric \"%s\" does not exists on monitor", m.metricName)
}
