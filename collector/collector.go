package collector

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
	"time"
)

type Collector interface {
	Get(metric monitors.MetricName) (float64, error)
	GetVector(metric monitors.MetricName) (*monitors.MetricVector, error)
	ProvidedMetrics() []monitors.MetricName
	ProvidedMetricVectors() []monitors.MetricName
}

func NewLCDCollector(cfg config.CollectorConfig) LCDCollector {
	return LCDCollector{
		Metrics:       make(map[monitors.MetricName]monitors.Monitor),
		MetricVectors: make(map[monitors.MetricName]monitors.Monitor),
		logger:        cfg.Logger,
		apiClient:     cfg.GetTerraClient(),
	}
}

type LCDCollector struct {
	Metrics       map[monitors.MetricName]monitors.Monitor
	MetricVectors map[monitors.MetricName]monitors.Monitor
	Monitors      []monitors.Monitor
	logger        *logrus.Logger
	apiClient     *client.TerraLiteForTerra
}

func (c LCDCollector) GetApiClient() *client.TerraLiteForTerra {
	return c.apiClient
}

func (c LCDCollector) GetLogger() *logrus.Logger {
	return c.logger
}

func (c LCDCollector) ProvidedMetrics() []monitors.MetricName {
	var metrics []monitors.MetricName
	for m := range c.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

func (c LCDCollector) ProvidedMetricVectors() []monitors.MetricName {
	var metrics []monitors.MetricName
	for m := range c.MetricVectors {
		metrics = append(metrics, m)
	}
	return metrics
}

func (c LCDCollector) Get(metric monitors.MetricName) (float64, error) {
	monitor, found := c.Metrics[metric]
	if !found {
		return 0, fmt.Errorf("monitor for metric \"%s\" not found", metric)
	}
	return monitor.GetMetrics()[metric].Get(), nil
}

func (c LCDCollector) GetVector(metric monitors.MetricName) (*monitors.MetricVector, error) {
	monitor, found := c.MetricVectors[metric]
	if !found {
		return nil, fmt.Errorf("monitor for metric vector \"%s\" not found", metric)
	}
	return monitor.GetMetricVectors()[metric], nil
}

func findMaps(key monitors.MetricName, maps ...map[monitors.MetricName]monitors.Monitor) (monitors.Monitor, bool) {
	for _, m := range maps {
		if wantedMonitor, found := m[key]; found {
			return wantedMonitor, true
		}
	}
	return nil, false
}

func (c *LCDCollector) RegisterMonitor(ctx context.Context, m monitors.Monitor) {
	for metric := range m.GetMetrics() {
		if wantedMonitor, found := findMaps(metric, c.Metrics, c.MetricVectors); found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.Metrics[metric] = m
	}
	for metric := range m.GetMetricVectors() {
		if wantedMonitor, found := findMaps(metric, c.Metrics, c.MetricVectors); found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.MetricVectors[metric] = m
	}
	c.Monitors = append(c.Monitors, m)

	// first initial data fetching
	err := m.Handler(ctx)
	if err != nil {
		c.logger.Errorf("failed to update %s data: %+v\n", m.Name(), err)
	}

	// running fetching data in background
	tk := time.NewTicker(30 * time.Second)
	go monitors.MustRunMonitor(ctx, m, tk, c.logger)

}
