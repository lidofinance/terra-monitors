package extractor

import (
	"fmt"

	"github.com/lidofinance/terra-monitors/internal/app/collector"
	"github.com/lidofinance/terra-monitors/internal/app/collector/monitors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func NewPromExtractor(c *collector.Collector, logger *logrus.Logger) PromExtractor {
	p := PromExtractor{}
	p.collector = c
	p.Gauges = make(map[monitors.MetricName]prometheus.Gauge)
	p.GaugeVectors = make(map[monitors.MetricName]*prometheus.GaugeVec)
	p.GaugeMetrics = []monitors.MetricName{}
	p.GaugeMetricVectors = []monitors.MetricName{}
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
	collector          *collector.Collector
	Gauges             map[monitors.MetricName]prometheus.Gauge
	GaugeVectors       map[monitors.MetricName]*prometheus.GaugeVec
	GaugeMetrics       []monitors.MetricName
	GaugeMetricVectors []monitors.MetricName
	log                *logrus.Logger
}

// Return fully qualified metric name to be able distinguish from system ones
func (p *PromExtractor) metricName(name monitors.MetricName) string {
    return "terra_" + string(name)
}

func (p *PromExtractor) addGauge(name monitors.MetricName) {
	p.Gauges[name] = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: p.metricName(name),
		})

	prometheus.MustRegister(p.Gauges[name])
	p.GaugeMetrics = append(p.GaugeMetrics, name)
}

func (p *PromExtractor) addGaugeVector(name monitors.MetricName) {
	p.GaugeVectors[name] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: p.metricName(name),
		},
		[]string{"label"},
	)

	prometheus.MustRegister(p.GaugeVectors[name])
	p.GaugeMetricVectors = append(p.GaugeMetricVectors, name)
}

func (p *PromExtractor) updateGaugeValue(name monitors.MetricName) error {
	value, err := p.collector.Get(name)
	if err != nil {
		return fmt.Errorf("failed to update metric \"%s\": %w", name, err)
	}

	p.Gauges[name].Set(value)
	return nil
}

func (p *PromExtractor) updateGaugeVectorValue(name monitors.MetricName) error {
	vector, err := p.collector.GetVector(name)
	if err != nil {
		return fmt.Errorf("failed to update metric \"%s\": %w", name, err)
	}
	p.GaugeVectors[name].Reset()
	for _, label := range vector.Labels() {
		p.GaugeVectors[name].With(prometheus.Labels{"label": label}).Set(vector.Get(label))
	}
	return nil
}

func (p PromExtractor) UpdateMetrics() {
	for _, gaugeName := range p.GaugeMetrics {
		err := p.updateGaugeValue(gaugeName)
		if err != nil {
			p.log.Errorf("failed to update gauge value \"%s\": %v", gaugeName, err)
		}
	}
	for _, gaugeName := range p.GaugeMetricVectors {
		err := p.updateGaugeVectorValue(gaugeName)
		if err != nil {
			p.log.Errorf("failed to update gauge value \"%s\": %v", gaugeName, err)
		}
	}
}
