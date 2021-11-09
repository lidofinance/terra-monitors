package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/signinfo"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/stretchr/testify/suite"
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
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/slashing_validator_info_jailed.json", networkGeneration))
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf(dir+"test_data/%s/slashing_success_response_blocks_jailed_tombstoned.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", types.TestValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", types.TestConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract): string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := stubs.NewTestLogger()
	valRepository := validators.NewValidatorsRepository(cfg, logger)
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
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/slashing_validator_info_not_jailed.json", networkGeneration))
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf(dir+"test_data/%s/slashing_success_response_no_slashing.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", types.TestValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", types.TestConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract): string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := stubs.NewTestLogger()
	valRepository := validators.NewValidatorsRepository(cfg, logger)
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
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/slashing_error.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		fmt.Sprintf(dir+"test_data/%s/slashing_success_response_blocks_jailed_tombstoned.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	var signingInfoEndpoint string
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		signingInfoEndpoint = fmt.Sprintf("/slashing/validators/%s/signing_info", types.TestValPublicKey)
	case config.NetworkGenerationColumbus5:
		signingInfoEndpoint = fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", types.TestConsAddress)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		signingInfoEndpoint: string(validatorSigningInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract): string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := stubs.NewTestLogger()
	valRepository := validators.NewValidatorsRepository(cfg, logger)
	signInfoRepository := signinfo.NewSignInfoRepository(cfg, logger)

	m := NewSlashingMonitor(cfg, logger, valRepository, signInfoRepository)
	err = m.Handler(context.Background())
	suite.Error(err)

	expectedErrorMessage := "failed to getValidatorsInfo"
	suite.Contains(err.Error(), expectedErrorMessage)

}
