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
	BlunaBondedAmount MetricName = "bluna_bonded_amount"
	BlunaExchangeRate MetricName = "bluna_exchange_rate"
)

func NewHubStateMonitor(cfg config.CollectorConfig) HubStateMonitor {
	m := HubStateMonitor{
		State:      &types.HubStateResponse{},
		HubAddress: cfg.HubContract,
		metrics:    make(map[MetricName]float64),
		apiClient:  cfg.GetTerraClient(),
		logger:     cfg.Logger,
	}

	return m
}

type HubStateMonitor struct {
	State      *types.HubStateResponse
	HubAddress string
	metrics    map[MetricName]float64
	apiClient  *client.TerraLiteForTerra
	logger     *logrus.Logger
}

func (h HubStateMonitor) Name() string {
	return "HubState"
}

func (h *HubStateMonitor) InitMetrics() {
	h.setStringMetric(BlunaBondedAmount, "0")
	h.setStringMetric(BlunaExchangeRate, "0")
}

func (h *HubStateMonitor) updateMetrics() {
	h.setStringMetric(BlunaBondedAmount, h.State.TotalBondAmount)
	h.setStringMetric(BlunaExchangeRate, h.State.ExchangeRate)
}

func (h *HubStateMonitor) Handler(ctx context.Context) error {
	hubReq, hubResp := types.GetHubStatePair()

	reqRaw, err := json.Marshal(&hubReq)
	if err != nil {
		return fmt.Errorf("failed to marshal HubState request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(h.HubAddress)
	p.SetQueryMsg(string(reqRaw))

	resp, err := h.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process HubState request: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &hubResp)
	if err != nil {
		return fmt.Errorf("failed to parse HubState body interface: %w", err)
	}

	h.logger.Infoln("updated HubState")
	h.State = &hubResp
	h.updateMetrics()
	return nil
}

func (h *HubStateMonitor) setStringMetric(m MetricName, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	h.metrics[m] = v
}

func (h HubStateMonitor) GetMetrics() map[MetricName]float64 {
	return h.metrics
}

func (h HubStateMonitor) GetMetricVectors() map[MetricName]MetricVector {
	return nil
}

func (h *HubStateMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *HubStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
