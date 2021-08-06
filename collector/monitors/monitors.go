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

type SimpleMetricValue struct {
	value float64
}

func (b *SimpleMetricValue) Get() float64 {
	return b.value
}

func (b *SimpleMetricValue) Set(f float64) {
	b.value = f
}

func (b *SimpleMetricValue) Add(f float64) {
	b.value += f
}

func (b *SimpleMetricValue) Init() {
	b.value = 0
}

// ReadOnceMetric once value is read its sets to zero value
// implemented specially for update global index bot monitor. We need to accumulate data in interval (last_check,current_check)
// once value are read we set the value to zero value
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
