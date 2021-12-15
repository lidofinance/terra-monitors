package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/config"

	"github.com/lidofinance/terra-repositories/signinfo"
	"github.com/lidofinance/terra-repositories/validators"

	"github.com/sirupsen/logrus"
)

const (
	SlashingNumJailedValidators     MetricName = "slashing_num_jailed_validators"
	SlashingNumTombstonedValidators MetricName = "slashing_num_tombstoned_validators"
	SlashingNumMissedBlocks         MetricName = "slashing_num_missed_blocks"
)

type SlashingMonitor struct {
	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	validatorsRepository repositories.ValidatorsRepository
	signInfoRepository   *signinfo.Repository
	logger               *logrus.Logger
	lock                 sync.RWMutex
}

func NewSlashingMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository repositories.ValidatorsRepository,
	signInfoRepository *signinfo.Repository,
) *SlashingMonitor {
	m := &SlashingMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		validatorsRepository: repository,
		signInfoRepository:   signInfoRepository,
		logger:               logger,
		lock:                 sync.RWMutex{},
	}

	m.InitMetrics()

	return m
}

func (m *SlashingMonitor) Name() string {
	return "Slashing"
}

func (m *SlashingMonitor) providedMetrics() []MetricName {
	return []MetricName{
		SlashingNumJailedValidators,
		SlashingNumTombstonedValidators,
	}
}

func (m *SlashingMonitor) providedMetricVectors() []MetricName {
	return []MetricName{
		SlashingNumMissedBlocks,
	}
}

func (m *SlashingMonitor) InitMetrics() {
	initMetrics(m.providedMetrics(), m.providedMetricVectors(), m.metrics, m.metricVectors)
}

func (m *SlashingMonitor) Handler(ctx context.Context) error {
	// tmp* for 2stage nonblocking update data
	tmpMetrics := make(map[MetricName]MetricValue)
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(m.providedMetrics(), m.providedMetricVectors(), tmpMetrics, tmpMetricVectors)

	validatorsInfo, err := getValidatorsInfo(ctx, m.validatorsRepository)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsInfo: %w", err)
	}

	for _, validatorInfo := range validatorsInfo {
		if err := m.signInfoRepository.Init(ctx, validatorInfo.PubKey); err != nil {
			m.logger.Errorf("failed to init signInfo repository for validator %s: %s", validatorInfo.Address, err)
			continue
		}

		missedBlocks, err := m.signInfoRepository.GetMissedBlockCounter()
		if err != nil {
			m.logger.Errorf("failed to Parse `missed_blocks_counter:`: %v", err)
		} else {
			tmpMetricVectors[SlashingNumMissedBlocks].Add(validatorInfo.Moniker, missedBlocks)
		}
		if validatorInfo.Jailed {
			tmpMetrics[SlashingNumJailedValidators].Add(1)
		}
		if m.signInfoRepository.GetTombstoned() {
			tmpMetrics[SlashingNumTombstonedValidators].Add(1)
		}
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	copyMetrics(tmpMetrics, m.metrics)
	copyVectors(tmpMetricVectors, m.metricVectors)

	m.logger.Infoln("updated", m.Name())
	return nil
}

func (m *SlashingMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *SlashingMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}

func getValidatorsInfo(ctx context.Context, validatorsRepository repositories.ValidatorsRepository) ([]validators.ValidatorInfo, error) {
	validatorsAddresses, err := validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to getWhitelistedValidatorsAddresses: %w", err)
	}

	// For each validator address, get the consensus public key (which is required to
	// later get the signing info).
	var validatorsInfo []validators.ValidatorInfo
	for _, validatorAddress := range validatorsAddresses {
		validatorInfo, err := validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get validator info: %w", err)
		}

		validatorsInfo = append(validatorsInfo, validatorInfo)
	}

	return validatorsInfo, nil
}
