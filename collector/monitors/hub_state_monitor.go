package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/client"
	"github.com/lidofinance/terra-monitors/client/wasm"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/sirupsen/logrus"
)

var (
	BlunaBondedAmount Metric = "bluna_bonded_amount"
	BlunaExchangeRate Metric = "bluna_exchange_rate"
)

func NewHubStateMintor(address string, apiClient *client.TerraLiteForTerra, logger *logrus.Logger) HubStateMonitor {
	m := HubStateMonitor{
		State:      &types.HubStateResponse{},
		HubAddress: address,
		apiClient:  apiClient,
	}

	if apiClient == nil {
		m.apiClient = client.NewHTTPClient(nil)
	}
	if logger == nil {
		m.logger = logging.NewDefaultLogger()
	}
	return m
}

type HubStateMonitor struct {
	State      *types.HubStateResponse
	HubAddress string
	apiClient  *client.TerraLiteForTerra
	logger     *logrus.Logger
}

func (h HubStateMonitor) Name() string {
	return "HubState"
}

func (h *HubStateMonitor) Handler(ctx context.Context) error {
	hubreq, hubresp := types.GetHubStatePair()

	reqRaw, err := json.Marshal(&hubreq)
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

	err = types.CastMapToStruct(resp.Payload.Result, &hubresp)
	if err != nil {
		return fmt.Errorf("failed to parse HubState body interface: %w", err)
	}

	h.logger.Infoln("updated HubState")
	h.State = &hubresp
	return nil
}

func (h HubStateMonitor) ProvidedMetrics() []Metric {
	return []Metric{
		BlunaExchangeRate,
		BlunaBondedAmount,
	}
}

func (h HubStateMonitor) Get(metric Metric) (float64, error) {
	switch metric {
	case BlunaBondedAmount:
		return strconv.ParseFloat(h.State.TotalBondAmount, 64)
	case BlunaExchangeRate:
		return strconv.ParseFloat(h.State.ExchangeRate, 64)
	}
	return 0, &MetricDoesNotExistError{metricName: metric}
}

func (h *HubStateMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *HubStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
