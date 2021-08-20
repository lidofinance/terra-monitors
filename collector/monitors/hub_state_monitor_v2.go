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
	StlunaBondedAmount MetricName = "stluna_bonded_amount"
	StlunaExchangeRate MetricName = "stluna_exchange_rate"
)

func NewHubStateMonitorV2(cfg config.CollectorConfig, logger *logrus.Logger) HubStateMonitorV2 {
	m := HubStateMonitorV2{
		State:      &types.HubStateResponseV2{},
		HubAddress: cfg.Addresses.HubContract,
		metrics:    make(map[MetricName]MetricValue),
		apiClient:  cfg.GetTerraClient(),
		logger:     logger,
	}
	m.InitMetrics()

	return m
}

type HubStateMonitorV2 struct {
	State      *types.HubStateResponseV2
	HubAddress string
	metrics    map[MetricName]MetricValue
	apiClient  *client.TerraLiteForTerra
	logger     *logrus.Logger
}

func (h HubStateMonitorV2) Name() string {
	return "HubState"
}

func (h *HubStateMonitorV2) InitMetrics() {
	h.setStringMetric(BlunaBondedAmount, "0")
	h.setStringMetric(BlunaExchangeRate, "0")
	h.setStringMetric(StlunaBondedAmount, "0")
	h.setStringMetric(StlunaExchangeRate, "0")
}

func (h *HubStateMonitorV2) updateMetrics() {
	h.setStringMetric(BlunaBondedAmount, h.State.TotalBondBlunaAmount)
	h.setStringMetric(BlunaExchangeRate, h.State.BlunaExchangeRate)
	h.setStringMetric(StlunaBondedAmount, h.State.TotalBondStlunaAmount)
	h.setStringMetric(StlunaExchangeRate, h.State.StlunaExchangeRate)
}

func (h *HubStateMonitorV2) Handler(ctx context.Context) error {
	hubReq, hubResp := types.GetHubStatePairV2()

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

func (h *HubStateMonitorV2) setStringMetric(m MetricName, rawValue string) {
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}
	if h.metrics[m] == nil {
		h.metrics[m] = &SimpleMetricValue{}
	}
	h.metrics[m].Set(v)
}

func (h HubStateMonitorV2) GetMetrics() map[MetricName]MetricValue {
	return h.metrics
}

func (h HubStateMonitorV2) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}
