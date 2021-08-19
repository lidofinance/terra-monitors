package types

import (
	"encoding/json"
	"fmt"
)

func CastMapToStruct(m interface{}, ret interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal body interface{}: %w", err)
	}

	err = json.Unmarshal(data, ret)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return nil
}

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

//should be corrected after contract migration according to the new response schema

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

type HubWhitelistedValidatorsRequest struct {
	WhitelistedValidators struct{} `json:"whitelisted_validators"`
}

type HubWhitelistedValidatorsResponse struct {
	Validators []string `json:"validators"`
}

func GetHubWhitelistedValidatorsPair() (HubWhitelistedValidatorsRequest, HubWhitelistedValidatorsResponse) {
	return HubWhitelistedValidatorsRequest{}, HubWhitelistedValidatorsResponse{}
}

type ValidatorRegistryValidatorsRequest struct {
	WhitelistedValidators struct{} `json:"get_validators_for_delegation"`
}

type Validator struct {
	Address         string `json:"address"`
	TotalDelegated  string `json:"total_delegated"`  //Uint128
}

type ValidatorRegistryValidatorsResponse = []Validator

func GetValidatorRegistryValidatorsPair() (ValidatorRegistryValidatorsRequest, ValidatorRegistryValidatorsResponse) {
	return ValidatorRegistryValidatorsRequest{}, ValidatorRegistryValidatorsResponse{}
}

type HubConfig struct {
	Creator                    string `json:"creator"`
	RewardDispatcherContract   string `json:"reward_dispatcher_contract"`
	ValidatorsRegistryContract string `json:"validators_registry_contract"`
	BlunaTokenContract         string `json:"bluna_token_contract"`
	StlunaTokenContract        string `json:"stluna_token_contract"`
	AirdropRegistryContract    string `json:"airdrop_registry_contract"`
}

type CommonConfigRequest struct {
	Config struct{} `json:"config"`
}

type HubParameters struct {
	EpochPeriod         uint64 `json:"epoch_period"`
	UnderlyingCoinDenom string `json:"underlying_coin_denom"`
	UnbondingPeriod     uint64 `json:"unbonding_period"`
	PegRecoveryFee      string `json:"peg_recovery_fee"` //Decimal128 as string
	ErThreshold         string `json:"er_threshold"`     //Decimal128 as string
	RewardDenom         string `json:"reward_denom"`
}

type HubParametersRequest struct {
	Parameters struct{} `json:"parameters"`
}

type BlunaRewardConfig struct {
	HubContract string `json:"hub_contract"`
	RewardDenom string `json:"reward_denom"`
}

type RewardDispatcherConfig struct {
	Owner               string `json:"owner"`
	HubContract         string `json:"hub_contract"`
	BlunaRewardContract string `json:"bluna_reward_contract"`
	StlunaRewardDenom   string `json:"stluna_reward_denom"`
	BlunaRewardDenom    string `json:"bluna_reward_denom"`
	LidoFeeAddress      string `json:"lido_fee_address"`
	LidoFeeRate         string `json:"lido_fee_rate"` //decimal128
}

type ValidatorsRegistryConfig struct {
	Owner       string `json:"owner"`
	HubContract string `json:"hub_contract"`
}

type AirDropRegistryConfig struct {
	Owner        string   `json:"owner"`
	HubContract  string   `json:"hub_contract"`
	AirDropToken []string `json:"airdrop_tokens"`
}

type ValidatorInfo struct {
	Address        string
	PubKey         string
	Moniker        string
	CommissionRate float64
}
