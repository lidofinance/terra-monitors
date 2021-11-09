package monitors

import (
	"context"
	"fmt"
	"sync"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/oracle"
	"github.com/sirupsen/logrus"
)

const (
	OracleMissedVoteRate MetricName = "oracle_missed_votes_rate"
)

type OracleVotesMonitor struct {
	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *terraClient.TerraLiteForTerra
	validatorsRepository validators.ValidatorsRepository
	logger               *logrus.Logger
	lock                 sync.RWMutex
}

func NewOracleVotesMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository validators.ValidatorsRepository,
) *OracleVotesMonitor {
	m := OracleVotesMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            client.New(cfg.LCD, logger),
		validatorsRepository: repository,
		logger:               logger,
	}
	m.InitMetrics()
	return &m
}

func (m *OracleVotesMonitor) Name() string {
	return "OracleVotesMonitor"
}

func (m *OracleVotesMonitor) InitMetrics() {
	initMetrics([]MetricName{}, []MetricName{OracleMissedVoteRate}, m.metrics, m.metricVectors)
}

func (m *OracleVotesMonitor) Handler(ctx context.Context) error {
	// tmp* for 2stage nonblocking update data
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, []MetricName{OracleMissedVoteRate}, nil, tmpMetricVectors)

	validatorsAddresses, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsAddress: %w", err)
	}

	oracleParamsResponse, err := m.apiClient.Oracle.GetOracleParameters(&oracle.GetOracleParametersParams{Context: ctx})
	if err != nil {
		return fmt.Errorf("failed to get oracle parameters: %w", err)
	}

	if err := oracleParamsResponse.GetPayload().Validate(nil); err != nil {
		return fmt.Errorf("failed to validate OracleParamsResponse: %w", err)
	}

	oracleParams := oracleParamsResponse.GetPayload().Result

	for _, validatorAddress := range validatorsAddresses {
		validatorInfo, err := m.validatorsRepository.GetValidatorInfo(ctx, validatorAddress)
		if err != nil {
			return fmt.Errorf("failed to GetValidatorInfo: %w", err)
		}

		missedVotePeriodsResponse, err := m.apiClient.Oracle.GetOracleVotersValidatorMiss(
			&oracle.GetOracleVotersValidatorMissParams{Validator: validatorAddress, Context: ctx},
		)
		if err != nil {
			return fmt.Errorf("failed to get missed vote periods: %w", err)
		}

		if err := missedVotePeriodsResponse.GetPayload().Validate(nil); err != nil {
			return fmt.Errorf("failed to validate missedVotePeriodsResponse: %w", err)
		}

		oracleMissedVotePeriods, err := cosmostypes.NewDecFromStr(missedVotePeriodsResponse.GetPayload().Result)
		if err != nil {
			return fmt.Errorf("failed to parse oracleMissedVotePeriods: %w", err)
		}

		oracleMissedVotePeriodsValue, err := oracleMissedVotePeriods.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse oracleMissedVotePeriods: %w", err)
		}

		// Every validator must vote during every params.VotePeriod
		// If during every SlashWindow a validator sends fewer votes than params.VoteThreshold votes he will be slashed.

		// We know params.SlashWindow, params.VotePeriod and the number of vote periods a validator missed
		// in this oracle slash window, so:
		// votePeriodsPerWindow = params.SlashWindow / params.VotePeriod
		// missedVotesRate = (missedPeriods / votePeriodsPerWindow) * 100%
		// If missedVotesRate greater than (100% - params.VoteThreshold) validator will be slashed
		// More info: https://docs.terra.money/dev/spec-oracle.html#slashing
		slashWindow, err := cosmostypes.NewDecFromStr(oracleParams.SlashWindow)
		if err != nil {
			return fmt.Errorf("failed to parse SlashWindow: %w", err)
		}

		slashWindowValue, err := slashWindow.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse slashWindow: %w", err)
		}

		votePeriod, err := cosmostypes.NewDecFromStr(oracleParams.VotePeriod)
		if err != nil {
			return fmt.Errorf("failed to parse VotePeriod: %w", err)
		}

		votePeriodValue, err := votePeriod.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse votePeriod: %w", err)
		}

		votePeriodsPerSlashWindow := slashWindowValue / votePeriodValue
		missedVotesRate := oracleMissedVotePeriodsValue / votePeriodsPerSlashWindow

		tmpMetricVectors[OracleMissedVoteRate].Set(validatorInfo.Moniker, missedVotesRate)
	}
	m.logger.Infoln("Oracle missed votes updated", m.Name())

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func (m *OracleVotesMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *OracleVotesMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}
