package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/staking"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

const (
	stakedLunaAmount MetricName = "staked_uluna_amount"
)

type StakedLunaAmountMonitor struct {
	metrics   map[MetricName]MetricValue
	apiClient *client.TerraRESTApis
	logger    *logrus.Logger
	lock      sync.RWMutex
}

func NewStakedLunaAmountMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *StakedLunaAmountMonitor {
	m := StakedLunaAmountMonitor{
		metrics:   make(map[MetricName]MetricValue),
		apiClient: utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:    logger,
		lock:      sync.RWMutex{},
	}
	m.InitMetrics()
	return &m
}

func (m *StakedLunaAmountMonitor) Name() string {
	return "StakedLunaAmount"
}

func (m *StakedLunaAmountMonitor) providedMetrics() []MetricName {
	return []MetricName{
		stakedLunaAmount,
	}
}

func (m *StakedLunaAmountMonitor) InitMetrics() {
	initMetrics(m.providedMetrics(), nil, m.metrics, nil)
}

func (m *StakedLunaAmountMonitor) Handler(ctx context.Context) error {
	p := staking.GetStakingPoolParams{}
	p.SetContext(ctx)

	resp, err := m.apiClient.Staking.GetStakingPool(&p)
	if err != nil {
		return fmt.Errorf("failed to process GetStakingPool request: %w", err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	m.setStringMetric(stakedLunaAmount, resp.Payload.Result.BondedTokens)
	m.logger.Infoln("updated StakedLunaAmount")
	return nil
}

func (m *StakedLunaAmountMonitor) setStringMetric(metric MetricName, rawValue string) {
	v, err := cosmostypes.NewDecFromStr(rawValue)
	if err != nil {
		m.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, metric, err)
		return
	}

	value, err := v.Float64()
	if err != nil {
		m.logger.Errorf("failed to get float64 value from string \"%s\" for metric \"%s\": %+v\n", rawValue, metric, err)
		return
	}

	if m.metrics[metric] == nil {
		m.metrics[metric] = &SimpleMetricValue{}
	}

	m.metrics[metric].Set(value)
}

func (m *StakedLunaAmountMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *StakedLunaAmountMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (m *StakedLunaAmountMonitor) SetLogger(l *logrus.Logger) {
	m.logger = l
}
