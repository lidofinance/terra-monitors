package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	GlobalIndex Metric = "global_index"
)

func NewRewardStateMonitor(cfg config.CollectorConfig) RewardStateMonitor {
	m := RewardStateMonitor{
		State:           &types.RewardStateResponse{},
		ContractAddress: cfg.RewardContract,
		metrics:         make(map[Metric]float64),
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}

	return m
}

type RewardStateMonitor struct {
	State           *types.RewardStateResponse
	ContractAddress string
	metrics         map[Metric]float64
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h RewardStateMonitor) Name() string {
	return "RewardState"
}

func (h *RewardStateMonitor) InitMetrics() {
	h.setStringMetric(GlobalIndex, "0")
}

func (h *RewardStateMonitor) updateMetrics() {
	h.setStringMetric(GlobalIndex, h.State.GlobalIndex)
}

func (h *RewardStateMonitor) Handler(ctx context.Context) error {
	rewardReq, rewardResp := types.GetRewardStatePair()

	reqRaw, err := json.Marshal(&rewardReq)
	if err != nil {
		return fmt.Errorf("failed to marshal RewardState request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(h.ContractAddress)
	p.SetQueryMsg(string(reqRaw))

	resp, err := h.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process RewardState request: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &rewardResp)
	if err != nil {
		return fmt.Errorf("failed to parse RewardState body interface: %w", err)
	}

	h.logger.Infoln("updated RewardState")
	h.State = &rewardResp
	h.updateMetrics()
	return nil
}

func (h *RewardStateMonitor) setStringMetric(m Metric, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h RewardStateMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}

func (h *RewardStateMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *RewardStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
