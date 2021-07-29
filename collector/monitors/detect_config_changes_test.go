package monitors

import (
	"context"
	"encoding/json"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/stretchr/testify/suite"
)

type DetectorTestSuite struct {
	suite.Suite
}

func (suite *DetectorTestSuite) SetupTest() {

}

func (suite *DetectorTestSuite) TestBlunaRewardConfig() {
	rewardConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.BlunaRewardConfig{
			HubContract: "test1",
			RewardDenom: "uusd",
		},
	}
	responseData, err := json.Marshal(rewardConfigFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	blunaRewardConfigMonitor := NewBlunaRewardConfigMonitor(cfg)

	err = blunaRewardConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal("uusd", blunaRewardConfigMonitor.State.RewardDenom)
	suite.Equal("test1", blunaRewardConfigMonitor.State.HubContract)

	crc32first := blunaRewardConfigMonitor.metrics[BlunaRewardConfigCRC32]

	//   changing response data
	rewardConfigFakeResponse = struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.BlunaRewardConfig{
			HubContract: "test2",
			RewardDenom: "uluna",
		},
	}
	responseData, err = json.Marshal(rewardConfigFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	blunaRewardConfigMonitor = NewBlunaRewardConfigMonitor(cfg)

	err = blunaRewardConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal("uluna", blunaRewardConfigMonitor.State.RewardDenom)
	suite.Equal("test2", blunaRewardConfigMonitor.State.HubContract)

	crc32second := blunaRewardConfigMonitor.metrics[BlunaRewardConfigCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorTestSuite) TestHubConfig() {
	hubConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubConfig{
			Creator:                    "creator",
			RewardDispatcherContract:   "1",
			ValidatorsRegistryContract: "2",
			BlunaTokenContract:         "3",
			StlunaTokenContract:        "4",
			AirdropRegistryContract:    "5",
		},
	}
	responseData, err := json.Marshal(hubConfigFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	blunaRewardConfigMonitor := NewHubConfigMonitor(cfg)

	err = blunaRewardConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal("creator", blunaRewardConfigMonitor.State.Creator)
	suite.Equal("1", blunaRewardConfigMonitor.State.RewardDispatcherContract)
	suite.Equal("2", blunaRewardConfigMonitor.State.ValidatorsRegistryContract)
	suite.Equal("3", blunaRewardConfigMonitor.State.BlunaTokenContract)
	suite.Equal("4", blunaRewardConfigMonitor.State.StlunaTokenContract)
	suite.Equal("5", blunaRewardConfigMonitor.State.AirdropRegistryContract)

	crc32first := blunaRewardConfigMonitor.metrics[HubConfigCRC32]

	//   changing response data
	hubConfigFakeResponse = struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubConfig{
			Creator:                    "creator_new",
			RewardDispatcherContract:   "1_new",
			ValidatorsRegistryContract: "2_new",
			BlunaTokenContract:         "3_new",
			StlunaTokenContract:        "4_new",
			AirdropRegistryContract:    "5_new",
		},
	}
	responseData, err = json.Marshal(hubConfigFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	blunaRewardConfigMonitor = NewHubConfigMonitor(cfg)

	err = blunaRewardConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal("creator_new", blunaRewardConfigMonitor.State.Creator)
	suite.Equal("1_new", blunaRewardConfigMonitor.State.RewardDispatcherContract)
	suite.Equal("2_new", blunaRewardConfigMonitor.State.ValidatorsRegistryContract)
	suite.Equal("3_new", blunaRewardConfigMonitor.State.BlunaTokenContract)
	suite.Equal("4_new", blunaRewardConfigMonitor.State.StlunaTokenContract)
	suite.Equal("5_new", blunaRewardConfigMonitor.State.AirdropRegistryContract)

	crc32second := blunaRewardConfigMonitor.metrics[HubConfigCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorTestSuite) TestRewardDispatcherConfig() {
	rewardDispatcherConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.RewardDispatcherConfig{
			Owner:               "owner",
			HubContract:         "hub",
			BlunaRewardContract: "bluna",
			StlunaRewardDenom:   "stluna",
			BlunaRewardDenom:    "blunareward",
			LidoFeeAddress:      "lidofee",
			LidoFeeRate:         "0.01",
		},
	}
	responseData, err := json.Marshal(rewardDispatcherConfigFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	rewardDispatcherConfigMonitor := NewRewardDispatcherConfigMonitor(cfg)

	err = rewardDispatcherConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(rewardDispatcherConfigFakeResponse.Result, *rewardDispatcherConfigMonitor.State)
	crc32first := rewardDispatcherConfigMonitor.metrics[RewardDispatcherConfigCRC32]

	//   changing response data
	newRewardDispatcherConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.RewardDispatcherConfig{
			Owner:               "owner_new",
			HubContract:         "hub_new",
			BlunaRewardContract: "bluna_new",
			StlunaRewardDenom:   "stluna_new",
			BlunaRewardDenom:    "blunareward_new",
			LidoFeeAddress:      "lidofee_new",
			LidoFeeRate:         "0.005",
		},
	}
	responseData, err = json.Marshal(newRewardDispatcherConfigFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	rewardDispatcherConfigMonitor = NewRewardDispatcherConfigMonitor(cfg)

	err = rewardDispatcherConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(newRewardDispatcherConfigFakeResponse.Result, *rewardDispatcherConfigMonitor.State)

	crc32second := rewardDispatcherConfigMonitor.metrics[RewardDispatcherConfigCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorTestSuite) TestValidatorsRegistryConfig() {
	validatorsRegistryConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.ValidatorsRegistryConfig{
			Owner:       "owner",
			HubContract: "hub",
		},
	}
	responseData, err := json.Marshal(validatorsRegistryConfigFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	validatorsRegistryConfigMonitor := NewValidatorsRegistryConfigMonitor(cfg)

	err = validatorsRegistryConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(validatorsRegistryConfigFakeResponse.Result, *validatorsRegistryConfigMonitor.State)
	crc32first := validatorsRegistryConfigMonitor.metrics[ValidatorsRegistryConfigCRC32]

	//   changing response data
	newValidatorsRegistryConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.ValidatorsRegistryConfig{
			Owner:       "owner_new",
			HubContract: "hub_new",
		},
	}
	responseData, err = json.Marshal(newValidatorsRegistryConfigFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	validatorsRegistryConfigMonitor = NewValidatorsRegistryConfigMonitor(cfg)

	err = validatorsRegistryConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(newValidatorsRegistryConfigFakeResponse.Result, *validatorsRegistryConfigMonitor.State)

	crc32second := validatorsRegistryConfigMonitor.metrics[RewardDispatcherConfigCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorTestSuite) TestHubParameters() {
	hubParametersFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubParameters{
			EpochPeriod:         10,
			UnderlyingCoinDenom: "uluna",
			UnbondingPeriod:     20,
			PegRecoveryFee:      "0.5",
			ErThreshold:         "1",
			RewardDenom:         "uusd",
		},
	}
	responseData, err := json.Marshal(hubParametersFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	hubParametersMonitor := NewHubParametersMonitor(cfg)

	err = hubParametersMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(hubParametersFakeResponse.Result, *hubParametersMonitor.State)
	crc32first := hubParametersMonitor.metrics[HubParametersCRC32]

	//   changing response data
	newHubParametersFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubParameters{
			EpochPeriod:         100,
			UnderlyingCoinDenom: "uusd",
			UnbondingPeriod:     200,
			PegRecoveryFee:      "1.5",
			ErThreshold:         "2",
			RewardDenom:         "uluna",
		},
	}
	responseData, err = json.Marshal(newHubParametersFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	hubParametersMonitor = NewHubParametersMonitor(cfg)

	err = hubParametersMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(newHubParametersFakeResponse.Result, *hubParametersMonitor.State)

	crc32second := hubParametersMonitor.metrics[HubParametersCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorTestSuite) TestAirDropRegistryConfig() {
	airdropRegistryConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.AirDropRegistryConfig{
			Owner:       "owner",
			HubContract: "hub",
			AirDropToken: []string{"token1"},
		},
	}
	responseData, err := json.Marshal(airdropRegistryConfigFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)
	airdropRegistryConfigMonitor := NewAirDropRegistryConfigMonitor(cfg)

	err = airdropRegistryConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(airdropRegistryConfigFakeResponse.Result, *airdropRegistryConfigMonitor.State)
	crc32first := airdropRegistryConfigMonitor.metrics[AirDropRegistryConfigCRC32]

	//   changing response data
	newAirDropRegistryConfigFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.AirDropRegistryConfig{
			Owner:       "owner_new",
			HubContract: "hub_new",
			AirDropToken: []string{"token1","token2"},
		},
	}
	responseData, err = json.Marshal(newAirDropRegistryConfigFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	airdropRegistryConfigMonitor = NewAirDropRegistryConfigMonitor(cfg)

	err = airdropRegistryConfigMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(newAirDropRegistryConfigFakeResponse.Result, *airdropRegistryConfigMonitor.State)

	crc32second := airdropRegistryConfigMonitor.metrics[AirDropRegistryConfigCRC32]

	//detects crc32 is changed due to changed config data
	suite.NotEqual(crc32first, crc32second)
}
