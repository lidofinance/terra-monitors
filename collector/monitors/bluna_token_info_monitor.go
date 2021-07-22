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
	BlunaTotalSupply Metric = "bluna_total_supply"
)

func NewBlunaTokenInfoMintor(cfg config.CollectorConfig) BlunaTokenInfoMonitor {
	m := BlunaTokenInfoMonitor{
		metrics:         make(map[Metric]float64),
		State:           &types.TokenInfoResponse{},
		ContractAddress: cfg.BlunaTokenInfoContract,
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}
	return m
}

type BlunaTokenInfoMonitor struct {
	metrics         map[Metric]float64
	State           *types.TokenInfoResponse
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h BlunaTokenInfoMonitor) Name() string {
	return "BlunaTokenInfo"
}

func (h *BlunaTokenInfoMonitor) InitMetrics() {
	h.setStringMetric(BlunaTotalSupply, "0")
}

func (h *BlunaTokenInfoMonitor) updateMetrics() {
	h.setStringMetric(BlunaTotalSupply, h.State.TotalSupply)
}

func (h *BlunaTokenInfoMonitor) Handler(ctx context.Context) error {
	rewardReq, rewardResp := types.GetCommonTokenInfoPair()

	reqRaw, err := json.Marshal(&rewardReq)
	if err != nil {
		return fmt.Errorf("failed to marshal BlunaTokenInfo request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(h.ContractAddress)
	p.SetQueryMsg(string(reqRaw))

	resp, err := h.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process BlunaTokenInfo request: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &rewardResp)
	if err != nil {
		return fmt.Errorf("failed to parse BlunaTokenInfo body interface: %w", err)
	}

	h.logger.Infoln("updated BlunaTokenInfo")
	h.State = &rewardResp
	h.updateMetrics()
	return nil
}

func (h *BlunaTokenInfoMonitor) setStringMetric(m Metric, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h BlunaTokenInfoMonitor) GetMetrics() map[Metric]float64 {
	return h.metrics
}

func (h *BlunaTokenInfoMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *BlunaTokenInfoMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
