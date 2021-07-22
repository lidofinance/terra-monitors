package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	BlunaTotalSupply Metric = "bluna_total_supply"
)

func NewBlunaTokenInfoMintor(address string, apiClient *client.TerraLiteForTerra, logger *logrus.Logger) BlunaTokenInfoMonitor {
	m := BlunaTokenInfoMonitor{
		State:           &types.TokenInfoResponse{},
		ContractAddress: address,
		apiClient:       apiClient,
		logger:          logger,
	}
	if apiClient == nil {
		m.apiClient = client.NewHTTPClient(nil)
	}
	if logger == nil {
		m.logger = logging.NewDefaultLogger()
	}
	return m
}

type BlunaTokenInfoMonitor struct {
	State           *types.TokenInfoResponse
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h BlunaTokenInfoMonitor) Name() string {
	return "BlunaTokenInfo"
}

func (h *BlunaTokenInfoMonitor) Handler(ctx context.Context) error {
	rewardreq, rewardresp := types.GetCommonTokenInfoPair()

	reqRaw, err := json.Marshal(&rewardreq)
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

	err = types.CastMapToStruct(resp.Payload.Result, &rewardresp)
	if err != nil {
		return fmt.Errorf("failed to parse BlunaTokenInfo body interface: %w", err)
	}

	h.logger.Infoln("updated BlunaTokenInfo")
	h.State = &rewardresp
	return nil
}

func (h BlunaTokenInfoMonitor) ProvidedMetrics() []Metric {
	return []Metric{
		BlunaTotalSupply,
	}
}

func (h BlunaTokenInfoMonitor) Get(metric Metric) (float64, error) {
	switch metric {
	case BlunaTotalSupply:
		return strconv.ParseFloat(h.State.TotalSupply, 64)
	}
	return 0, &MetricDoesNotExistError{metricName: metric}
}

func (h *BlunaTokenInfoMonitor) SetApiClient(client *client.TerraLiteForTerra) {
	h.apiClient = client
}

func (h *BlunaTokenInfoMonitor) SetLogger(l *logrus.Logger) {
	h.logger = l
}
