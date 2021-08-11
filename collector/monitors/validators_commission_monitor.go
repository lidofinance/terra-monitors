package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	ValidatorsCommission MetricName = "validators_commission"
)

type ValidatorsCommissionMonitor struct {
	metrics       map[MetricName]MetricValue
	metricVectors map[MetricName]*MetricVector
	// tmp* for 2stage nonblocking update data
	tmpMetrics           map[MetricName]MetricValue
	tmpMetricVectors     map[MetricName]*MetricVector
	apiClient            *client.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	logger               *logrus.Logger
	lock                 sync.Mutex
}

func NewValidatorsFeeMonitor(cfg config.CollectorConfig, repository ValidatorsRepository) ValidatorsCommissionMonitor {
	m := ValidatorsCommissionMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		tmpMetrics:           make(map[MetricName]MetricValue),
		tmpMetricVectors:     make(map[MetricName]*MetricVector),
		apiClient:            cfg.GetTerraClient(),
		validatorsRepository: repository,
		logger:               cfg.Logger,
	}
	m.InitMetrics()
	initMetrics([]MetricName{}, []MetricName{ValidatorsCommission}, m.tmpMetrics, m.tmpMetricVectors)
	return m
}

func (m *ValidatorsCommissionMonitor) Name() string {
	return "ValidatorsCommission"
}

func (m *ValidatorsCommissionMonitor) InitMetrics() {
	initMetrics([]MetricName{}, []MetricName{ValidatorsCommission}, m.metrics, m.metricVectors)
}

func (m *ValidatorsCommissionMonitor) Handler(ctx context.Context) error {
	//m.InitMetrics()
	initMetrics([]MetricName{}, []MetricName{ValidatorsCommission}, m.tmpMetrics, m.tmpMetricVectors)

	validatorsAddress, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsAddress: %w", err)
	}

	for _, validatorAddress := range validatorsAddress {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		m.tmpMetricVectors[ValidatorsCommission].Set(validatorInfo.Moniker, validatorInfo.CommissionRate)
	}
	m.logger.Infoln("validators commission updated", m.Name())

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(m.tmpMetricVectors, m.metricVectors)

	return nil
}

func (m *ValidatorsCommissionMonitor) GetMetrics() map[MetricName]MetricValue {
	return m.metrics
}

func (m ValidatorsCommissionMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return m.metricVectors
}
