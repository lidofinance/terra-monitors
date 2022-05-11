package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/staking"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

const (
	stakedLunaAmount   MetricName = "staked_uluna_amount"
	stakingPoolInfoUrl string     = "https://fcd.terra.dev/staking/pool"
)

type StakedLunaAmountMonitor struct {
	metrics map[MetricName]MetricValue
	client  *http.Client
	logger  *logrus.Logger
	lock    sync.RWMutex
}

func NewStakedLunaAmountMonitor(logger *logrus.Logger) *StakedLunaAmountMonitor {
	m := StakedLunaAmountMonitor{
		metrics: make(map[MetricName]MetricValue),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
		lock:   sync.RWMutex{},
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

	resp, err := m.client.Get(stakingPoolInfoUrl)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("get staking pool request failed: %w", err)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read get staking pool resp body: %w", err)
	}

	var pool stakingPoolInfoResp
	if err := json.Unmarshal(d, &pool); err != nil {
		return fmt.Errorf("failed to unmarshal staking pool resp to stakingPoolInfoResp: %w", err)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	m.setStringMetric(stakedLunaAmount, pool.Result.BondedTokens)
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

type stakingPoolInfoResp struct {
	Height string `json:"height"`
	Result struct {
		NotBondedTokens string `json:"not_bonded_tokens"`
		BondedTokens    string `json:"bonded_tokens"`
	} `json:"result"`
}
