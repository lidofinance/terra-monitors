package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
	"hash/crc32"
	"strconv"
)

const (
	HubParametersEpochPeriod     Metric = "hub_parameters_epoch_period"
	HubParametersUnbondingPeriod Metric = "hub_parameters_unbonding_period"
	HubParametersPegRecoveryFee  Metric = "hub_parameters_peg_recovery_fee"
	HubParametersErThreshold     Metric = "hub_parameters_er_threshold"
	HubParametersCRC32           Metric = "hub_parameters_crc32"
)

type HubParametersMonitor struct {
	metrics         map[Metric]float64
	State           *types.HubParameters
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func NewHubParametersMonitor(cfg config.CollectorConfig) HubParametersMonitor {
	m := HubParametersMonitor{
		metrics:         make(map[Metric]float64),
		State:           &types.HubParameters{},
		ContractAddress: cfg.HubContract,
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}

	return m
}

func (h HubParametersMonitor) Name() string {
	return "HubParameters"
}

func (h *HubParametersMonitor) InitMetrics() {
	h.metrics[HubParametersCRC32] = 0
	h.metrics[HubParametersEpochPeriod] = 0
	h.metrics[HubParametersUnbondingPeriod] = 0
	h.metrics[HubParametersPegRecoveryFee] = 0
	h.metrics[HubParametersErThreshold] = 0
}

func (h *HubParametersMonitor) setStringMetric(m Metric, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h *HubParametersMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[HubParametersCRC32] = float64(crc32.ChecksumIEEE(data))
	h.metrics[HubParametersEpochPeriod] = float64(h.State.EpochPeriod)
	h.metrics[HubParametersUnbondingPeriod] = float64(h.State.UnbondingPeriod)
	h.setStringMetric(HubParametersPegRecoveryFee, h.State.PegRecoveryFee)
	h.setStringMetric(HubParametersErThreshold, h.State.ErThreshold)
}

func (h *HubParametersMonitor) Handler(ctx context.Context) error {
	hubReq, hubResp := types.HubParametersRequest{}, types.HubParameters{}

	reqRaw, err := json.Marshal(&hubReq)
	if err != nil {
		return fmt.Errorf("failed to marshal HubParameters request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(h.ContractAddress)
	p.SetQueryMsg(string(reqRaw))

	resp, err := h.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process HubParameters request: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &hubResp)
	if err != nil {
		return fmt.Errorf("failed to parse HubParameters body interface: %w", err)
	}

	h.logger.Infoln("updated HubParameters")
	h.State = &hubResp
	h.updateMetrics()
	return nil
}

func (h HubParametersMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}
