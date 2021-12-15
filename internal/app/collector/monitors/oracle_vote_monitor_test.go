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

type OracleVotesMonitorTestSuite struct {
	suite.Suite
}

func (suite *OracleVotesMonitorTestSuite) SetupTest() {

}

func (suite *OracleVotesMonitorTestSuite) testSuccessfulRequest(networkGenerations string) {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/slashing_validator_info_not_jailed.json", networkGenerations))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	oracleParams, err := ioutil.ReadFile(dir + "test_data/oracle_parameters.json")
	suite.NoError(err)

	oracleMissedVotePeriods, err := ioutil.ReadFile(dir + "test_data/oracle_missed_vote_periods.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract):  string(whitelistedValidators),
		"/oracle/parameters": string(oracleParams),
		fmt.Sprintf("/oracle/voters/%s/miss", types.TestValAddress): string(oracleMissedVotePeriods),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGenerations
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	m := NewOracleVotesMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedValidatorsCommission := 0.1
	actualValidatorsCommission := metricVectors[OracleMissedVoteRate].Get(types.TestMoniker)

	suite.Equal(expectedValidatorsCommission, actualValidatorsCommission)
}

func (suite *OracleVotesMonitorTestSuite) TestFailedValidatorsFeeRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	oracleParams, err := ioutil.ReadFile(dir + "test_data/slashing_error.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract):  string(whitelistedValidators),
		"/oracle/parameters": string(oracleParams),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)

	m := NewOracleVotesMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.Error(err)
}
