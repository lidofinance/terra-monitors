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
	GetMetrics() map[MetricName]MetricValue
	// GetMetricVectors - provides set of a labeled values fetched by Handler method
	GetMetricVectors() map[MetricName]MetricVector
}

type MetricName string

type MetricVector map[string]float64



type MetricValue interface {
	Get() float64
	Set(float64)
	Add(float64)
	Init()
}

type BasicMetricValue struct {
	value float64
}

func (b *BasicMetricValue) Get() float64 {
	return b.value
}

func (b *BasicMetricValue) Set(f float64) {
	b.value = f
}

func (b *BasicMetricValue) Add(f float64) {
	b.value += f
}

func (b *BasicMetricValue) Init() {
	b.value = 0
}

// ReadOnceMetric once value is read its sets to zero value
type ReadOnceMetric struct {
	value float64
}

func (a *ReadOnceMetric) Get() float64 {
	defer a.Init()
	return a.value
}

func (a *ReadOnceMetric) Set(f float64) {
	a.value = f
}

func (a *ReadOnceMetric) Add(f float64) {
	a.value += f
}

func (a *ReadOnceMetric) Init() {
	a.value = 0
}
