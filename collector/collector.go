package collector

import (
	"context"
	"fmt"

	"github.com/lidofinance/terra-monitors/client"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/sirupsen/logrus"
)

type Collector interface {
	Get(metric monitors.Metric) (float64, error)
	ProvidedMetrics() []monitors.Metric
	UpdateData(ctx context.Context) error
}

func NewLCDCollector(cfg config.CollectorConfig) LCDCollector {
	return LCDCollector{
		Metrics:   make(map[monitors.Metric]monitors.Monitor),
		logger:    cfg.Logger,
		apiClient: cfg.GetTerraClient(),
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
	return monitor.GetMetrics()[metric], nil
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
	m.InitMetrics()
	for metric := range m.GetMetrics() {
		if wantedMonitor, found := c.Metrics[metric]; found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.Metrics[metric] = m
	}
	c.Monitors = append(c.Monitors, m)
}
