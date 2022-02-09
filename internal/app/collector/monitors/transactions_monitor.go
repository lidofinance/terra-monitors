package monitors

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/transactions"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	LastTransactionHeight MetricName = "last_transaction_height"
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
		addresses:     cfg.MonitoredAccountAddresses,
		metricVectors: make(map[MetricName]*MetricVector),
		lock:          sync.RWMutex{},
	}
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
			m.logger.Errorf("Could not query transactions for %v. Error: %v", address, err)
			return err
		}

		height, err := fetchHeightFromLastTx(txs)
		if err != nil {
			m.logger.Errorf("Could not height from transactions for %v. Error: %v", address, err)
			return err
		}

		tmpMetricVectors[LastTransactionHeight].Set(address, *height)

		m.logger.Infof("Successfully retrieved last transaction height for %v", address)
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
	return []MetricName{LastTransactionHeight}
}

func (m *TransactionsMonitor) queryTxs(ctx context.Context, address string) (*models.GetTxListResult, error) {
	// TODO: why limit in float?
	var limit float64 = 1
	p := transactions.GetV1TxsParams{}
	txsParams := p.
		// WithChainID(). TODO: set ChainId?
		WithContext(ctx).
		WithLimit(&limit).
		WithAccount(&address)
	txs, err := m.apiClient.Transactions.GetV1Txs(txsParams)
	fmt.Printf("%v", txs)

	if err != nil {
		return nil, err
	}

	return txs.GetPayload(), nil
}

func fetchHeightFromLastTx(txs *models.GetTxListResult) (*float64, error) {
	if len(txs.Txs) == 0 {
		return nil, errors.New("empty transaction list")
	}

	tx := txs.Txs[0]

	// FIXME: Is it transaction total height or block height????
	var height float64
	_, err := fmt.Sscanf(*tx.Height, "%g", &height)
	if err != nil {
		return nil, err
	}

	return &height, nil // Interested only in first transaction TODO: breaks only first loop?
}
