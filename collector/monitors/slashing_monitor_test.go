package monitors

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
)

const (
	testValAddress   = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0spyk"
	testValPublicKey = "terravalconspub1zcjduepqw2hyr7u7y70z5kdewn00xuq0wwcvnn0s7x5pjqcdpn80qsyctcpqcjhz4c"
)

type SlashingMonitorTestSuite struct {
	suite.Suite
}

func (suite *SlashingMonitorTestSuite) SetupTest() {

}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestWithSlashing() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_blocks_jailed_tombstoned.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):                  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)


	valRepository := NewV1ValidatorsRepository(cfg)

	m := NewSlashingMonitor(cfg, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)


	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()
	var (
		expectedNumTombstonedValidators MetricValue = &SimpleMetricValue{1}
		expectedNumJailedValidators     MetricValue = &SimpleMetricValue{1}
		expectedNumMissedBlocks         float64     = 5
	)
	var actualMissedBlocks float64
	for _, missedBlocks := range metricVectors[SlashingNumMissedBlocks] {
		actualMissedBlocks += missedBlocks
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestNoSlashing() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_no_slashing.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):                  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)


	valRepository := NewV1ValidatorsRepository(cfg)

	m := NewSlashingMonitor(cfg, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)


	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()

	var (
		expectedNumTombstonedValidators MetricValue = &SimpleMetricValue{0}
		expectedNumJailedValidators     MetricValue = &SimpleMetricValue{0}
		expectedNumMissedBlocks         float64     = 0
	)
	var actualMissedBlocks float64
	for _, missedBlocks := range metricVectors[SlashingNumMissedBlocks] {
		actualMissedBlocks += missedBlocks
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestFailedSlashingRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_blocks_jailed_tombstoned.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):                  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)


	valRepository := NewV1ValidatorsRepository(cfg)

	m := NewSlashingMonitor(cfg, valRepository)
	err = m.Handler(context.Background())
	suite.Error(err)

	expectedErrorMessage := "failed to getValidatorsPublicKeys"
	suite.Contains(err.Error(),expectedErrorMessage)


}
