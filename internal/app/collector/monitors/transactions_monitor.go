package monitors

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/transactions"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	MonitoredTransactionId MetricName = "monitored_transaction_id"
)

type TransactionsMonitor struct {
	apiClient     *client.TerraRESTApis
	logger        *logrus.Logger
	addresses     []string
	metricVectors map[MetricName]*MetricVector
	lock          sync.RWMutex
}

func NewTransactionsMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *TransactionsMonitor {
	m := TransactionsMonitor{
		apiClient:     utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:        logger,
		addresses:     cfg.Addresses.MonitoredAccounts,
		metricVectors: make(map[MetricName]*MetricVector),
		lock:          sync.RWMutex{},
	}
	m.logger.Infof("Initialized transactions monitor for addresses: %v", m.addresses)
	m.InitMetrics()

	return &m
}

func (m *TransactionsMonitor) Name() string {
	return "TransactionsMonitor"
}

func (m *TransactionsMonitor) GetMetrics() map[MetricName]MetricValue {
	return nil
}

func (m *TransactionsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}

func (m *TransactionsMonitor) Handler(ctx context.Context) error {
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, m.providedMetricVectors(), nil, tmpMetricVectors)

	for _, address := range m.addresses {
		txs, err := m.queryTxs(ctx, address)
		if err != nil {
			m.logger.Errorf("Could not query transactions for %s. Error: %v", address, err)
			return err
		}

		txID, err := fetchIDFromLastTx(txs)
		if err != nil {
			m.logger.Errorf("Could not fetch ID from transactions for %v. Error: %v", address, err)
			return err
		}

		tmpMetricVectors[MonitoredTransactionId].Set(address, *txID)

		m.logger.Infof("Successfully retrieved last transaction txID for %v", address)
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func (m *TransactionsMonitor) InitMetrics() {
	initMetrics(nil, m.providedMetricVectors(), nil, m.metricVectors)
}

func (m *TransactionsMonitor) providedMetricVectors() []MetricName {
	return []MetricName{MonitoredTransactionId}
}

func (m *TransactionsMonitor) queryTxs(ctx context.Context, address string) (*models.GetTxListResult, error) {
	var limit float64 = 10
	txsParams := transactions.GetV1TxsParams{}

	txsParams.SetContext(ctx)
	txsParams.SetLimit(&limit)
	txsParams.SetAccount(&address)
	txsParams.SetTimeout(10 * time.Second)
	txs, err := m.apiClient.Transactions.GetV1Txs(&txsParams)

	if err != nil {
		return nil, err
	}

	return txs.GetPayload(), nil
}

func fetchIDFromLastTx(txs *models.GetTxListResult) (*float64, error) {
	if len(txs.Txs) == 0 {
		return nil, errors.New("empty transaction list")
	}

	txId := float64(txs.Txs[0].ID)
	return &txId, nil
}
