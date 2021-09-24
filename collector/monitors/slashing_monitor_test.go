package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/collector/monitors/signinfo"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/stretchr/testify/suite"
)

const (
	testValAddress   = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0spyk"
	testValPublicKey = "terravalconspub1zcjduepqw2hyr7u7y70z5kdewn00xuq0wwcvnn0s7x5pjqcdpn80qsyctcpqcjhz4c"
	testConsAddress  = "terravalcons1rfaxjug6md5jrz3c0uctyt6pzd50xyxlc2tf5m"
)

type SlashingMonitorTestSuite struct {
	suite.Suite
}

func (suite *SlashingMonitorTestSuite) SetupTest() {

}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestWithSlashing() {
	suite.testSuccessfulRequestWithSlashing(config.NetworkGenerationColumbus4)
	suite.testSuccessfulRequestWithSlashing(config.NetworkGenerationColumbus5)
}

func (suite *SlashingMonitorTestSuite) testSuccessfulRequestWithSlashing(networkGeneration string) {
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/slashing_validator_info_jailed.json", networkGeneration))
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf("./test_data/%s/slashing_success_response_blocks_jailed_tombstoned.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", testConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	signInfoRepository := signinfo.NewSignInfoRepository(cfg, logger)

	m := NewSlashingMonitor(cfg, logger, valRepository, signInfoRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()
	var (
		expectedNumTombstonedValidators MetricValue = &SimpleMetricValue{value: 1}
		expectedNumJailedValidators     MetricValue = &SimpleMetricValue{value: 1}
		expectedNumMissedBlocks         float64     = 5
	)
	var actualMissedBlocks float64
	for _, missedBlocks := range metricVectors[SlashingNumMissedBlocks].Labels() {
		actualMissedBlocks += metricVectors[SlashingNumMissedBlocks].Get(missedBlocks)
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestNoSlashing() {
	suite.testSuccessfulRequestNoSlashing(config.NetworkGenerationColumbus4)
	suite.testSuccessfulRequestNoSlashing(config.NetworkGenerationColumbus5)
}

func (suite *SlashingMonitorTestSuite) testSuccessfulRequestNoSlashing(networkGeneration string) {
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/slashing_validator_info_not_jailed.json", networkGeneration))
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf("./test_data/%s/slashing_success_response_no_slashing.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", testConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	signInfoRepository := signinfo.NewSignInfoRepository(cfg, logger)

	m := NewSlashingMonitor(cfg, logger, valRepository, signInfoRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()

	var (
		expectedNumTombstonedValidators MetricValue = &SimpleMetricValue{value: 0}
		expectedNumJailedValidators     MetricValue = &SimpleMetricValue{value: 0}
		expectedNumMissedBlocks         float64     = 0
	)
	var actualMissedBlocks float64
	for _, missedBlocks := range metricVectors[SlashingNumMissedBlocks].Labels() {
		actualMissedBlocks += metricVectors[SlashingNumMissedBlocks].Get(missedBlocks)
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestFailedSlashingRequest() {
	suite.testFailedSlashingRequest(config.NetworkGenerationColumbus4)
	suite.testFailedSlashingRequest(config.NetworkGenerationColumbus5)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) testFailedSlashingRequest(networkGeneration string) {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf("./test_data/%s/slashing_success_response_blocks_jailed_tombstoned.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", testConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	signInfoRepository := signinfo.NewSignInfoRepository(cfg, logger)

	m := NewSlashingMonitor(cfg, logger, valRepository, signInfoRepository)
	err = m.Handler(context.Background())
	suite.Error(err)

	expectedErrorMessage := "failed to getValidatorsInfo"
	suite.Contains(err.Error(), expectedErrorMessage)

}
