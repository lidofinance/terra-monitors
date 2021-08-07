package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	ValidatorsFee MetricName = "validators_fee"
)

type ValidatorsFeeMonitor struct {
	metrics              map[MetricName]float64
	metricVectors        map[MetricName]MetricVector
	apiClient            *client.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	logger               *logrus.Logger
}

func NewValidatorsFeeMonitor(cfg config.CollectorConfig, repository ValidatorsRepository) ValidatorsFeeMonitor {
	m := ValidatorsFeeMonitor{
		metrics:              make(map[MetricName]float64),
		metricVectors:        make(map[MetricName]MetricVector),
		apiClient:            cfg.GetTerraClient(),
		validatorsRepository: repository,
		logger:               cfg.Logger,
	}

	return m
}

func (m *ValidatorsFeeMonitor) Name() string {
	return "ValidatorsFee"
}

func (m *ValidatorsFeeMonitor) InitMetrics() {
	m.metricVectors = map[MetricName]MetricVector{
		ValidatorsFee: make(MetricVector),
	}
}

func (m *ValidatorsFeeMonitor) Handler(ctx context.Context) error {
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

		m.metricVectors[ValidatorsFee][validatorInfo.Moniker] = validatorInfo.CommissionRate
	}
	m.logger.Infoln("validators fee updated", m.Name())
	return nil
}

func (m *ValidatorsFeeMonitor) GetMetrics() map[MetricName]float64 {
	return m.metrics
}

func (m ValidatorsFeeMonitor) GetMetricVectors() map[MetricName]MetricVector {
	return m.metricVectors
}
