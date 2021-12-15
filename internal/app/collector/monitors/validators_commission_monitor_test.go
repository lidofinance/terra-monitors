package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/stretchr/testify/suite"
)

type ValidatorsCommissionTestSuite struct {
	suite.Suite
}

func (suite *ValidatorsCommissionTestSuite) SetupTest() {

}

func (suite *ValidatorsCommissionTestSuite) TestSuccessfulRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract):  string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	m := NewValidatorsFeeMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedValidatorsCommission := 0.08
	actualValidatorsCommission := metricVectors[ValidatorsCommission].Get(types.TestMoniker)

	suite.Equal(expectedValidatorsCommission, actualValidatorsCommission)
}

func (suite *ValidatorsCommissionTestSuite) TestFailedValidatorsFeeRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/slashing_error.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract):  string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	m := NewValidatorsFeeMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.Error(err)
}

func (suite *ValidatorsCommissionTestSuite) TestFailedV2ValidatorsRepository() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/slashing_error.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),

		// error format is the same
		fmt.Sprintf("/wasm/contracts/%s/store", types.ValidatorsRegistryContract): string(validatorInfoData),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V2Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.Nil(validators)
	suite.Error(err)

	_, err = valRepository.GetValidatorInfo(context.Background(), types.TestValAddress)
	suite.Error(err)
}

func (suite *ValidatorsCommissionTestSuite) TestFailedValidatorsRepository() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/slashing_error.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),

		// error format is the same
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract): string(validatorInfoData),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.Nil(validators)
	suite.Error(err)

	_, err = valRepository.GetValidatorInfo(context.Background(), types.TestValAddress)
	suite.Error(err)
}
