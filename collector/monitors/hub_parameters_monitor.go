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
	HubParametersEpochPeriod     MetricName = "hub_parameters_epoch_period"
	HubParametersUnbondingPeriod MetricName = "hub_parameters_unbonding_period"
	HubParametersPegRecoveryFee  MetricName = "hub_parameters_peg_recovery_fee"
	HubParametersErThreshold     MetricName = "hub_parameters_er_threshold"
	HubParametersCRC32           MetricName = "hub_parameters_crc32"
)

type HubParametersMonitor struct {
	metrics         map[MetricName]MetricValue
	State           *types.HubParameters
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func NewHubParametersMonitor(cfg config.CollectorConfig) HubParametersMonitor {
	m := HubParametersMonitor{
		metrics:         make(map[MetricName]MetricValue),
		State:           &types.HubParameters{},
		ContractAddress: cfg.HubContract,
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}
	m.InitMetrics()

	return m
}

func (h HubParametersMonitor) Name() string {
	return "HubParameters"
}
func (h *HubParametersMonitor) providedMetrics() []MetricName {
	return []MetricName{
		HubParametersCRC32,
		HubParametersEpochPeriod,
		HubParametersUnbondingPeriod,
		HubParametersPegRecoveryFee,
		HubParametersErThreshold,
	}
}

func (h *HubParametersMonitor) InitMetrics() {
	for _, metric := range h.providedMetrics() {
		if h.metrics[metric] == nil {
			h.metrics[metric] = &SimpleMetricValue{}
		}
		h.metrics[metric].Set(0)
	}
}

func (h *HubParametersMonitor) setStringMetric(m MetricName, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	if h.metrics[m] == nil {
		h.metrics[m] = &SimpleMetricValue{}
	}
	h.metrics[m].Set(v)
}

func (h *HubParametersMonitor) updateMetrics() {
	data, err := json.Marshal(h.State)
	if err != nil {
		h.logger.Errorf("failed to marshal %s: %s", h.Name(), err)
	}
	h.metrics[HubParametersCRC32].Set(float64(crc32.ChecksumIEEE(data)))
	h.metrics[HubParametersEpochPeriod].Set(float64(h.State.EpochPeriod))
	h.metrics[HubParametersUnbondingPeriod].Set(float64(h.State.UnbondingPeriod))
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

func (h HubParametersMonitor) GetMetrics() map[MetricName]MetricValue {
	return h.metrics
}

func (h HubParametersMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}
