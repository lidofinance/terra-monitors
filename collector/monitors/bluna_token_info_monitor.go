package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	BlunaTotalSupply MetricName = "bluna_total_supply"
)

func NewBlunaTokenInfoMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *BlunaTokenInfoMonitor {
	m := BlunaTokenInfoMonitor{
		State:           &types.TokenInfoResponse{},
		ContractAddress: cfg.Addresses.BlunaTokenInfoContract,
		metrics:         make(map[MetricName]MetricValue),
		apiClient:       client.New(cfg.LCD, logger),
		logger:          logger,
		lock:            sync.RWMutex{},
	}
	m.InitMetrics()
	return &m
}

type BlunaTokenInfoMonitor struct {
	State           *types.TokenInfoResponse
	ContractAddress string
	metrics         map[MetricName]MetricValue
	apiClient       *terraClient.TerraLiteForTerra
	logger          *logrus.Logger
	lock            sync.RWMutex
}

func (h *BlunaTokenInfoMonitor) Name() string {
	return "BlunaTokenInfo"
}

func (h *BlunaTokenInfoMonitor) InitMetrics() {
	h.setStringMetric(BlunaTotalSupply, "0")
}

func (h *BlunaTokenInfoMonitor) updateMetrics() {
	h.lock.Lock()
	defer h.lock.Unlock()
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

func (h *BlunaTokenInfoMonitor) setStringMetric(m MetricName, rawValue string) {
	v, err := cosmostypes.NewDecFromStr(rawValue)
	if err != nil {
		h.logger.Errorf("failed to set value \"%s\" to metric \"%s\": %+v\n", rawValue, m, err)
	}

	value, err := v.Float64()
	if err != nil {
		h.logger.Errorf("failed to get float64 value from string \"%s\" for metric \"%s\": %+v\n", rawValue, m, err)
	}

	if h.metrics[m] == nil {
		h.metrics[m] = &SimpleMetricValue{}
	}

	h.metrics[m].Set(value)
}

func (h *BlunaTokenInfoMonitor) GetMetrics() map[MetricName]MetricValue {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.metrics
}

func (h *BlunaTokenInfoMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (h *BlunaTokenInfoMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
