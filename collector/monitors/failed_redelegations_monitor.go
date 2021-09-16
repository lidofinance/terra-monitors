package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/staking"
	"github.com/sirupsen/logrus"
)

const (
	FailedRedelegations = "failed_redelegations"
)

type FailedRedelegationsMonitor struct {
	lock sync.RWMutex

	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *client.TerraLiteForTerra
	logger               *logrus.Logger
	validatorsRepository ValidatorsRepository

	hubAddress string
}

func NewFailedRedelegationsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository ValidatorsRepository,
) *FailedRedelegationsMonitor {
	m := FailedRedelegationsMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            cfg.GetTerraClient(),
		logger:               logger,
		validatorsRepository: repository,
		hubAddress:           cfg.Addresses.HubContract,
	}
	m.InitMetrics()

	return &m
}

func (m *FailedRedelegationsMonitor) Name() string {
	return "FailedRedelegationsMonitor"
}

func (m *FailedRedelegationsMonitor) providedMetrics() []MetricName {
	return []MetricName{
		FailedRedelegations,
	}
}

func (m *FailedRedelegationsMonitor) InitMetrics() {
	initMetrics([]MetricName{}, []MetricName{FailedRedelegations}, m.metrics, m.metricVectors)
}

func (m *FailedRedelegationsMonitor) GetMetrics() map[MetricName]MetricValue {
	return m.metrics
}

func (m *FailedRedelegationsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return m.metricVectors
}

func (m *FailedRedelegationsMonitor) Handler(ctx context.Context) error {
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, []MetricName{FailedRedelegations}, nil, tmpMetricVectors)

	whitelistedValidators, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whiltelisted whitelistedValidators for %s: %w", m.Name(), err)
	}

	validatorsResponse, err := m.apiClient.Staking.GetStakingDelegatorsDelegatorAddrValidators(&staking.GetStakingDelegatorsDelegatorAddrValidatorsParams{
		DelegatorAddr: m.hubAddress,
		Context:       ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to get whitelistedValidators of delegator: %w", err)
	}

	if err := validatorsResponse.GetPayload().Validate(nil); err != nil {
		return fmt.Errorf("failed to validate delegator's validators response: %w", err)
	}

	for _, validator := range validatorsResponse.GetPayload().Result {
		if err := validator.OperatorAddress.Validate(nil); err != nil {
			return fmt.Errorf("failed to validate validator's address: %w", err)
		}

		if !contains(whitelistedValidators, string(validator.OperatorAddress)) {
			tmpMetricVectors[FailedRedelegations].Set(string(validator.OperatorAddress), 1)
		}
	}

	m.logger.Infoln("updated", m.Name())

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
