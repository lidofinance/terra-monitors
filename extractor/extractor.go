package extractor

import (
	"context"
	"fmt"

	"github.com/lidofinance/terra-monitors/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func NewPromExtractor(c collector.Collector, logger *logrus.Logger) PromExtractor {
	p := PromExtractor{}
	p.collector = c
	p.Gauges = make(map[collector.Metrics]prometheus.Gauge)
	p.GaugeMetrics = []collector.Metrics{}
	p.log = logger
	p.addGauge(collector.BlunaTotalSupply)
	p.addGauge(collector.GlobalIndex)
	return p
}

type PromExtractor struct {
	collector    collector.Collector
	Gauges       map[collector.Metrics]prometheus.Gauge
	GaugeMetrics []collector.Metrics
	log          *logrus.Logger
}

func (p *PromExtractor) addGauge(name collector.Metrics) {
	p.Gauges[name] = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: string(name),
		})
	prometheus.MustRegister(p.Gauges[name])
	p.GaugeMetrics = append(p.GaugeMetrics, name)
}

func (p *PromExtractor) updateGaugeValue(name collector.Metrics) error {
	value, err := p.collector.Get(name)
	if err != nil {
		return fmt.Errorf("failed to update metric \"%s\": %w", name, err)
	}
	p.Gauges[name].Set(value)
	return nil
}

func (p PromExtractor) UpdateMetrics(ctx context.Context) {

	err := p.collector.UpdateData(ctx)
	if err != nil {
		p.log.Errorf("failed to update collector data: %v", err)
	}
	for _, gaugeName := range p.GaugeMetrics {
		err = p.updateGaugeValue(gaugeName)
		if err != nil {
			p.log.Errorf("failed to update gauge value \"%s\": %v", gaugeName, err)
		}
	}
}
