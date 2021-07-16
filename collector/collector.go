package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Metrics string

var (
	BlunaBondedAmount Metrics = "bluna_bonded_amount"
	BlunaTotalSupply  Metrics = "bluna_total_supply"
	GlobalIndex       Metrics = "global_index"
)

func parseRequestBody(resp *http.Response, ret interface{}) error {
	m := struct {
		Result interface{} `json:"result"`
		Error  string      `json:"error"`
	}{
		Result: ret,
	}
	decoder := json.NewDecoder(resp.Body)
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

func NewLCDCollector(endpoint string, RewardContractAddress string, logger *logrus.Logger) LCDCollector {
	return LCDCollector{
		logger:                logger,
		LCDEndpoint:           endpoint,
		RewardContractAddress: RewardContractAddress,
		HttpClient:            http.DefaultClient,
	}
}

type LCDCollector struct {
	logger                *logrus.Logger
	LCDEndpoint           string
	HubAddress            string
	RewardContractAddress string
	BlunaContractAddress  string
	CollectedData         CollectedData
	HttpClient            HttpClient
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

func (c LCDCollector) buildRequest(ctx context.Context, contractAddress string, query interface{}) (*http.Request, error) {
	url := fmt.Sprintf("%s/wasm/contracts/%s/store", c.LCDEndpoint, contractAddress)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prapare request for %+v query: %w", query, err)
	}
	q := req.URL.Query()
	queryRaw, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %+v query: %w", query, err)
	}
	q.Add("query_msg", string(queryRaw))
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (c LCDCollector) processRequest(req *http.Request, ret interface{}) error {
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	err = parseRequestBody(resp, ret)
	if err != nil {
		return fmt.Errorf("failed to process response body: %w", err)
	}
	return nil
}

func (c LCDCollector) buildAndProcessRequest(
	ctx context.Context,
	contractAddress string,
	query interface{},
	ret interface{},
) error {
	req, err := c.buildRequest(ctx, contractAddress, query)
	if err != nil {
		return fmt.Errorf(
			"failed to build request %+v for %s contract: %w",
			query,
			contractAddress,
			err,
		)
	}
	err = c.processRequest(req, ret)
	if err != nil {
		return fmt.Errorf("failed to process %+v request for %s contract: %w", query, contractAddress, err)
	}
	return nil
}

func (c *LCDCollector) getRewardState(ctx context.Context) (*RewardStateResponse, error) {
	req, resp := GetRewardResponseStatePair()
	err := c.buildAndProcessRequest(ctx, c.RewardContractAddress, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to process RewardState request: %w", err)
	}
	c.logger.Infoln("updated reward state info")
	return &resp, nil
}

func (c *LCDCollector) getBlunaTokenInfo(ctx context.Context) (*TokenInfoResponse, error) {
	req, resp := GetCommonTokenInfoPair()
	err := c.buildAndProcessRequest(ctx, c.BlunaContractAddress, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to process BlunaTokenInfo request: %w", err)
	}
	c.logger.Infoln("updated bluna token info")
	return &resp, nil
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
