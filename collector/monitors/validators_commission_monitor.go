package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	ValidatorsCommission MetricName = "validators_commission"
)

type ValidatorsCommissionMonitor struct {
	metrics              map[MetricName]float64
	metricVectors        map[MetricName]MetricVector
	apiClient            *client.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	logger               *logrus.Logger
}

func NewValidatorsFeeMonitor(cfg config.CollectorConfig, repository ValidatorsRepository) ValidatorsCommissionMonitor {
	m := ValidatorsCommissionMonitor{
		metrics:              make(map[MetricName]float64),
		metricVectors:        make(map[MetricName]MetricVector),
		apiClient:            cfg.GetTerraClient(),
		validatorsRepository: repository,
		logger:               cfg.Logger,
	}

	return m
}

func (m *ValidatorsCommissionMonitor) Name() string {
	return "ValidatorsCommission"
}

func (m *ValidatorsCommissionMonitor) InitMetrics() {
	m.metricVectors = map[MetricName]MetricVector{
		ValidatorsCommission: make(MetricVector),
	}
}

func (m *ValidatorsCommissionMonitor) Handler(ctx context.Context) error {
	m.InitMetrics()

	validatorsAddress, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsAddress: %w", err)
	}

	for _, validatorAddress := range validatorsAddress {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		m.metricVectors[ValidatorsCommission][validatorInfo.Moniker] = validatorInfo.CommissionRate
	}
	m.logger.Infoln("validators commission updated", m.Name())
	return nil
}

func (m *ValidatorsCommissionMonitor) GetMetrics() map[MetricName]float64 {
	return m.metrics
}

func (m ValidatorsCommissionMonitor) GetMetricVectors() map[MetricName]MetricVector {
	return m.metricVectors
}
