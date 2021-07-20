package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/client"
	"github.com/lidofinance/terra-monitors/client/wasm"
	"github.com/sirupsen/logrus"
)

var (
	BlunaTotalSupply Metrics = "bluna_total_supply"
)

func NewBlunaTokenInfoMintor(address string) BlunaTokenInfoMonitor {
	return BlunaTokenInfoMonitor{
		State:           &TokenInfoResponse{},
		ContractAddress: address,
	}
}

type BlunaTokenInfoMonitor struct {
	State           *TokenInfoResponse
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h BlunaTokenInfoMonitor) Name() string {
	return "BlunaTokenInfo"
}

func (h *BlunaTokenInfoMonitor) Handler(ctx context.Context) error {
	rewardreq, rewardresp := GetCommonTokenInfoPair()
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
	err = ParseRequestBody(resp.Payload.Result, &rewardresp)
	if err != nil {
		return fmt.Errorf("failed to parse BlunaTokenInfo body interface: %w", err)
	}

	h.logger.Infoln("updated BlunaTokenInfo")
	h.State = &rewardresp
	return nil
}

func (h BlunaTokenInfoMonitor) ProvidedMetrics() []Metrics {
	return []Metrics{
		BlunaTotalSupply,
	}
}

func (h BlunaTokenInfoMonitor) Get(metric Metrics) (float64, error) {
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
