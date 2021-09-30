package monitors

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client_bombay"
	"github.com/lidofinance/terra-monitors/openapi/client_bombay/query"
	"github.com/sirupsen/logrus"
)

const (
	FailedRedelegations = "failed_redelegations"
)

type FailedRedelegationsMonitor struct {
	lock sync.RWMutex

	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *terraClient.TerraLiteForTerra
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
		apiClient:            client.NewBombay(cfg.LCD, logger),
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

	delegationsResponse, err := m.apiClient.Query.DelegatorDelegations(&query.DelegatorDelegationsParams{
		DelegatorAddr: m.hubAddress,
		Context:       ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to get whitelistedValidators of delegator: %w", err)
	}

	if err := delegationsResponse.GetPayload().Validate(nil); err != nil {
		return fmt.Errorf("failed to validate delegator's validators response: %w", err)
	}

	// New cosmos endpoint has a pagination there with a default limit 100 entities per query.
	// The hub contract has a much less delegations (16 now, 100+ validators is impossible I think),
	// so there is no need in pagination logic
	for _, delegation := range delegationsResponse.GetPayload().DelegationResponses {
		if err := delegation.Validate(nil); err != nil {
			return fmt.Errorf("failed to validate delegation: %w", err)
		}

		if delegation.Delegation == nil {
			return fmt.Errorf("failed to validate delegation: delegaion is nil")
		}

		delegatedAmount := uint64(0)
		if delegation.Balance != nil {
			delegatedAmount, err = strconv.ParseUint(delegation.Balance.Amount, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse delegation amount: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get delegation balance: balance is nil")
		}

		tmpMetricVectors[FailedRedelegations].Set(delegation.Delegation.ValidatorAddress, 0)
		// if delegated amount is greater than zero and the whitelisted validators don't contain a validator
		// that means a redelegation was not successful
		if delegatedAmount > 0 && !contains(whitelistedValidators, delegation.Delegation.ValidatorAddress) {
			tmpMetricVectors[FailedRedelegations].Set(delegation.Delegation.ValidatorAddress, 1)
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
