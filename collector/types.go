package collector

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/lidofinance/terra-monitors/client/wasm"
)

func GetRewardStatePair() (CommonStateRequest, RewardStateResponse) {
	return CommonStateRequest{}, RewardStateResponse{}
}

type RewardStateResponse struct {
	GlobalIndex       string `json:"global_index"`        //decimal128
	PrevRewardBalance string `json:"prev_reward_balance"` //uint128
	TotalBalance      string `json:"total_balance"`       //uint128
}

type CommonStateRequest struct {
	State struct{} `json:"state"`
}

func GetCommonTokenInfoPair() (TokenInfoRequest, TokenInfoResponse) {
	return TokenInfoRequest{}, TokenInfoResponse{}
}

type TokenInfoRequest struct {
	TokenInfo struct{} `json:"token_info"`
}

type TokenInfoResponse struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply string `json:"total_supply"` //uint128
}

//current response type
// {"exchange_rate":"1.000007185273335589","total_bond_amount":"79192023592216","last_index_modification":1626348317,"prev_hub_balance":"319234626534","actual_unbonded_amount":"0","last_unbonded_time":1626198926,"last_processed_batch":32}

//next gen response type
// pub struct StateResponse {
//     pub bluna_exchange_rate: Decimal,
//     pub stluna_exchange_rate: Decimal,
//     pub total_bond_bluna_amount: Uint128,
//     pub total_bond_stluna_amount: Uint128,
//     pub last_index_modification: u64,
//     pub prev_hub_balance: Uint128,
//     pub actual_unbonded_amount: Uint128,
//     pub last_unbonded_time: u64,
//     pub last_processed_batch: u64,
// }

//shoud be corrtced after contract migration according new response schema
type HubStateResponse struct {
	ExchangeRate          string `json:"exchange_rate"`     //decimal
	TotalBondAmount       string `json:"total_bond_amount"` //uint128
	LastIndexModification uint64 `json:"last_index_modification"`
	PrevHubBalance        string `json:"prev_hub_balance"`       //uint128
	ActualUnbondedAmount  string `json:"actual_unbonded_amount"` //uint128
	LastUnbondedTime      uint64 `json:"last_unbonded_time"`
	LastProcessedBatch    uint64 `json:"last_processed_batch"`
}

func GetHubStatePair() (CommonStateRequest, HubStateResponse) {
	return CommonStateRequest{}, HubStateResponse{}
}

type CollectedData struct {
	HubState       HubStateResponse
	RewardState    RewardStateResponse
	BlunaTokenInfo TokenInfoResponse
}

type PayloadExtractor struct {
	PayloadResult interface{}
}

// ReadResponse reads a server response into the received o.
func (o *PayloadExtractor) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := wasm.NewGetWasmContractsContractAddressStoreOK()
		result.Payload = &wasm.GetWasmContractsContractAddressStoreOKBody{}
		err := ParseRequestBody(response.Body(), &o.PayloadResult)
		result.Payload.Result = o.PayloadResult
		if err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

type MetricDoesNotExistsError struct {
	metricName Metrics
}

func (m *MetricDoesNotExistsError) Error() string {
	return fmt.Sprintf("metric \"%s\" does not exists on monitor", m.metricName)
}
