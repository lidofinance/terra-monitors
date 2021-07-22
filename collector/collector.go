package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

type Collector interface {
	Get(metric monitors.Metric) (float64, error)
	ProvidedMetrics() []monitors.Metric
	UpdateData(ctx context.Context) error
}

func NewLCDCollector(logger *logrus.Logger) LCDCollector {
	return LCDCollector{
		Metrics:   make(map[monitors.Metric]monitors.Monitor),
		logger:    logger,
		apiClient: client.NewHTTPClient(nil),
	}
}

type LCDCollector struct {
	Metrics   map[monitors.Metric]monitors.Monitor
	Monitors  []monitors.Monitor
	logger    *logrus.Logger
	apiClient *client.TerraLiteForTerra
}

func (c LCDCollector) GetApiClient() *client.TerraLiteForTerra {
	return c.apiClient
}

func (c LCDCollector) GetLogger() *logrus.Logger {
	return c.logger
}

func (c LCDCollector) ProvidedMetrics() []monitors.Metric {
	var metrics []monitors.Metric
	for m := range c.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

func (c LCDCollector) Get(metric monitors.Metric) (float64, error) {
	monitor, found := c.Metrics[metric]
	if !found {
		return 0, fmt.Errorf("monitor for metric \"%s\" not found", metric)
	}

	return monitor.Get(metric)
}

func (c *LCDCollector) UpdateData(ctx context.Context) error {
	for _, monitor := range c.Monitors {
		err := monitor.Handler(ctx)
		if err != nil {
			return fmt.Errorf("failed to update data: %w", err)
		}
	}
	return nil
}

func (c *LCDCollector) RegisterMonitor(m monitors.Monitor) {
	for _, metric := range m.ProvidedMetrics() {
		if wantedMonitor, found := c.Metrics[metric]; found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.Metrics[metric] = m

		var doesNotExistError *monitors.MetricDoesNotExistError
		_, err := m.Get(metric)
		if err != nil && errors.As(err, &doesNotExistError) {
			panic(fmt.Sprintf("register monitor %s failed. Metric validation error. %+v", m.Name(), err))
		}
	}
	c.Monitors = append(c.Monitors, m)
}
