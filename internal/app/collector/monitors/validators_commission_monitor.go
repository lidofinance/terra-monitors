package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"

	"github.com/sirupsen/logrus"
)

const (
	ValidatorsCommission MetricName = "validators_commission"
)

type ValidatorsCommissionMonitor struct {
	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *client.TerraRESTApis
	validatorsRepository repositories.ValidatorsRepository
	logger               *logrus.Logger
	lock                 sync.RWMutex
}

func NewValidatorsFeeMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository repositories.ValidatorsRepository,
) *ValidatorsCommissionMonitor {
	m := ValidatorsCommissionMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		validatorsRepository: repository,
		logger:               logger,
	}
	m.InitMetrics()
	return &m
}

func (m *ValidatorsCommissionMonitor) Name() string {
	return "ValidatorsCommission"
}

func (m *ValidatorsCommissionMonitor) InitMetrics() {
	initMetrics([]MetricName{}, []MetricName{ValidatorsCommission}, m.metrics, m.metricVectors)
}

func (m *ValidatorsCommissionMonitor) Handler(ctx context.Context) error {
	// tmp* for 2stage nonblocking update data
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, []MetricName{ValidatorsCommission}, nil, tmpMetricVectors)

	validatorsAddress, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsAddress: %w", err)
	}

	for _, validatorAddress := range validatorsAddress {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		tmpMetricVectors[ValidatorsCommission].Set(validatorInfo.Moniker, validatorInfo.CommissionRate)
	}
	m.logger.Infoln("validators commission updated", m.Name())

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func (m *ValidatorsCommissionMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *ValidatorsCommissionMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}
