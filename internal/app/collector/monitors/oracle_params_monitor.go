package monitors

import (
	"context"
	"fmt"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/query"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

const (
	OracleMissedVotesWindow MetricName = "oracle_missed_votes_window"
)

type OracleParamsMonitor struct {
	metrics   map[MetricName]MetricValue
	logger    *logrus.Logger
	apiClient *client.TerraRESTApis
}

func NewOracleParamsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
) *OracleParamsMonitor {
	m := &OracleParamsMonitor{
		metrics:   make(map[MetricName]MetricValue),
		logger:    logger,
		apiClient: utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
	}

	m.InitMetrics()
	return m
}

func (s *OracleParamsMonitor) providedMetrics() []MetricName {
	return []MetricName{
		OracleMissedVotesWindow,
	}
}

func (s *OracleParamsMonitor) InitMetrics() {
	initMetrics(s.providedMetrics(), nil, s.metrics, nil)
}

func (s *OracleParamsMonitor) Name() string {
	return "OracleParamsMonitor"
}

func (s *OracleParamsMonitor) Handler(ctx context.Context) error {
	resp, err := s.apiClient.Query.OracleParams(
		&query.OracleParamsParams{
			Context: ctx,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get oracle params: %w", err)
	}

	err = resp.GetPayload().Validate(nil)
	if err != nil {
		return fmt.Errorf("failed to validate oracle params: %w", err)
	}

	sw, err := cosmostypes.NewDecFromStr(resp.GetPayload().Params.SlashWindow)
	if err != nil {
		return fmt.Errorf("failed to parse missed votes window: %w", err)
	}

	votesWindow, err := sw.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert missed votes value from cosmostypes.Dec to float64: %w", err)
	}

	s.metrics[OracleMissedVotesWindow].Set(votesWindow)
	return nil
}

func (s *OracleParamsMonitor) GetMetrics() map[MetricName]MetricValue {
	return s.metrics
}

func (s *OracleParamsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}
