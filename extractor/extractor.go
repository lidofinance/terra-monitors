package extractor

import (
	"context"
	"fmt"

	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func NewPromExtractor(c collector.Collector, logger *logrus.Logger) PromExtractor {
	p := PromExtractor{}
	p.collector = c
	p.Gauges = make(map[monitors.Metric]prometheus.Gauge)
	p.GaugeVectors = make(map[monitors.Metric]*prometheus.GaugeVec)
	p.GaugeMetrics = []monitors.Metric{}
	p.GaugeMetricVectors = []monitors.Metric{}
	p.log = logger
	for _, m := range p.collector.ProvidedMetrics() {
		p.addGauge(m)
	}
	for _, m := range p.collector.ProvidedMetricVectors() {
		p.addGaugeVector(m)
	}
	return p
}

type PromExtractor struct {
	collector          collector.Collector
	Gauges             map[monitors.Metric]prometheus.Gauge
	GaugeVectors       map[monitors.Metric]*prometheus.GaugeVec
	GaugeMetrics       []monitors.Metric
	GaugeMetricVectors []monitors.Metric
	log                *logrus.Logger
}

func (p *PromExtractor) addGauge(name monitors.Metric) {
	p.Gauges[name] = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: string(name),
		})

	prometheus.MustRegister(p.Gauges[name])
	p.GaugeMetrics = append(p.GaugeMetrics, name)
}

func (p *PromExtractor) addGaugeVector(name monitors.Metric) {
	p.GaugeVectors[name] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: string(name),
		},
		[]string{"label"},
	)

	prometheus.MustRegister(p.GaugeVectors[name])
	p.GaugeMetricVectors = append(p.GaugeMetricVectors, name)
}

func (p *PromExtractor) updateGaugeValue(name monitors.Metric) error {
	value, err := p.collector.Get(name)
	if err != nil {
		return fmt.Errorf("failed to update metric \"%s\": %w", name, err)
	}

	p.Gauges[name].Set(value)
	return nil
}

func (p *PromExtractor) updateGaugeVectorValue(name monitors.Metric) error {
	vector, err := p.collector.GetVector(name)
	if err != nil {
		return fmt.Errorf("failed to update metric \"%s\": %w", name, err)
	}
	for label := range vector {
		p.GaugeVectors[name].With(prometheus.Labels{"label": label}).Set(vector[label])
	}
	return nil
}

func (p PromExtractor) UpdateMetrics(ctx context.Context) {
	errors := p.collector.UpdateData(ctx)
	for _,err := range errors {
		p.log.Errorf("failed to update collector data: %v", err)
	}

	for _, gaugeName := range p.GaugeMetrics {
		err := p.updateGaugeValue(gaugeName)
		if err != nil {
			p.log.Errorf("failed to update gauge value \"%s\": %v", gaugeName, err)
		}
	}
	for _, gaugeName := range p.GaugeMetricVectors {
		err = p.updateGaugeVectorValue(gaugeName)
		if err != nil {
			p.log.Errorf("failed to update gauge value \"%s\": %v", gaugeName, err)
		}
	}
}
