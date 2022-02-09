package monitors

import (
	"context"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	LastTransactionHeight MetricName = "last_transaction_height"
)

type TransactionsMonitor struct {
	lastTransactionHeight SimpleMetricValue
	apiClient             *client.TerraRESTApis
	logger                *logrus.Logger
}

func NewTransactionsMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *TransactionsMonitor {
	m := TransactionsMonitor{
		lastTransactionHeight: SimpleMetricValue{},
		apiClient:             utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:                logger,
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
