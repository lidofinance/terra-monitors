package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const AirDropRegistryConfigCRC32 Metric = "airdrop_registry_config_crc32"


type AirDropRegistryConfigMonitor struct {
	State   *types.AirDropRegistryConfig
	metrics map[Metric]float64
	BaseMonitor
}

func NewAirDropRegistryConfigMonitor(cfg config.CollectorConfig) AirDropRegistryConfigMonitor {
	m := AirDropRegistryConfigMonitor{
		State:   &types.AirDropRegistryConfig{},
		metrics: make(map[Metric]float64),
		BaseMonitor: BaseMonitor{ContractAddress: cfg.ValidatorsRegistryContract,
			apiClient: cfg.GetTerraClient(),
			logger:    cfg.Logger,
		},
	}

	return m
}


func (h AirDropRegistryConfigMonitor) Name() string {
	return "AirDropRegistryConfig"
}

func (h *AirDropRegistryConfigMonitor) InitMetrics() {
	h.metrics[AirDropRegistryConfigCRC32] = 0
}

func (h *AirDropRegistryConfigMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[AirDropRegistryConfigCRC32] = float64(crc32.ChecksumIEEE(data))
}

func (h *AirDropRegistryConfigMonitor) Handler(ctx context.Context) error {
	registryReq, registryResp := types.CommonConfigRequest{}, types.AirDropRegistryConfig{}

	err := makeStoreQuery(&registryResp, registryReq, ctx, h)
	if err != nil {
		return fmt.Errorf("failed to make %s request: %w", h.Name(), err)
	}
	h.State = &registryResp
	h.updateMetrics()
	return nil
}

func (h AirDropRegistryConfigMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}
