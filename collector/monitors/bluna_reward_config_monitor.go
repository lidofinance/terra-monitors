package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const BlunaRewardConfigCRC32 Metric = "bluna_reward_config_crc32"


type BlunaRewardConfigMonitor struct {
	State   *types.BlunaRewardConfig
	metrics map[Metric]float64
	BaseMonitor
}

func NewBlunaRewardConfigMonitor(cfg config.CollectorConfig) BlunaRewardConfigMonitor {
	m := BlunaRewardConfigMonitor{
		State:   &types.BlunaRewardConfig{},
		metrics: make(map[Metric]float64),
		BaseMonitor: BaseMonitor{ContractAddress: cfg.BlunaRewardContract,
			apiClient: cfg.GetTerraClient(),
			logger:    cfg.Logger,
		},
	}

	return m
}


func (h BlunaRewardConfigMonitor) Name() string {
	return "BlunaRewardConfig"
}

func (h *BlunaRewardConfigMonitor) InitMetrics() {
	h.metrics[BlunaRewardConfigCRC32] = 0
}

func (h *BlunaRewardConfigMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[BlunaRewardConfigCRC32] = float64(crc32.ChecksumIEEE(data))
}

func (h *BlunaRewardConfigMonitor) Handler(ctx context.Context) error {
	blunaReq, blunaResp := types.CommonConfigRequest{}, types.BlunaRewardConfig{}

	err := makeStoreQuery(&blunaResp, blunaReq, ctx, h)
	if err != nil {
		return fmt.Errorf("failed to make %s request: %w", h.Name(), err)
	}
	h.State = &blunaResp
	h.updateMetrics()
	return nil
}

func (h BlunaRewardConfigMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}
