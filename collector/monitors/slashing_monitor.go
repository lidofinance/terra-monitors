package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/monitors/signinfo"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
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
	apiClient            *terraClient.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	signInfoRepository   signinfo.Repository
	logger               *logrus.Logger
	lock                 sync.RWMutex
}

func NewSlashingMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository ValidatorsRepository,
	signInfoRepository signinfo.Repository,
) *SlashingMonitor {
	m := &SlashingMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            client.New(cfg.LCD, logger),
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

	validatorsInfo, err := m.getValidatorsInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsInfo: %w", err)
	}

	for _, validatorInfo := range validatorsInfo {

		err := m.signInfoRepository.Init(ctx, validatorInfo.PubKey)
		if err != nil {
			m.logger.Errorf("failed to init signInfo repository for validator %s: %s", validatorInfo.Address, err)
			continue
		}

		if validatorInfo.Jailed {
			tmpMetrics[SlashingNumJailedValidators].Add(1)
		}
		tmpMetricVectors[SlashingNumMissedBlocks].Add(validatorInfo.Moniker, m.signInfoRepository.GetMissedBlockCounter())
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

func (m *SlashingMonitor) getValidatorsInfo(ctx context.Context) ([]types.ValidatorInfo, error) {
	validatorsAddresses, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to getWhitelistedValidatorsAddresses: %w", err)
	}

	// For each validator address, get the consensus public key (which is required to
	// later get the signing info).
	var validatorsInfo []types.ValidatorInfo
	for _, validatorAddress := range validatorsAddresses {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get validator info: %w", err)
		}

		validatorsInfo = append(validatorsInfo, validatorInfo)
	}

	return validatorsInfo, nil
}
