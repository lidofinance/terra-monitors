package monitors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/wasm"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

var (
	GlobalIndex MetricName = "global_index"
)

func NewRewardStateMonitor(cfg config.CollectorConfig, logger *logrus.Logger) RewardStateMonitor {
	m := RewardStateMonitor{
		State:           &types.RewardStateResponse{},
		ContractAddress: cfg.Addresses.RewardContract,
		metrics:         make(map[MetricName]MetricValue),
		apiClient:       utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:          logger,
	}
	m.InitMetrics()

	return m
}

type RewardStateMonitor struct {
	State           *types.RewardStateResponse
	ContractAddress string
	metrics         map[MetricName]MetricValue
	apiClient       *client.TerraRESTApis
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

func (h *RewardStateMonitor) setStringMetric(m MetricName, rawValue string) {
	v, err := cosmostypes.NewDecFromStr(rawValue)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	if h.metrics[m] == nil {
		h.metrics[m] = &SimpleMetricValue{}
	}

	value, err := v.Float64()
	if err != nil {
		h.logger.Errorf("failed to get float64 value from string \"%s\" for metric \"%s\": %+v\n", rawValue, m, err)
	}

	h.metrics[m].Set(value)
}

func (h RewardStateMonitor) GetMetrics() map[MetricName]MetricValue {
	return h.metrics
}

func (h RewardStateMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (h *RewardStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
