package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const ValidatorsRegistryConfigCRC32 Metric = "validators_registry_config_crc32"


type ValidatorsRegistryConfigMonitor struct {
	State   *types.ValidatorsRegistryConfig
	metrics map[Metric]float64
	BaseMonitor
}

func NewValidatorsRegistryConfigMonitor(cfg config.CollectorConfig) ValidatorsRegistryConfigMonitor {
	m := ValidatorsRegistryConfigMonitor{
		State:   &types.ValidatorsRegistryConfig{},
		metrics: make(map[Metric]float64),
		BaseMonitor: BaseMonitor{ContractAddress: cfg.ValidatorsRegistryContract,
			apiClient: cfg.GetTerraClient(),
			logger:    cfg.Logger,
		},
	}

	return m
}


func (h ValidatorsRegistryConfigMonitor) Name() string {
	return "ValidatorsRegistryConfig"
}

func (h *ValidatorsRegistryConfigMonitor) InitMetrics() {
	h.metrics[ValidatorsRegistryConfigCRC32] = 0
}

func (h *ValidatorsRegistryConfigMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[ValidatorsRegistryConfigCRC32] = float64(crc32.ChecksumIEEE(data))
}

func (h *ValidatorsRegistryConfigMonitor) Handler(ctx context.Context) error {
	registryReq, registryResp := types.CommonConfigRequest{}, types.ValidatorsRegistryConfig{}

	err := makeStoreQuery(&registryResp, registryReq, ctx, h)
	if err != nil {
		return fmt.Errorf("failed to make %s request: %w", h.Name(), err)
	}
	h.State = &registryResp
	h.updateMetrics()
	return nil
}

func (h ValidatorsRegistryConfigMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}