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
	StlunaTotalSupply MetricName = "stluna_total_supply"
)

func NewStlunaTokenInfoMonitor(cfg config.CollectorConfig) StlunaTokenInfoMonitor {
	m := StlunaTokenInfoMonitor{
		State:           &types.TokenInfoResponse{},
		ContractAddress: cfg.StlunaTokenInfoContract,
		metrics:         make(map[MetricName]float64),
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}
	return m
}

type StlunaTokenInfoMonitor struct {
	State           *types.TokenInfoResponse
	ContractAddress string
	metrics         map[MetricName]float64
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h StlunaTokenInfoMonitor) Name() string {
	return "StlunaTokenInfo"
}

func (h *StlunaTokenInfoMonitor) InitMetrics() {
	h.setStringMetric(StlunaTotalSupply, "0")
}

func (h *StlunaTokenInfoMonitor) updateMetrics() {
	h.setStringMetric(StlunaTotalSupply, h.State.TotalSupply)
}

func (h *StlunaTokenInfoMonitor) Handler(ctx context.Context) error {
	rewardReq, rewardResp := types.GetCommonTokenInfoPair()

	reqRaw, err := json.Marshal(&rewardReq)
	if err != nil {
		return fmt.Errorf("failed to marshal StlunaTokenInfo request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(h.ContractAddress)
	p.SetQueryMsg(string(reqRaw))

	resp, err := h.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process StlunaTokenInfo request: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &rewardResp)
	if err != nil {
		return fmt.Errorf("failed to parse StlunaTokenInfo body interface: %w", err)
	}

	h.logger.Infoln("updated StlunaTokenInfo")
	h.State = &rewardResp
	h.updateMetrics()
	return nil
}

func (h *StlunaTokenInfoMonitor) setStringMetric(m MetricName, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h StlunaTokenInfoMonitor) GetMetrics() map[MetricName]float64 {
	return h.metrics
}

func (h StlunaTokenInfoMonitor) GetMetricVectors() map[MetricName]MetricVector {
	return nil
}

func (h *StlunaTokenInfoMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *StlunaTokenInfoMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
