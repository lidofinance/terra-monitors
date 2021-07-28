package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const HubConfigCRC32 Metric = "hub_config_crc32"

type HubConfigMonitor struct {
	State   *types.HubConfig
	metrics map[Metric]float64
	BaseMonitor
}

func NewHubConfigMonitor(cfg config.CollectorConfig) HubConfigMonitor {
	m := HubConfigMonitor{
		State:   &types.HubConfig{},
		metrics: make(map[Metric]float64),
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
	h.metrics[HubConfigCRC32] = 0
}

func (h *HubConfigMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[HubConfigCRC32] = float64(crc32.ChecksumIEEE(data))
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
