package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	BlunaBondedAmount MetricName = "bluna_bonded_amount"
	BlunaExchangeRate MetricName = "bluna_exchange_rate"
)

func NewHubStateMonitor(cfg config.CollectorConfig, logger *logrus.Logger) Monitor {

	switch cfg.BassetContractsVersion {
	case config.V1Contracts:
		{
			m1 := HubStateMonitor{
				State:      &types.HubStateResponseV1{},
				HubAddress: cfg.Addresses.HubContract,
				metrics:    make(map[MetricName]MetricValue),
				apiClient:  client.New(cfg.LCD, logger),
				logger:     logger,
			}
			m1.InitMetrics()
			return &m1
		}
	case config.V2Contracts:
		{
			m2 := HubStateMonitorV2{
				State:      &types.HubStateResponseV2{},
				HubAddress: cfg.Addresses.HubContract,
				metrics:    make(map[MetricName]MetricValue),
				apiClient:  client.New(cfg.LCD, logger),
				logger:     logger,
			}
			m2.InitMetrics()
			return &m2
		}
	default:
		panic("unknown contracts version")
	}
}

type HubStateMonitor struct {
	State      *types.HubStateResponseV1
	HubAddress string
	metrics    map[MetricName]MetricValue
	apiClient  *terraClient.TerraLiteForTerra
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
	hubReq, hubResp := types.GetHubStatePairV1()

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
	if h.metrics[m] == nil {
		h.metrics[m] = &SimpleMetricValue{}
	}
	h.metrics[m].Set(v)
}

func (h HubStateMonitor) GetMetrics() map[MetricName]MetricValue {
	return h.metrics
}

func (h HubStateMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (h *HubStateMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
