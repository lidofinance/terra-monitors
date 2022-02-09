package monitors

import (
	"context"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/sirupsen/logrus"
)

type TransactionsMonitor struct {
	lastTransactionHeight SimpleMetricValue
}

func NewTransactionsMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *TransactionsMonitor {
	m := TransactionsMonitor{
		lastTransactionHeight: SimpleMetricValue{},
	}
	return &m
}

func (m *TransactionsMonitor) Name() string {
	return "TransactionsMonitor"
}

func (m *TransactionsMonitor) GetMetrics() map[MetricName]MetricValue {
	return map[MetricName]MetricValue{
		LastTransactionHeight: &m.lastTransactionHeight,
	}
}

func (m *TransactionsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (m *TransactionsMonitor) Handler(ctx context.Context) error {
	// TODO
	m.lastTransactionHeight.Set(1_000)
	return nil
}
