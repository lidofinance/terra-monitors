package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/lidofinance/terra-monitors/client"
	"github.com/sirupsen/logrus"
)

type Metrics string

func ParseRequestBody(body interface{}, ret interface{}) error {
	date, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body interface{}: %w", err)
	}
	err = json.Unmarshal(date, ret)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return nil
}

type Collector interface {
	Get(metric Metrics) (float64, error)
	UpdateData(ctx context.Context) error
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewLCDCollector(RewardContractAddress string, logger *logrus.Logger) LCDCollector {
	return LCDCollector{
		Metrics:   make(map[Metrics]Monitor),
		logger:    logger,
		apiClient: client.NewHTTPClient(nil),
	}
}

type LCDCollector struct {
	Metrics       map[Metrics]Monitor
	Monitors      []Monitor
	logger        *logrus.Logger
	CollectedData CollectedData
	apiClient     *client.TerraLiteForTerra
}

func (c *LCDCollector) SetTransport(transport runtime.ClientTransport) {
	c.apiClient.SetTransport(transport)
}

func (c LCDCollector) Get(metric Metrics) (float64, error) {
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

func (c *LCDCollector) RegisterMonitor(m Monitor) {
	for _, metric := range m.ProvidedMetrics() {
		c.Metrics[metric] = m
	}
	c.Monitors = append(c.Monitors, m)
	m.SetApiClient(c.apiClient)
	m.SetLogger(c.logger)
}

type Monitor interface {
	SetApiClient(*client.TerraLiteForTerra)
	SetLogger(*logrus.Logger)
	Handler(ctx context.Context) error
	ProvidedMetrics() []Metrics
	Get(m Metrics) (float64, error)
}
