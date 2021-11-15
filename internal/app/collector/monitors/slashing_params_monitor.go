package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/client"
	"github.com/lidofinance/terra-monitors/openapi/client_bombay"
	"github.com/lidofinance/terra-monitors/openapi/client_bombay/query"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	SlashingSignedBlocksWindow MetricName = "slashing_signed_blocks_window"
)

type SlashingParamsMonitor struct {
	metrics   map[MetricName]MetricValue
	logger    *logrus.Logger
	apiClient *client_bombay.TerraLiteForTerra
}

func NewSlashingParamsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
) *SlashingParamsMonitor {
	m := &SlashingParamsMonitor{
		metrics:   make(map[MetricName]MetricValue),
		logger:    logger,
		apiClient: client.NewBombay(cfg.LCD, logger),
	}

	m.InitMetrics()
	return m
}

func (s *SlashingParamsMonitor) providedMetrics() []MetricName {
	return []MetricName{
		SlashingSignedBlocksWindow,
	}
}

func (s *SlashingParamsMonitor) InitMetrics() {
	initMetrics(s.providedMetrics(), nil, s.metrics, nil)
}

func (s *SlashingParamsMonitor) Name() string {
	return "SlashingParamsMonitor"
}

func (s *SlashingParamsMonitor) Handler(ctx context.Context) error {
	resp, err := s.apiClient.Query.SlashingParams(
		&query.SlashingParamsParams{
			Context: ctx,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get slashing params: %w", err)
	}

	err = resp.GetPayload().Validate(nil)
	if err != nil {
		return fmt.Errorf("failed to validate slashing params: %w", err)
	}

	bw, err := strconv.ParseFloat(resp.GetPayload().Params.SignedBlocksWindow, 64)
	if err != nil {
		return fmt.Errorf("failed to parse signed blocks window: %w", err)
	}

	s.metrics[SlashingSignedBlocksWindow].Set(bw)
	return nil
}

func (s *SlashingParamsMonitor) GetMetrics() map[MetricName]MetricValue {
	return s.metrics
}

func (s *SlashingParamsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}
