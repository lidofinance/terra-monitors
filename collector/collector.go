package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-openapi/runtime"
	"github.com/lidofinance/terra-monitors/client"
	"github.com/lidofinance/terra-monitors/client/wasm"
	"github.com/sirupsen/logrus"
)

type Metrics string

var (
	BlunaBondedAmount Metrics = "bluna_bonded_amount"
	BlunaTotalSupply  Metrics = "bluna_total_supply"
	GlobalIndex       Metrics = "global_index"
)

// type

func ParseRequestBody(body io.ReadCloser, ret interface{}) error {
	m := struct {
		Result interface{} `json:"result"`
		Error  string      `json:"error"`
	}{
		Result: ret,
	}
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if m.Error != "" {
		return fmt.Errorf(m.Error)
	}
	return nil
}

type Collector interface {
	Get(metric Metrics) (float64, error)
	UpdateData(ctx context.Context) error
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewLCDCollector(RewardContractAddress string, logger *logrus.Logger) LCDCollector {
	return LCDCollector{
		logger:                logger,
		rewardContractAddress: RewardContractAddress,
		apiClient:             client.NewHTTPClient(nil),
	}
}

type LCDCollector struct {
	logger                *logrus.Logger
	LCDEndpoint           string
	HubAddress            string
	rewardContractAddress string
	BlunaContractAddress  string
	CollectedData         CollectedData
	apiClient             *client.TerraLiteForTerra
}

func (c *LCDCollector) SetTransport(transport runtime.ClientTransport) {
	c.apiClient.SetTransport(transport)
}

func (c LCDCollector) Get(metric Metrics) (float64, error) {
	switch metric {
	case BlunaBondedAmount:
		return strconv.ParseFloat(c.CollectedData.HubState.TotalBondAmount, 64)
	case GlobalIndex:
		return strconv.ParseFloat(c.CollectedData.RewardState.GlobalIndex, 64)
	case BlunaTotalSupply:
		return strconv.ParseFloat(c.CollectedData.BlunaTokenInfo.TotalSupply, 64)
	}
	return 0, fmt.Errorf("metric \"%s\" not found", metric)
}

func (c *LCDCollector) getRewardState(ctx context.Context) (*RewardStateResponse, error) {
	stateReq, stateResp := GetRewardResponseStatePair()
	reqRaw, err := json.Marshal(&stateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(c.rewardContractAddress)
	p.SetQueryMsg(string(reqRaw))
	_, err = c.apiClient.Wasm.GetWasmContractsContractAddressStore(&p,
		//custom reader unmarshals json data to Go struct
		func(op *runtime.ClientOperation) {
			op.Reader = &PayloadExtractor{PayloadResult: &stateResp}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to process RewardState request: %w", err)
	}
	c.logger.Infoln("updated reward state info")
	return &stateResp, nil
}

func (c *LCDCollector) getBlunaTokenInfo(ctx context.Context) (*TokenInfoResponse, error) {
	tokreq, tokresp := GetCommonTokenInfoPair()
	reqRaw, err := json.Marshal(&tokreq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(c.BlunaContractAddress)
	p.SetQueryMsg(string(reqRaw))
	_, err = c.apiClient.Wasm.GetWasmContractsContractAddressStore(&p,
		//custom reader unmarshals json data to Go struct
		func(op *runtime.ClientOperation) {
			op.Reader = &PayloadExtractor{PayloadResult: &tokresp}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to process BlunaTokenInfo request: %w", err)
	}
	c.logger.Infoln("updated bluna token info")
	return &tokresp, nil
}

func (c *LCDCollector) UpdateData(ctx context.Context) error {
	blunaTokenInfo, err := c.getBlunaTokenInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bluna token info: %w", err)
	}
	c.CollectedData.BlunaTokenInfo = *blunaTokenInfo

	rewardState, err := c.getRewardState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reward state: %w", err)
	}
	c.CollectedData.RewardState = *rewardState
	return nil
}
