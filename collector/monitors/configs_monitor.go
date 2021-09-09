package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sync"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

const (
	AirDropRegistryConfigCRC32    string     = "airdrop_registry_config_crc32"
	BlunaRewardConfigCRC32        string     = "bluna_reward_config_crc32"
	HubConfigCRC32                string     = "hub_config_crc32"
	RewardDispatcherConfigCRC32   string     = "reward_dispatcher_config_crc32"
	ValidatorsRegistryConfigCRC32 string     = "validators_registry_config_crc32"
	ConfigCRC32                   MetricName = "config_crc32"
)

type ConfigsCRC32Monitor struct {
	Contracts        map[string]string
	metrics          map[MetricName]MetricValue
	metricVectors    map[MetricName]*MetricVector
	apiClient        *client.TerraLiteForTerra
	logger           *logrus.Logger
	lock             sync.RWMutex
	contractsVersion string
}

func NewConfigsCRC32Monitor(cfg config.CollectorConfig, logger *logrus.Logger) *ConfigsCRC32Monitor {
	m := ConfigsCRC32Monitor{
		Contracts: map[string]string{
			cfg.Addresses.AirDropRegistryContract: AirDropRegistryConfigCRC32,
			cfg.Addresses.HubContract:             HubConfigCRC32,
			cfg.Addresses.RewardContract:          BlunaRewardConfigCRC32,
		},
		metrics:          make(map[MetricName]MetricValue),
		metricVectors:    make(map[MetricName]*MetricVector),
		apiClient:        cfg.GetTerraClient(),
		logger:           logger,
		lock:             sync.RWMutex{},
		contractsVersion: cfg.BassetContractsVersion,
	}
	if m.contractsVersion == config.V2Contracts {
		m.Contracts[cfg.Addresses.ValidatorsRegistryContract] = ValidatorsRegistryConfigCRC32
		m.Contracts[cfg.Addresses.RewardsDispatcherContract] = RewardDispatcherConfigCRC32
	}

	m.InitMetrics()

	return &m
}

func (m *ConfigsCRC32Monitor) providedMetricVectors() []MetricName {
	return []MetricName{ConfigCRC32}
}

func (m *ConfigsCRC32Monitor) Name() string {
	return "ConfigsCRC32Monitor"
}

func (m *ConfigsCRC32Monitor) InitMetrics() {
	initMetrics(nil, m.providedMetricVectors(), nil, m.metricVectors)
}

func (m *ConfigsCRC32Monitor) GetMetrics() map[MetricName]MetricValue {
	return nil
}

func (m *ConfigsCRC32Monitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}

func (m *ConfigsCRC32Monitor) Handler(ctx context.Context) error {
	confReq := types.CommonConfigRequest{}

	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, m.providedMetricVectors(), nil, tmpMetricVectors)

	reqRaw, err := json.Marshal(&confReq)
	if err != nil {
		return fmt.Errorf("failed to marshal %s request: %w", m.Name(), err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetQueryMsg(string(reqRaw))
	for contract, label := range m.Contracts {
		p.SetContractAddress(contract)

		resp, err := m.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
		if err != nil {
			m.logger.Errorf("failed to process %s request for label %s: %+v", m.Name(), label, err)
			continue
		}

		data, err := json.Marshal(resp.Payload.Result)
		if err != nil {
			m.logger.Errorf("failed to marshal %s: %+v", m.Name(), err)
		}

		tmpMetricVectors[ConfigCRC32].Set(label, float64(crc32.ChecksumIEEE(data)))
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	m.logger.Infoln("updated ", m.Name())
	return nil
}
