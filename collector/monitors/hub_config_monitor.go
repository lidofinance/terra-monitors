package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const (
	HubConfigCreatorCRC32                    Metric = "hub_config_creator_crc32"
	HubConfigRewardDispatcherContractCRC32   Metric = "hub_config_reward_dispatcher_contract_crc32"
	HubConfigValidatorsRegistryContractCRC32 Metric = "hub_config_validators_registry_contract_crc32"
	HubConfigBlunaTokeContractCRC32          Metric = "hub_config_bluna_token_contract_crc32"
	HubConfigStlunaTokenContractCRC32        Metric = "hub_config_stluna_token_contract_crc32"
	HubConfigAirdropRegistryContractCRC32    Metric = "hub_config_airdrop_registry_contract_crc32"
)

type HubConfigMonitor struct {
	metrics map[Metric]float64
	State   *types.HubConfig
	BaseMonitor
}

func NewHubConfigMonitor(cfg config.CollectorConfig) HubConfigMonitor {
	m := HubConfigMonitor{
		metrics: make(map[Metric]float64),
		State:   &types.HubConfig{},
		BaseMonitor: BaseMonitor{ContractAddress: cfg.HubContract,
			apiClient: cfg.GetTerraClient(),
			logger:    cfg.Logger,
		},
	}

	return m
}

func (h HubConfigMonitor) Name() string {
	return "HubConfig"
}

func (h *HubConfigMonitor) InitMetrics() {
	h.metrics[HubConfigCreatorCRC32] = 0
	h.metrics[HubConfigRewardDispatcherContractCRC32] = 0
	h.metrics[HubConfigValidatorsRegistryContractCRC32] = 0
	h.metrics[HubConfigBlunaTokeContractCRC32] = 0
	h.metrics[HubConfigStlunaTokenContractCRC32] = 0
	h.metrics[HubConfigAirdropRegistryContractCRC32] = 0
}

func (h *HubConfigMonitor) updateMetrics() {
	h.metrics[HubConfigCreatorCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.Creator)))
	h.metrics[HubConfigRewardDispatcherContractCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.RewardDispatcherContract)))
	h.metrics[HubConfigValidatorsRegistryContractCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.ValidatorsRegistryContract)))
	h.metrics[HubConfigBlunaTokeContractCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.BlunaTokenContract)))
	h.metrics[HubConfigStlunaTokenContractCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.StlunaTokenContract)))
	h.metrics[HubConfigAirdropRegistryContractCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.AirdropRegistryContract)))
}

func (h *HubConfigMonitor) Handler(ctx context.Context) error {
	hubReq, hubResp := types.CommonConfigRequest{}, types.HubConfig{}

	err := makeStoreQuery(&hubResp, hubReq, ctx, h)
	if err != nil {
		return fmt.Errorf("failed to make %s request: %w", h.Name(), err)
	}
	h.State = &hubResp
	h.updateMetrics()
	return nil
}

func (h HubConfigMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}
