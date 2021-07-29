package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"hash/crc32"
)

const RewardDispatcherConfigCRC32 Metric = "reward_dispatcher_config_crc32"


type RewardDispatcherConfigMonitor struct {
	State   *types.RewardDispatcherConfig
	metrics map[Metric]float64
	BaseMonitor
}

func NewRewardDispatcherConfigMonitor(cfg config.CollectorConfig) RewardDispatcherConfigMonitor {
	m := RewardDispatcherConfigMonitor{
		State:   &types.RewardDispatcherConfig{},
		metrics: make(map[Metric]float64),
		BaseMonitor: BaseMonitor{ContractAddress: cfg.RewardDispatcherConfigContract,
			apiClient: cfg.GetTerraClient(),
			logger:    cfg.Logger,
		},
	}

	return m
}


func (h RewardDispatcherConfigMonitor) Name() string {
	return "RewardDispatcherConfig"
}

func (h *RewardDispatcherConfigMonitor) InitMetrics() {
	h.metrics[RewardDispatcherConfigCRC32] = 0
}

func (h *RewardDispatcherConfigMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[RewardDispatcherConfigCRC32] = float64(crc32.ChecksumIEEE(data))
}

func (h *RewardDispatcherConfigMonitor) Handler(ctx context.Context) error {
	blunaReq, blunaResp := types.CommonConfigRequest{}, types.RewardDispatcherConfig{}

	err := makeStoreQuery(&blunaResp, blunaReq, ctx, h)
	if err != nil {
		return fmt.Errorf("failed to make %s request: %w", h.Name(), err)
	}
	h.State = &blunaResp
	h.updateMetrics()
	return nil
}

func (h RewardDispatcherConfigMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}