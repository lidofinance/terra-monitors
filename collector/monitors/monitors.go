package monitors

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const UUSDDenom = "uusd"

type Monitor interface {
	Name() string
	// Handler fetches the data to inner storage
	Handler(ctx context.Context) error
	// GetMetrics - provides single value metrics fetched by Handler method
	GetMetrics() map[MetricName]MetricValue
	// GetMetricVectors - provides set of a labeled values fetched by Handler method
	GetMetricVectors() map[MetricName]*MetricVector
}

type MetricName string

type MetricVector struct {
	values map[string]float64
	lock   sync.RWMutex
}

func (mv *MetricVector) Get(label string) float64 {
	mv.lock.RLock()
	defer mv.lock.RUnlock()
	return mv.values[label]
}

func (mv *MetricVector) Set(label string, value float64) {
	mv.lock.Lock()
	defer mv.lock.Unlock()
	mv.values[label] = value
}

func (mv *MetricVector) Add(label string, delta float64) {
	mv.lock.Lock()
	defer mv.lock.Unlock()
	mv.values[label] += delta
}

func (mv *MetricVector) Labels() []string {
	labels := make([]string, len(mv.values))
	c := 0
	for label := range mv.values {
		labels[c] = label
		c++
	}
	return labels
}

func NewMetricVector() *MetricVector {
	return &MetricVector{
		values: make(map[string]float64),
	}
}

type MetricValue interface {
	Get() float64
	Set(float64)
	Add(float64)
}

type SimpleMetricValue struct {
	value float64
	lock  sync.Mutex
}

func (b *SimpleMetricValue) Get() float64 {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.value
}

func (b *SimpleMetricValue) Set(f float64) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.value = f
}

func (b *SimpleMetricValue) Add(f float64) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.value += f
}

// ReadOnceMetric once value is read its sets to zero value
// implemented specially for update global index bot monitor. We need to accumulate data in interval (last_check,current_check)
// once value are read we set the value to zero value
type ReadOnceMetric struct {
	value float64
	lock  sync.Mutex
}

func (a *ReadOnceMetric) Get() float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	v := a.value
	a.value = 0
	return v
}

func (a *ReadOnceMetric) Set(f float64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.value = f
}

func (a *ReadOnceMetric) Add(f float64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.value += f
}

func MustRunMonitor(ctx context.Context, m Monitor, tk *time.Ticker, logger *logrus.Logger) {
	if tk == nil {
		panic("you must to initialize ticker first")
	}
	for {
		select {
		case <-tk.C:
			err := m.Handler(context.Background())
			if err != nil {
				logger.Errorf("failed to update %s data: %+v\n", m.Name(), err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func initMetrics(providedMetrics []MetricName, providedMetricVectors []MetricName, metrics map[MetricName]MetricValue, vectors map[MetricName]*MetricVector) {
	for _, metric := range providedMetrics {
		if metrics[metric] == nil {
			metrics[metric] = &SimpleMetricValue{}
		}
		metrics[metric].Set(0)
	}
	for _, metric := range providedMetricVectors {
		vectors[metric] = NewMetricVector()
	}
}

func copyMetrics(src, dst map[MetricName]MetricValue) {
	for k, v := range src {
		dst[k].Set(v.Get())
	}
}

func copyVectors(src, dst map[MetricName]*MetricVector) {
	for metricVector, vector := range src {
		dst[metricVector] = NewMetricVector()
		for _, label := range vector.Labels() {
			dst[metricVector].Set(label, vector.Get(label))
		}
	}
}
