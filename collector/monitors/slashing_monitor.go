package monitors

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	SlashingNumJailedValidators     MetricName = "slashing_num_jailed_validators"
	SlashingNumTombstonedValidators MetricName = "slashing_num_tombstoned_validators"
	SlashingNumMissedBlocks         MetricName = "slashing_num_missed_blocks"
)

const (
	jailTimeLayout = "2006-01-02T15:04:05Z"
)

type SlashingMonitor struct {
	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *client.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	logger               *logrus.Logger
	lock                 sync.RWMutex
}

func NewSlashingMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository ValidatorsRepository,
) *SlashingMonitor {
	m := &SlashingMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            cfg.GetTerraClient(),
		validatorsRepository: repository,
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
		signingInfoResponse, err := m.apiClient.Transactions.GetSlashingValidatorsValidatorPubKeySigningInfo(
			&transactions.GetSlashingValidatorsValidatorPubKeySigningInfoParams{
				ValidatorPubKey: validatorInfo.PubKey,
				Context:         ctx,
			},
		)
		if err != nil {
			m.logger.Errorf("failed to GetSlashingSigningInfos for validator %s: %s", validatorInfo.Address, err)
			continue
		}
		if err := signingInfoResponse.GetPayload().Validate(nil); err != nil {
			m.logger.Errorf("failed to validate SignInfo for validator %s: %s", validatorInfo.Address, err)
			continue
		}

		var signingInfo = signingInfoResponse.GetPayload().Result

		if validatorInfo.Jailed {
			tmpMetrics[SlashingNumJailedValidators].Add(1)
		}

		// No blocks is sent as "", not as "0".
		if len(*signingInfo.MissedBlocksCounter) > 0 {
			// If the current block is greater than minHeight and the validator's MissedBlocksCounter is
			// greater than maxMissed, they will be slashed. So numMissedBlocks > 0 does not mean that we
			// are already slashed, but is alarming. Note: Liveness slashes do NOT lead to a tombstoning.
			// https://docs.terra.money/dev/spec-slashing.html#begin-block
			numMissedBlocks, err := strconv.ParseInt(*signingInfo.MissedBlocksCounter, 10, 64)
			if err != nil {
				m.logger.Errorf("failed to Parse `missed_blocks_counter:`: %s", err)
			} else {
				if numMissedBlocks > 0 {
					tmpMetricVectors[SlashingNumMissedBlocks].Add(validatorInfo.Moniker, float64(numMissedBlocks))
				}
			}
		}

		if *signingInfo.Tombstoned {
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
