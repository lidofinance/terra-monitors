package monitors

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/delegations"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/math"
	"github.com/sirupsen/logrus"
)

const (
	DelegationsDistributionImbalance MetricName = "delegations_distribution_imbalance"
)

type DelegationsDistributionMonitor struct {
	lock sync.RWMutex

	metrics       map[MetricName]MetricValue
	metricVectors map[MetricName]*MetricVector
	logger        *logrus.Logger

	validatorsRepository  validators.ValidatorsRepository
	delegationsRepository delegations.Repository

	hubAddress string
	nMads      int64
}

func NewDelegationsDistributionMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	validatorsRepository validators.ValidatorsRepository,
	delegationsRepository delegations.Repository,
) *DelegationsDistributionMonitor {
	m := &DelegationsDistributionMonitor{
		lock: sync.RWMutex{},

		metrics:       make(map[MetricName]MetricValue),
		metricVectors: make(map[MetricName]*MetricVector),
		logger:        logger,

		validatorsRepository:  validatorsRepository,
		delegationsRepository: delegationsRepository,

		hubAddress: cfg.Addresses.HubContract,
		nMads:      cfg.DelegationsDistributionConfig.NumMedianAbsoluteDeviations,
	}

	m.InitMetrics()

	return m
}

func (m *DelegationsDistributionMonitor) Name() string {
	return "DelegationsDistribution"
}

func (m *DelegationsDistributionMonitor) providedMetrics() []MetricName {
	return []MetricName{}
}

func (m *DelegationsDistributionMonitor) providedMetricVectors() []MetricName {
	return []MetricName{
		DelegationsDistributionImbalance,
	}
}

func (m *DelegationsDistributionMonitor) InitMetrics() {
	initMetrics(m.providedMetrics(), m.providedMetricVectors(), m.metrics, m.metricVectors)
}

func (m *DelegationsDistributionMonitor) Handler(ctx context.Context) error {
	// for 2stage nonblocking update data
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(m.providedMetrics(), m.providedMetricVectors(), make(map[MetricName]MetricValue), tmpMetricVectors)

	delegationsResponse, err := m.delegationsRepository.GetDelegationsFromAddress(ctx, m.hubAddress)
	if err != nil {
		return fmt.Errorf("failed to GetDelegationsFromAddress: %w", err)
	}

	var delegationAmounts []*big.Int
	for _, delegation := range delegationsResponse {
		delegationAmounts = append(delegationAmounts, delegation.DelegationAmount.BigInt())
	}

	var outliers = math.GetMeanAbsoluteDeviationOutliers(delegationAmounts, m.nMads)
	for idx := range delegationsResponse {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, delegationsResponse[idx].ValidatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		label := fmt.Sprintf("%s (%s)", delegationsResponse[idx].ValidatorAddress, validatorInfo.Moniker)
		tmpMetricVectors[DelegationsDistributionImbalance].Set(label, 0)

		if containsInt(outliers, idx) {
			tmpMetricVectors[DelegationsDistributionImbalance].Set(label, 1)
		}
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	m.logger.Infoln("updated", m.Name())
	return nil
}

func (m *DelegationsDistributionMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *DelegationsDistributionMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}

func containsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
