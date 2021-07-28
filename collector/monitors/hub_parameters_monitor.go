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
	HubParametersEpochPeriod Metric = "hub_parameters_epoch_period"
	//HubParametersUnderlyingCoinDenomCRC32 crc32 checksum of CoinDenom string to detect string changes in grafana alerting
	HubParametersUnderlyingCoinDenomCRC32 Metric = "hub_parameters_underlying_coin_denom_crc32"
	HubParametersUnbondingPeriod          Metric = "hub_parameters_unbonding_period"
	HubParametersPegRecoveryFee           Metric = "hub_parameters_peg_recovery_fee"
	HubParametersErThreshold              Metric = "hub_parameters_er_threshold"
	//HubParametersRewardDenomCRC32 crc32 checksum of RewardDenom string to detect string changes in grafana alerting
	HubParametersRewardDenomCRC32 Metric = "hub_parameters_reward_denom_crc32"
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
	h.metrics[HubParametersEpochPeriod] = 0
	h.metrics[HubParametersUnderlyingCoinDenomCRC32] = 0
	h.metrics[HubParametersUnbondingPeriod] = 0
	h.metrics[HubParametersPegRecoveryFee] = 0
	h.metrics[HubParametersErThreshold] = 0
	h.metrics[HubParametersRewardDenomCRC32] = 0
}

func (h *HubParametersMonitor) setStringMetric(m Metric, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h *HubParametersMonitor) updateMetrics() {
	h.metrics[HubParametersEpochPeriod] = float64(h.State.EpochPeriod)
	h.metrics[HubParametersUnderlyingCoinDenomCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.UnderlyingCoinDenom)))
	h.metrics[HubParametersUnbondingPeriod] = float64(h.State.UnbondingPeriod)
	h.setStringMetric(HubParametersPegRecoveryFee, h.State.PegRecoveryFee)
	h.setStringMetric(HubParametersErThreshold, h.State.ErThreshold)
	h.metrics[HubParametersRewardDenomCRC32] = float64(crc32.ChecksumIEEE([]byte(h.State.RewardDenom)))
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


