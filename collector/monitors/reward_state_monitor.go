package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	GlobalIndex Metric = "global_index"
)

func NewRewardStateMintor(address string, apiClient *client.TerraLiteForTerra, logger *logrus.Logger) RewardStateMonitor {
	m := RewardStateMonitor{
		State:           &types.RewardStateResponse{},
		ContractAddress: address,
		apiClient:       apiClient,
		logger:          logger,
	}

	if apiClient == nil {
		m.apiClient = client.NewHTTPClient(nil)
	}
	if logger == nil {
		m.logger = logging.NewDefaultLogger()
	}
	return m
}

type RewardStateMonitor struct {
	State           *types.RewardStateResponse
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h RewardStateMonitor) Name() string {
	return "RewardState"
}

func (h *RewardStateMonitor) Handler(ctx context.Context) error {
	rewardreq, rewardresp := types.GetRewardStatePair()

	reqRaw, err := json.Marshal(&rewardreq)
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

	err = types.CastMapToStruct(resp.Payload.Result, &rewardresp)
	if err != nil {
		return fmt.Errorf("failed to parse RewardState body interface: %w", err)
	}

	h.logger.Infoln("updated RewardState")
	h.State = &rewardresp
	return nil
}

func (h RewardStateMonitor) ProvidedMetrics() []Metric {
	return []Metric{
		GlobalIndex,
	}
}

func (h RewardStateMonitor) Get(metric Metric) (float64, error) {
	switch metric {
	case GlobalIndex:
		return strconv.ParseFloat(h.State.GlobalIndex, 64)
	}
	return 0, &MetricDoesNotExistError{metricName: metric}
}

func (h *RewardStateMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *RewardStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
