package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/monitors/delegations"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client_bombay"
	"github.com/sirupsen/logrus"
)

const (
	FailedRedelegations = "failed_redelegations"
)

type FailedRedelegationsMonitor struct {
	lock sync.RWMutex

	metrics       map[MetricName]MetricValue
	metricVectors map[MetricName]*MetricVector
	apiClient     *terraClient.TerraLiteForTerra
	logger        *logrus.Logger

	validatorsRepository  ValidatorsRepository
	delegationsRepository delegations.Repository

	hubAddress string
}

func NewFailedRedelegationsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository ValidatorsRepository,
	delegationsRepository delegations.Repository,
) *FailedRedelegationsMonitor {
	m := FailedRedelegationsMonitor{
		metrics:       make(map[MetricName]MetricValue),
		metricVectors: make(map[MetricName]*MetricVector),
		apiClient:     client.NewBombay(cfg.LCD, logger),
		logger:        logger,

		validatorsRepository:  repository,
		delegationsRepository: delegationsRepository,

		hubAddress: cfg.Addresses.HubContract,
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

	delegationsResponse, err := m.delegationsRepository.GetDelegationsFromAddress(ctx, m.hubAddress)
	if err != nil {
		return fmt.Errorf("failed to GetDelegationsFromAddress: %w", err)
	}

	for _, delegation := range delegationsResponse {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, delegation.ValidatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		label := fmt.Sprintf("%s (%s)", delegation.ValidatorAddress, validatorInfo.Moniker)

		tmpMetricVectors[FailedRedelegations].Set(label, 0)

		// if delegated amount is greater than zero and the whitelisted validators don't contain a validator
		// that means a redelegation was not successful
		if !delegation.DelegationAmount.IsZero() && !containsString(whitelistedValidators, delegation.ValidatorAddress) {
			tmpMetricVectors[FailedRedelegations].Set(label, 1)
		}
	}

	m.logger.Infoln("updated", m.Name())

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
