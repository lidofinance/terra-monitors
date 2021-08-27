package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

const (
	AirDropRegistryConfigCRC32    MetricName = "airdrop_registry_config_crc32"
	BlunaRewardConfigCRC32        MetricName = "bluna_reward_config_crc32"
	HubConfigCRC32                MetricName = "hub_config_crc32"
	RewardDispatcherConfigCRC32   MetricName = "reward_dispatcher_config_crc32"
	ValidatorsRegistryConfigCRC32 MetricName = "validators_registry_config_crc32"
)

type ConfigsCRC32Monitor struct {
	Contracts        map[string]MetricName
	metrics          map[MetricName]MetricValue
	apiClient        *client.TerraLiteForTerra
	logger           *logrus.Logger
	contractsVersion string
}

func NewConfigsCRC32Monitor(cfg config.CollectorConfig, logger *logrus.Logger) ConfigsCRC32Monitor {
	m := ConfigsCRC32Monitor{
		Contracts: map[string]MetricName{
			cfg.Addresses.AirDropRegistryContract: AirDropRegistryConfigCRC32,
			cfg.Addresses.HubContract:             HubConfigCRC32,
			cfg.Addresses.RewardContract:          BlunaRewardConfigCRC32,
		},
		metrics:          make(map[MetricName]MetricValue),
		apiClient:        cfg.GetTerraClient(),
		logger:           logger,
		contractsVersion: cfg.BassetContractsVersion,
	}
	if m.contractsVersion == "2" {
		m.Contracts[cfg.Addresses.ValidatorsRegistryContract] = ValidatorsRegistryConfigCRC32
		m.Contracts[cfg.Addresses.RewardsDispatcherContract] = RewardDispatcherConfigCRC32
	}

	m.InitMetrics()

	return m
}

func (m *ConfigsCRC32Monitor) providedMetrics() []MetricName {
	providedMetrics := []MetricName{
		AirDropRegistryConfigCRC32,
		HubConfigCRC32,
		BlunaRewardConfigCRC32,
	}
	// ValidatorsRegistry and RewardDispatcher contracts are presented only for v2 contracts,
	// we dont need these metrics on v1 monitor instance
	if m.contractsVersion == "2" {
		providedMetrics = append(providedMetrics, ValidatorsRegistryConfigCRC32, RewardDispatcherConfigCRC32)
	}
	return providedMetrics
}

func (m ConfigsCRC32Monitor) Name() string {
	return "ConfigsCRC32Monitor"
}

func (m *ConfigsCRC32Monitor) InitMetrics() {
	for _, metric := range m.providedMetrics() {
		if m.metrics[metric] == nil {
			m.metrics[metric] = &SimpleMetricValue{}
		}
		m.metrics[metric].Set(0)
	}
}

func (m ConfigsCRC32Monitor) GetMetrics() map[MetricName]MetricValue {
	return m.metrics
}

func (m ConfigsCRC32Monitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (m *ConfigsCRC32Monitor) Handler(ctx context.Context) error {
	confReq := types.CommonConfigRequest{}

	reqRaw, err := json.Marshal(&confReq)
	if err != nil {
		return fmt.Errorf("failed to marshal %s request: %w", m.Name(), err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetQueryMsg(string(reqRaw))
	for contract, metric := range m.Contracts {
		p.SetContractAddress(contract)

		resp, err := m.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
		if err != nil {
			m.logger.Errorf("failed to process %s request for metric %s: %+v", m.Name(), metric, err)
			continue
		}

		data, err := json.Marshal(resp.Payload.Result)
		if err != nil {
			m.logger.Errorf("failed to marshal %s: %+v", m.Name(), err)
		}

		m.metrics[metric].Set(float64(crc32.ChecksumIEEE(data)))
	}
	m.logger.Infoln("updated ", m.Name())
	return nil
}
