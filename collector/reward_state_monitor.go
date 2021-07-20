package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/client"
	"github.com/lidofinance/terra-monitors/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	GlobalIndex Metrics = "global_index"
)

func NewRewardStateMintor(address string) RewardStateMonitor {
	return RewardStateMonitor{
		State:           &RewardStateResponse{},
		ContractAddress: address,
	}
}

type RewardStateMonitor struct {
	State           *RewardStateResponse
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h RewardStateMonitor) Name() string {
	return "RewardState"
}

func (h *RewardStateMonitor) Handler(ctx context.Context) error {
	rewardreq, rewardresp := GetRewardStatePair()
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
	err = ParseRequestBody(resp.Payload.Result, &rewardresp)
	if err != nil {
		return fmt.Errorf("failed to parse RewardState body interface: %w", err)
	}

	h.logger.Infoln("updated RewardState")
	h.State = &rewardresp
	return nil
}

func (h RewardStateMonitor) ProvidedMetrics() []Metrics {
	return []Metrics{
		GlobalIndex,
	}
}

func (h RewardStateMonitor) Get(metric Metrics) (float64, error) {
	switch metric {
	case GlobalIndex:
		return strconv.ParseFloat(h.State.GlobalIndex, 64)
	}
	return 0, &MetricDoesNotExistsError{metricName: metric}
}

func (h *RewardStateMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *RewardStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
